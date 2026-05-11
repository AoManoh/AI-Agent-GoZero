package svc

import (
	"context"
	"fmt"
	"strings"

	"GoZero-AI/internal/sessionmode"
	"GoZero-AI/internal/statuserr"

	"github.com/pgvector/pgvector-go"
	"github.com/sashabaranov/go-openai"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ResumeStore struct {
	db             sqlx.SqlConn
	openAIClient   *openai.Client
	embeddingModel string
}

type resumeSessionState struct {
	UserId   int64  `db:"user_id"`
	IsActive bool   `db:"is_active"`
	Mode     string `db:"mode"`
}

type sessionStateReader interface {
	QueryRowCtx(ctx context.Context, v any, query string, args ...any) error
}

func NewResumeStore(db sqlx.SqlConn, openAIClient *openai.Client, embeddingModel string) *ResumeStore {
	return &ResumeStore{
		db:             db,
		openAIClient:   openAIClient,
		embeddingModel: embeddingModel,
	}
}

func (s *ResumeStore) SaveResume(ctx context.Context, userID int64, artifactID, sessionID, title, filename, mode string, chunks []string) (int64, error) {
	if len(chunks) == 0 {
		return 0, fmt.Errorf("简历内容为空")
	}
	artifactID = strings.TrimSpace(artifactID)
	sessionID = strings.TrimSpace(sessionID)
	if artifactID == "" {
		return 0, fmt.Errorf("简历资产ID不能为空")
	}
	if sessionID != "" {
		if _, _, err := ensureResumeSessionWritable(ctx, s.db, userID, sessionID); err != nil {
			return 0, err
		}
	}
	if strings.TrimSpace(title) == "" {
		title = filename
	} else {
		title = strings.TrimSpace(title)
	}
	if strings.TrimSpace(filename) == "" {
		return 0, fmt.Errorf("简历文件名不能为空")
	}
	filename = strings.TrimSpace(filename)

	if mode = strings.TrimSpace(mode); mode == "" {
		mode = sessionmode.KeyMemory
	}

	embeddings := make([]pgvector.Vector, 0, len(chunks))
	for _, chunk := range chunks {
		embedding, err := s.generateEmbedding(ctx, chunk)
		if err != nil {
			return 0, err
		}
		embeddings = append(embeddings, embedding)
	}

	var version int64
	err := s.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		if sessionID != "" {
			existing, found, err := ensureResumeSessionWritable(ctx, session, userID, sessionID)
			if err != nil {
				return err
			}
			modeKey := resolveResumeSessionMode(existing.Mode, mode)
			if !found {
				modeKey = resolveResumeSessionMode("", mode)
			}

			if _, err := session.ExecCtx(ctx, `INSERT INTO "public"."chat_sessions" (session_id, user_id, title, mode, is_active)
	VALUES ($1, $2, $3, $4, true)
	ON CONFLICT (session_id) DO UPDATE
	SET user_id = COALESCE("public"."chat_sessions".user_id, EXCLUDED.user_id),
	    title = CASE
	        WHEN "public"."chat_sessions".title = $5 AND EXCLUDED.title <> '' THEN EXCLUDED.title
	        ELSE "public"."chat_sessions".title
	    END,
	    mode = CASE
	        WHEN "public"."chat_sessions".mode IS NULL OR btrim("public"."chat_sessions".mode) = '' THEN EXCLUDED.mode
	        ELSE "public"."chat_sessions".mode
	    END,
	    updated_at = now()`,
				sessionID, userID, title, modeKey, defaultSessionTitle); err != nil {
				return err
			}
			if _, _, err := ensureResumeSessionWritable(ctx, session, userID, sessionID); err != nil {
				return err
			}
		}

		if _, err := session.ExecCtx(ctx, `DELETE FROM "public"."vector_store" WHERE user_id = $1 AND chat_id = $2 AND doc_type = 'resume'`, userID, artifactID); err != nil {
			return err
		}

		if err := session.QueryRowCtx(ctx, &version, `select coalesce(max(version), 0) + 1
from "public"."resume_documents"
where user_id = $1 and artifact_id = $2`, userID, artifactID); err != nil {
			return err
		}
		if _, err := session.ExecCtx(ctx, `update "public"."resume_documents"
set is_current = false, updated_at = now()
where user_id = $1 and artifact_id = $2 and is_current = true`, userID, artifactID); err != nil {
			return err
		}
		if _, err := session.ExecCtx(ctx, `insert into "public"."resume_documents"
(artifact_id, user_id, session_id, version, title, filename, status, parse_stage, parse_progress, processed_chunk_count, failed_chunk_count, parse_error_code, parse_error_message, parse_retryable, chunk_count, is_current, uploaded_at, updated_at)
values ($1, $2, $3, $4, $5, $6, 'ready', 'ready', 100, $7, 0, '', '', false, $7, true, now(), now())`,
			artifactID, userID, sessionID, version, title, filename, len(chunks)); err != nil {
			return err
		}

		for idx, chunk := range chunks {
			if _, err := session.ExecCtx(ctx, `INSERT INTO "public"."vector_store" (chat_id, user_id, role, content, embedding, doc_type) VALUES ($1, $2, $3, $4, $5, 'resume')`,
				artifactID, userID, resumeRole, chunk, embeddings[idx]); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}
	return version, nil
}

func (s *ResumeStore) generateEmbedding(ctx context.Context, content string) (pgvector.Vector, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return pgvector.NewVector(nil), fmt.Errorf("简历分片不能为空")
	}

	embedding, err := s.openAIClient.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: openai.EmbeddingModel(s.embeddingModel),
		Input: []string{trimmed},
	})
	if err != nil {
		return pgvector.NewVector(nil), fmt.Errorf("生成简历向量失败: %w", err)
	}
	if len(embedding.Data) == 0 {
		return pgvector.NewVector(nil), fmt.Errorf("生成简历向量失败: 空响应")
	}

	return pgvector.NewVector(embedding.Data[0].Embedding), nil
}

func ensureResumeSessionWritable(ctx context.Context, reader sessionStateReader, userID int64, chatID string) (resumeSessionState, bool, error) {
	var state resumeSessionState
	err := reader.QueryRowCtx(ctx, &state, `SELECT user_id, is_active, mode FROM "public"."chat_sessions" WHERE session_id = $1 LIMIT 1`, chatID)
	switch err {
	case nil:
		if state.UserId != userID {
			return resumeSessionState{}, false, statuserr.Forbidden("无权访问该会话")
		}
		if !state.IsActive {
			return resumeSessionState{}, false, statuserr.Conflict("会话已删除，请创建新会话")
		}
		return state, true, nil
	case sqlx.ErrNotFound:
		return resumeSessionState{}, false, nil
	default:
		return resumeSessionState{}, false, err
	}
}

func resolveResumeSessionMode(storedMode, requestedMode string) string {
	if strings.TrimSpace(storedMode) != "" {
		return sessionmode.NormalizeKey(storedMode)
	}
	return sessionmode.NormalizeKey(requestedMode)
}

const (
	resumeRole          = "resume"
	defaultSessionTitle = "新对话"
)
