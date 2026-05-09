-- 粘贴我们原来的建表文件，这里贴一下吧
-- v2版本：启用 pgvector 扩展
CREATE EXTENSION IF NOT EXISTS vector;

-- 启用 cube 和 earthdistance 扩展（基础依赖）
CREATE EXTENSION IF NOT EXISTS cube;
CREATE EXTENSION IF NOT EXISTS earthdistance;

-- 安装 pg_trgm 扩展（支持 jsonb 相似度操作）
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ----------------------------
-- 删除已存在的vector_store和knowledge_base表（如果存在）
-- ----------------------------
DROP TABLE IF EXISTS "public"."session_evaluations";
DROP TABLE IF EXISTS "public"."chat_sessions";
DROP TABLE IF EXISTS "public"."vector_store";
DROP TABLE IF EXISTS "public"."knowledge_base";
DROP TABLE IF EXISTS "public"."users";

CREATE TABLE "public"."users" (
                                   "id" BIGSERIAL PRIMARY KEY,
                                   "username" VARCHAR(64) UNIQUE NOT NULL,
                                   "password_hash" VARCHAR(255) NOT NULL,
                                   "created_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO "public"."users" ("id", "username", "password_hash")
VALUES (1, 'your_username', '$2a$10$b1r0Ng24On7XGaHKvOuzmOzr3do5f4Y7wmqvUidhDrO3Ujpw3XwYq')
ON CONFLICT ("id") DO NOTHING;

SELECT setval(pg_get_serial_sequence('public.users', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM "public"."users"), 1), true);

-- 创建新表
CREATE TABLE "public"."vector_store" (
                                         "id" BIGSERIAL PRIMARY KEY,
                                         "chat_id" varchar(255) NOT NULL,
                                         "user_id" BIGINT REFERENCES "public"."users"("id") ON DELETE CASCADE,
                                         "role" varchar(50) NOT NULL,   -- 新增角色字段
                                         "content" TEXT NOT NULL,
                                         "embedding" vector(1536) NOT NULL,
                                         "doc_type" VARCHAR(50) NOT NULL DEFAULT 'message',
                                         "created_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 创建索引
CREATE INDEX idx_vector_store_chat_id ON vector_store (chat_id);
CREATE INDEX idx_vector_store_user_id_type ON vector_store (user_id, doc_type);
CREATE INDEX idx_vector_store_chat_user_type ON vector_store (chat_id, user_id, doc_type);
CREATE INDEX idx_vector_store_created_at ON vector_store (created_at DESC);

-- 创建知识库内容表
CREATE TABLE "public"."knowledge_base" (
                                           "id" BIGSERIAL PRIMARY KEY,
                                           "user_id" BIGINT NOT NULL DEFAULT 1 REFERENCES "public"."users"("id") ON DELETE CASCADE,
                                           "title" VARCHAR(255) NOT NULL,
                                           "content" TEXT NOT NULL,
                                           "embedding" vector(1536) NOT NULL,
                                           "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 创建索引
CREATE INDEX idx_knowledge_base_title ON knowledge_base (title);
CREATE INDEX idx_knowledge_base_user_id ON knowledge_base (user_id);

CREATE TABLE "public"."chat_sessions" (
                                          "id" BIGSERIAL PRIMARY KEY,
                                          "session_id" VARCHAR(64) UNIQUE NOT NULL,
                                          "user_id" BIGINT NOT NULL REFERENCES "public"."users"("id") ON DELETE CASCADE,
                                          "title" VARCHAR(200) NOT NULL DEFAULT '新对话',
                                          "mode" VARCHAR(64) NOT NULL DEFAULT 'Interview',
                                          "direction_key" VARCHAR(64) NOT NULL DEFAULT 'go_backend',
                                          "direction_label" VARCHAR(80) NOT NULL DEFAULT 'Go 后端',
                                          "difficulty_level" INTEGER NOT NULL DEFAULT 3,
                                          "difficulty_label" VARCHAR(32) NOT NULL DEFAULT '中级',
                                          "interviewer_style" VARCHAR(64) NOT NULL DEFAULT 'senior',
                                          "interviewer_style_label" VARCHAR(80) NOT NULL DEFAULT '资深技术官',
                                          "focus_areas" JSONB NOT NULL DEFAULT '[]'::jsonb,
                                          "follow_up_depth" VARCHAR(16) NOT NULL DEFAULT 'N+3',
                                          "estimated_minutes" INTEGER NOT NULL DEFAULT 30,
                                          "progress_percent" INTEGER NOT NULL DEFAULT 0,
                                          "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                          "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                          "last_message_at" TIMESTAMPTZ,
                                          "started_at" TIMESTAMPTZ,
                                          "completed_at" TIMESTAMPTZ,
                                          "duration_seconds" INTEGER NOT NULL DEFAULT 0,
                                          "message_count" INTEGER NOT NULL DEFAULT 0,
                                          "is_active" BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX idx_chat_sessions_user_id_last_message_at ON chat_sessions (user_id, last_message_at DESC);
CREATE INDEX idx_chat_sessions_user_id_is_active ON chat_sessions (user_id, is_active);
CREATE INDEX idx_chat_sessions_user_completed_at ON chat_sessions (user_id, completed_at DESC);

CREATE TABLE "public"."resume_documents" (
                                             "id" BIGSERIAL PRIMARY KEY,
                                             "user_id" BIGINT NOT NULL REFERENCES "public"."users"("id") ON DELETE CASCADE,
                                             "session_id" VARCHAR(64) NOT NULL REFERENCES "public"."chat_sessions"("session_id") ON DELETE CASCADE,
                                             "version" BIGINT NOT NULL DEFAULT 1,
                                             "title" VARCHAR(200) NOT NULL,
                                             "filename" VARCHAR(255) NOT NULL,
                                             "status" VARCHAR(32) NOT NULL DEFAULT 'ready',
                                             "chunk_count" INTEGER NOT NULL DEFAULT 0,
                                             "is_current" BOOLEAN NOT NULL DEFAULT true,
                                             "uploaded_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                             "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                             UNIQUE ("user_id", "session_id", "version")
);

CREATE INDEX idx_resume_documents_user_current ON resume_documents (user_id, is_current, updated_at DESC);
CREATE INDEX idx_resume_documents_session_current ON resume_documents (session_id, user_id, is_current);

CREATE TABLE "public"."session_evaluations" (
                                                "id" BIGSERIAL PRIMARY KEY,
                                                "session_id" VARCHAR(64) NOT NULL REFERENCES "public"."chat_sessions"("session_id") ON DELETE CASCADE,
                                                "user_id" BIGINT NOT NULL REFERENCES "public"."users"("id") ON DELETE CASCADE,
                                                "status" VARCHAR(32) NOT NULL DEFAULT 'draft',
                                                "summary" TEXT NOT NULL DEFAULT '',
                                                "user_turns" BIGINT NOT NULL DEFAULT 0,
                                                "assistant_turns" BIGINT NOT NULL DEFAULT 0,
                                                "overall_score" DOUBLE PRECISION NOT NULL DEFAULT 0,
                                                "rubric_version" VARCHAR(32) NOT NULL DEFAULT 'rubric-v1',
                                                "score_source" VARCHAR(32) NOT NULL DEFAULT 'heuristic',
                                                "dimensions" JSONB NOT NULL DEFAULT '[]'::jsonb,
                                                "strengths" JSONB NOT NULL DEFAULT '[]'::jsonb,
                                                "risks" JSONB NOT NULL DEFAULT '[]'::jsonb,
                                                "suggestions" JSONB NOT NULL DEFAULT '[]'::jsonb,
                                                "evidence" JSONB NOT NULL DEFAULT '[]'::jsonb,
                                                "source_last_message_id" BIGINT,
                                                "source_last_message_at" TIMESTAMPTZ,
                                                "first_generated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                                "generated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                                "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                                UNIQUE ("session_id", "user_id")
);

CREATE INDEX idx_session_evaluations_user_id ON session_evaluations (user_id);
CREATE INDEX idx_session_evaluations_session_id ON session_evaluations (session_id);

CREATE TABLE "public"."session_evaluation_items" (
                                                     "id" BIGSERIAL PRIMARY KEY,
                                                     "session_id" VARCHAR(64) NOT NULL REFERENCES "public"."chat_sessions"("session_id") ON DELETE CASCADE,
                                                     "user_id" BIGINT NOT NULL REFERENCES "public"."users"("id") ON DELETE CASCADE,
                                                     "turn_index" INTEGER NOT NULL,
                                                     "question" TEXT NOT NULL DEFAULT '',
                                                     "answer" TEXT NOT NULL DEFAULT '',
                                                     "ai_comment" TEXT NOT NULL DEFAULT '',
                                                     "score" DOUBLE PRECISION NOT NULL DEFAULT 0,
                                                     "max_score" DOUBLE PRECISION NOT NULL DEFAULT 5,
                                                     "tags" JSONB NOT NULL DEFAULT '[]'::jsonb,
                                                     "source_message_id" BIGINT,
                                                     "source_message_at" TIMESTAMPTZ,
                                                     "generated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                                     "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                                     UNIQUE ("session_id", "user_id", "turn_index")
);

CREATE INDEX idx_session_evaluation_items_user_session ON session_evaluation_items (user_id, session_id, turn_index);
