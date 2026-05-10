package model

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pgvector/pgvector-go"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	InterviewQuestionsModel interface {
		List(ctx context.Context, opts InterviewQuestionListOptions) ([]InterviewQuestion, int64, error)
		FindOne(ctx context.Context, keyOrID string) (*InterviewQuestion, []InterviewQuestionSource, error)
		Stats(ctx context.Context) (*InterviewQuestionStats, error)
		DirectionCounts(ctx context.Context) (map[string]int64, error)
		AttachToSession(ctx context.Context, session sqlx.Session, userID int64, sessionID string, question InterviewQuestion) error
	}

	defaultInterviewQuestionsModel struct {
		conn  sqlx.SqlConn
		table string
	}

	InterviewQuestionListOptions struct {
		DirectionKey string
		Difficulty   int64
		FocusKeys    []string
		Keyword      string
		Limit        int64
		Offset       int64
		Sort         string
	}

	InterviewQuestion struct {
		Id                   int64        `db:"id"`
		QuestionKey          string       `db:"question_key"`
		DirectionKey         string       `db:"direction_key"`
		FocusKey             string       `db:"focus_key"`
		FocusLabel           string       `db:"focus_label"`
		DifficultyLevel      int64        `db:"difficulty_level"`
		DifficultyLabel      string       `db:"difficulty_label"`
		Title                string       `db:"title"`
		Prompt               string       `db:"prompt"`
		ExpectedSignals      []byte       `db:"expected_signals"`
		FollowUps            []byte       `db:"follow_ups"`
		EvaluationDimensions []byte       `db:"evaluation_dimensions"`
		Tags                 []byte       `db:"tags"`
		SourceRefs           []byte       `db:"source_refs"`
		BatchKey             string       `db:"batch_key"`
		BatchLabel           string       `db:"batch_label"`
		Sequence             int64        `db:"sequence"`
		BatchSequence        int64        `db:"batch_sequence"`
		Status               string       `db:"status"`
		QualityScore         float64      `db:"quality_score"`
		UsageCount           int64        `db:"usage_count"`
		LastUsedAt           sql.NullTime `db:"last_used_at"`
		ContentHash          string       `db:"content_hash"`
		CreatedAt            time.Time    `db:"created_at"`
		UpdatedAt            time.Time    `db:"updated_at"`
		SourceCount          int64        `db:"source_count"`
	}

	InterviewQuestionSource struct {
		Id          int64     `db:"id"`
		QuestionId  int64     `db:"question_id"`
		SourceKey   string    `db:"source_key"`
		SourceTitle string    `db:"source_title"`
		SourceUrl   string    `db:"source_url"`
		SourceType  string    `db:"source_type"`
		LicenseNote string    `db:"license_note"`
		BatchKey    string    `db:"batch_key"`
		CreatedAt   time.Time `db:"created_at"`
	}

	InterviewQuestionStats struct {
		Total        int64        `db:"total"`
		UpdatedAt    sql.NullTime `db:"updated_at"`
		Directions   []InterviewQuestionDirectionStat
		Difficulties []InterviewQuestionDifficultyStat
	}

	InterviewQuestionDirectionStat struct {
		DirectionKey string `db:"direction_key"`
		Count        int64  `db:"count"`
	}

	InterviewQuestionDifficultyStat struct {
		DifficultyLevel int64 `db:"difficulty_level"`
		Count           int64 `db:"count"`
	}
)

const (
	interviewQuestionSelectFields = `q.id, q.question_key, q.direction_key, q.focus_key, q.focus_label,
q.difficulty_level, q.difficulty_label, q.title, q.prompt, q.expected_signals,
q.follow_ups, q.evaluation_dimensions, q.tags, q.source_refs, q.batch_key,
q.batch_label, q.sequence, q.batch_sequence, q.status, q.quality_score,
q.usage_count, q.last_used_at, q.content_hash, q.created_at, q.updated_at,
(select count(*) from "public"."interview_question_sources" s where s.question_id = q.id) as source_count`
	defaultInterviewQuestionLimit = int64(50)
	maxInterviewQuestionLimit     = int64(2000)
	embeddingDimension            = 1536
)

func NewInterviewQuestionsModel(conn sqlx.SqlConn) InterviewQuestionsModel {
	return &defaultInterviewQuestionsModel{
		conn:  conn,
		table: `"public"."interview_questions"`,
	}
}

func (m *defaultInterviewQuestionsModel) List(ctx context.Context, opts InterviewQuestionListOptions) ([]InterviewQuestion, int64, error) {
	where, args := buildInterviewQuestionWhere(opts)
	countQuery := fmt.Sprintf(`select count(*) from %s q %s`, m.table, where)

	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []InterviewQuestion{}, 0, nil
	}

	limit := normalizeInterviewQuestionLimit(opts.Limit)
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	listArgs := append([]any{}, args...)
	limitPlaceholder := nextPlaceholder(&listArgs, limit)
	offsetPlaceholder := nextPlaceholder(&listArgs, offset)
	query := fmt.Sprintf(`select %s
from %s q
%s
order by %s
limit %s offset %s`, interviewQuestionSelectFields, m.table, where, interviewQuestionOrderBy(opts.Sort), limitPlaceholder, offsetPlaceholder)

	var rows []InterviewQuestion
	err := m.conn.QueryRowsCtx(ctx, &rows, query, listArgs...)
	switch err {
	case nil:
		return rows, total, nil
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return []InterviewQuestion{}, total, nil
	default:
		return nil, 0, err
	}
}

func (m *defaultInterviewQuestionsModel) FindOne(ctx context.Context, keyOrID string) (*InterviewQuestion, []InterviewQuestionSource, error) {
	trimmed := strings.TrimSpace(keyOrID)
	if trimmed == "" {
		return nil, nil, ErrNotFound
	}

	var query string
	var args []any
	if id, err := strconv.ParseInt(trimmed, 10, 64); err == nil && id > 0 {
		query = fmt.Sprintf(`select %s from %s q where q.id = $1 and q.status = 'ready' limit 1`, interviewQuestionSelectFields, m.table)
		args = []any{id}
	} else {
		query = fmt.Sprintf(`select %s from %s q where q.question_key = $1 and q.status = 'ready' limit 1`, interviewQuestionSelectFields, m.table)
		args = []any{trimmed}
	}

	var question InterviewQuestion
	err := m.conn.QueryRowCtx(ctx, &question, query, args...)
	switch err {
	case nil:
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return nil, nil, ErrNotFound
	default:
		return nil, nil, err
	}

	sources, err := m.findSources(ctx, question.Id)
	if err != nil {
		return nil, nil, err
	}
	return &question, sources, nil
}

func (m *defaultInterviewQuestionsModel) Stats(ctx context.Context) (*InterviewQuestionStats, error) {
	var aggregate struct {
		Total     int64        `db:"total"`
		UpdatedAt sql.NullTime `db:"updated_at"`
	}
	if err := m.conn.QueryRowCtx(ctx, &aggregate, `select count(*) as total, max(updated_at) as updated_at
from "public"."interview_questions"
where status = 'ready'`); err != nil {
		return nil, err
	}
	stats := InterviewQuestionStats{
		Total:     aggregate.Total,
		UpdatedAt: aggregate.UpdatedAt,
	}

	if err := m.conn.QueryRowsCtx(ctx, &stats.Directions, `select direction_key, count(*) as count
from "public"."interview_questions"
where status = 'ready'
group by direction_key
order by direction_key`); err != nil && err != sqlx.ErrNotFound && err != sql.ErrNoRows {
		return nil, err
	}

	if err := m.conn.QueryRowsCtx(ctx, &stats.Difficulties, `select difficulty_level, count(*) as count
from "public"."interview_questions"
where status = 'ready'
group by difficulty_level
order by difficulty_level`); err != nil && err != sqlx.ErrNotFound && err != sql.ErrNoRows {
		return nil, err
	}

	return &stats, nil
}

func (m *defaultInterviewQuestionsModel) DirectionCounts(ctx context.Context) (map[string]int64, error) {
	var rows []InterviewQuestionDirectionStat
	err := m.conn.QueryRowsCtx(ctx, &rows, `select direction_key, count(*) as count
from "public"."interview_questions"
where status = 'ready'
group by direction_key`)
	if err != nil {
		if err == sqlx.ErrNotFound || err == sql.ErrNoRows {
			return map[string]int64{}, nil
		}
		return nil, err
	}
	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.DirectionKey] = row.Count
	}
	return counts, nil
}

func (m *defaultInterviewQuestionsModel) AttachToSession(ctx context.Context, session sqlx.Session, userID int64, sessionID string, question InterviewQuestion) error {
	zeroEmbedding := pgvector.NewVector(make([]float32, embeddingDimension))
	if _, err := session.ExecCtx(ctx, `insert into "public"."vector_store"
(chat_id, user_id, role, content, embedding, doc_type, created_at)
values ($1, $2, 'assistant', $3, $4, 'message', now())`,
		sessionID, userID, question.Prompt, zeroEmbedding); err != nil {
		return err
	}

	if _, err := session.ExecCtx(ctx, `insert into "public"."session_question_events"
(session_id, user_id, question_id, question_key, turn_index, source, question_snapshot, created_at)
values ($1, $2, $3, $4, 1, 'bank', $5, now())
on conflict (session_id, user_id, turn_index) do update set
    question_id = excluded.question_id,
    question_key = excluded.question_key,
    source = excluded.source,
    question_snapshot = excluded.question_snapshot`,
		sessionID, userID, question.Id, question.QuestionKey, question.Prompt); err != nil {
		return err
	}

	if _, err := session.ExecCtx(ctx, `update "public"."chat_sessions"
set message_count = (
        select count(*)
        from "public"."vector_store"
        where chat_id = $1 and user_id = $2 and doc_type = 'message'
    ),
    last_message_at = now(),
    updated_at = now()
where session_id = $1 and user_id = $2`, sessionID, userID); err != nil {
		return err
	}

	_, err := session.ExecCtx(ctx, `update "public"."interview_questions"
set usage_count = usage_count + 1,
    last_used_at = now(),
    updated_at = now()
where id = $1`, question.Id)
	return err
}

func (m *defaultInterviewQuestionsModel) findSources(ctx context.Context, questionID int64) ([]InterviewQuestionSource, error) {
	var sources []InterviewQuestionSource
	err := m.conn.QueryRowsCtx(ctx, &sources, `select id, question_id, source_key, source_title, source_url,
source_type, license_note, batch_key, created_at
from "public"."interview_question_sources"
where question_id = $1
order by id asc`, questionID)
	switch err {
	case nil:
		return sources, nil
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return []InterviewQuestionSource{}, nil
	default:
		return nil, err
	}
}

func buildInterviewQuestionWhere(opts InterviewQuestionListOptions) (string, []any) {
	clauses := []string{"q.status = 'ready'"}
	args := make([]any, 0)

	if direction := strings.TrimSpace(opts.DirectionKey); direction != "" {
		clauses = append(clauses, "q.direction_key = "+nextPlaceholder(&args, direction))
	}
	if opts.Difficulty > 0 {
		clauses = append(clauses, "q.difficulty_level = "+nextPlaceholder(&args, opts.Difficulty))
	}
	focusKeys := uniqueNonEmpty(opts.FocusKeys)
	if len(focusKeys) > 0 {
		placeholders := make([]string, 0, len(focusKeys))
		for _, key := range focusKeys {
			placeholders = append(placeholders, nextPlaceholder(&args, key))
		}
		clauses = append(clauses, "q.focus_key in ("+strings.Join(placeholders, ", ")+")")
	}
	if keyword := strings.TrimSpace(opts.Keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		placeholder := nextPlaceholder(&args, pattern)
		clauses = append(clauses, `(q.title ilike `+placeholder+` or q.prompt ilike `+placeholder+` or exists (
            select 1 from jsonb_array_elements_text(q.tags) tag where tag ilike `+placeholder+`
        ))`)
	}

	return "where " + strings.Join(clauses, " and "), args
}

func nextPlaceholder(args *[]any, value any) string {
	*args = append(*args, value)
	return fmt.Sprintf("$%d", len(*args))
}

func normalizeInterviewQuestionLimit(limit int64) int64 {
	if limit <= 0 {
		return defaultInterviewQuestionLimit
	}
	if limit > maxInterviewQuestionLimit {
		return maxInterviewQuestionLimit
	}
	return limit
}

func interviewQuestionOrderBy(sort string) string {
	switch strings.TrimSpace(sort) {
	case "new":
		return "q.updated_at desc, q.id desc"
	case "diff":
		return "q.difficulty_level asc, q.sequence asc, q.id asc"
	default:
		return "q.usage_count desc, q.quality_score desc, q.sequence asc, q.id asc"
	}
}

func uniqueNonEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}
