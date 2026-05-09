-- =================================================================
-- 数据库安全升级脚本 v5 (非破坏性, 兼容旧 embedding 存储形态)
-- 功能：在保留现有数据的基础上，升级表结构并为所有字段添加注释。
-- =================================================================

BEGIN; -- 使用事务，确保所有操作要么全部成功，要么全部失败

-- ----------------------------
-- 步骤 1: 创建 `users` 表并插入管理员 (如果不存在)
-- ----------------------------
CREATE TABLE IF NOT EXISTS "public"."users" (
                                                "id" BIGSERIAL PRIMARY KEY,
                                                "username" VARCHAR(64) UNIQUE NOT NULL,
                                                "password_hash" VARCHAR(255) NOT NULL,
                                                "created_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 为 users 表的所有字段添加注释
COMMENT ON TABLE "public"."users" IS '存储用户信息，用于身份认证';
COMMENT ON COLUMN "public"."users"."id" IS '用户的唯一标识符 (主键)';
COMMENT ON COLUMN "public"."users"."username" IS '用户的登录名 (唯一)';
COMMENT ON COLUMN "public"."users"."password_hash" IS '存储经过哈希和加盐处理后的用户密码';
COMMENT ON COLUMN "public"."users"."created_at" IS '用户账户的创建时间戳';

-- 插入管理员用户 (id=1)，如果尚不存在
DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM "public"."users" WHERE id = 1) THEN
            PERFORM setval(pg_get_serial_sequence('public.users', 'id'), 1, false);
            INSERT INTO "public"."users" (id, username, password_hash) VALUES (1, 'your_username', '$2a$10$b1r0Ng24On7XGaHKvOuzmOzr3do5f4Y7wmqvUidhDrO3Ujpw3XwYq'); -- 占位密码 your_password 的 bcrypt 哈希，部署前请替换
        END IF;
    END $$;


-- ----------------------------
-- 步骤 2: 升级 `knowledge_base` 表
-- ----------------------------
-- 2.1: 新增 user_id 字段，并假设所有现有知识都属于管理员(id=1)
ALTER TABLE "public"."knowledge_base" ADD COLUMN IF NOT EXISTS "user_id" BIGINT;
UPDATE "public"."knowledge_base" SET "user_id" = 1 WHERE "user_id" IS NULL;
ALTER TABLE "public"."knowledge_base" ALTER COLUMN "user_id" SET DEFAULT 1;
ALTER TABLE "public"."knowledge_base" ALTER COLUMN "user_id" SET NOT NULL;

-- 2.2: 将 embedding 字段从 JSONB 转换为 vector
DO $$
    DECLARE
        embedding_udt text;
        array_rows bigint;
        string_rows bigint;
        other_rows bigint;
        existing_dims integer;
        distinct_dims integer;
    BEGIN
        SELECT udt_name
        INTO embedding_udt
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'knowledge_base'
          AND column_name = 'embedding';

        IF embedding_udt = 'jsonb' THEN
            SELECT COUNT(*) FILTER (WHERE jsonb_typeof(embedding) = 'array'),
                   COUNT(*) FILTER (WHERE jsonb_typeof(embedding) = 'string'),
                   COUNT(*) FILTER (WHERE jsonb_typeof(embedding) NOT IN ('array', 'string'))
            INTO array_rows, string_rows, other_rows
            FROM "public"."knowledge_base"
            WHERE embedding IS NOT NULL;

            IF array_rows = 0 AND string_rows = 0 THEN
                ALTER TABLE "public"."knowledge_base"
                    ALTER COLUMN "embedding" TYPE vector(1536) USING ("embedding"::text::vector);
            ELSIF array_rows > 0 AND string_rows = 0 AND other_rows = 0 THEN
                SELECT COUNT(DISTINCT jsonb_array_length(embedding)),
                       COALESCE(MAX(jsonb_array_length(embedding)), 1536)
                INTO distinct_dims, existing_dims
                FROM "public"."knowledge_base"
                WHERE embedding IS NOT NULL;

                IF distinct_dims > 1 THEN
                    RAISE NOTICE 'skip converting knowledge_base.embedding: inconsistent vector dimensions in existing jsonb array data';
                ELSIF existing_dims = 1536 THEN
                    ALTER TABLE "public"."knowledge_base"
                        ALTER COLUMN "embedding" TYPE vector(1536) USING ("embedding"::text::vector);
                ELSE
                    RAISE NOTICE 'skip converting knowledge_base.embedding: existing dimension % does not match expected 1536', existing_dims;
                END IF;
            ELSIF string_rows > 0 AND array_rows = 0 AND other_rows = 0 THEN
                BEGIN
                    SELECT COUNT(DISTINCT jsonb_array_length((convert_from(decode(trim(both '"' from embedding::text), 'base64'), 'UTF8'))::jsonb)),
                           COALESCE(MAX(jsonb_array_length((convert_from(decode(trim(both '"' from embedding::text), 'base64'), 'UTF8'))::jsonb)), 1536)
                    INTO distinct_dims, existing_dims
                    FROM "public"."knowledge_base"
                    WHERE embedding IS NOT NULL;

                    IF distinct_dims > 1 THEN
                        RAISE NOTICE 'skip converting knowledge_base.embedding: inconsistent vector dimensions in base64 jsonb string data';
                    ELSIF existing_dims = 1536 THEN
                        ALTER TABLE "public"."knowledge_base"
                            ALTER COLUMN "embedding" TYPE vector(1536)
                            USING ((convert_from(decode(trim(both '"' from embedding::text), 'base64'), 'UTF8'))::vector);
                    ELSE
                        RAISE NOTICE 'skip converting knowledge_base.embedding: existing base64 dimension % does not match expected 1536', existing_dims;
                    END IF;
                EXCEPTION
                    WHEN OTHERS THEN
                        RAISE NOTICE 'skip converting knowledge_base.embedding: unsupported jsonb string payload: %', SQLERRM;
                END;
            ELSE
                RAISE NOTICE 'skip converting knowledge_base.embedding: mixed jsonb payload shapes detected';
            END IF;
        END IF;
    END $$;


-- 2.3: 添加外键约束 (如果不存在)
DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_kb_user_id') THEN
            ALTER TABLE "public"."knowledge_base" ADD CONSTRAINT fk_kb_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        END IF;
    END $$;

-- 2.4: 更新索引 (使用 DROP IF EXISTS + CREATE 模式)
DROP INDEX IF EXISTS idx_knowledge_base_title;
DROP INDEX IF EXISTS idx_kb_user_id;
ALTER TABLE "public"."knowledge_base" ADD COLUMN IF NOT EXISTS "source" TEXT NOT NULL DEFAULT '';
ALTER TABLE "public"."knowledge_base" ADD COLUMN IF NOT EXISTS "visibility" VARCHAR(32) NOT NULL DEFAULT 'public';
ALTER TABLE "public"."knowledge_base" ADD COLUMN IF NOT EXISTS "status" VARCHAR(32) NOT NULL DEFAULT 'ready';
ALTER TABLE "public"."knowledge_base" ADD COLUMN IF NOT EXISTS "version" BIGINT NOT NULL DEFAULT 1;
ALTER TABLE "public"."knowledge_base" ADD COLUMN IF NOT EXISTS "content_hash" VARCHAR(64);
ALTER TABLE "public"."knowledge_base" ADD COLUMN IF NOT EXISTS "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now();
UPDATE "public"."knowledge_base"
SET "visibility" = CASE WHEN user_id = 1 THEN 'public' ELSE 'private' END
WHERE "visibility" IS NULL OR btrim("visibility") = '';
UPDATE "public"."knowledge_base"
SET "status" = 'ready'
WHERE "status" IS NULL OR btrim("status") = '';
UPDATE "public"."knowledge_base"
SET "version" = 1
WHERE "version" IS NULL OR "version" <= 0;
UPDATE "public"."knowledge_base"
SET "updated_at" = COALESCE("updated_at", "created_at", now())
WHERE "updated_at" IS NULL;
CREATE INDEX idx_kb_user_id ON "public"."knowledge_base" (user_id);
DROP INDEX IF EXISTS idx_kb_document_identity;
CREATE INDEX idx_kb_document_identity ON "public"."knowledge_base" (user_id, title, source, version);
DROP INDEX IF EXISTS idx_kb_visibility_status;
CREATE INDEX idx_kb_visibility_status ON "public"."knowledge_base" (visibility, status, updated_at DESC);
DROP INDEX IF EXISTS idx_kb_embedding;
DO $$
    DECLARE
        embedding_udt text;
    BEGIN
        SELECT udt_name
        INTO embedding_udt
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'knowledge_base'
          AND column_name = 'embedding';

        IF embedding_udt = 'vector' THEN
            CREATE INDEX idx_kb_embedding ON "public"."knowledge_base" USING hnsw (embedding vector_l2_ops);
        ELSE
            RAISE NOTICE 'skip creating idx_kb_embedding: knowledge_base.embedding is still %', embedding_udt;
        END IF;
    END $$;

-- 2.5: 为 knowledge_base 表的所有字段添加/更新注释
COMMENT ON TABLE "public"."knowledge_base" IS '存储可复用的RAG数据源，包括公共和私有知识';
COMMENT ON COLUMN "public"."knowledge_base"."id" IS '知识条目的唯一标识符 (主键)';
COMMENT ON COLUMN "public"."knowledge_base"."user_id" IS '知识的所有者ID。业务约定 user_id = 1 为管理员上传的公共知识。';
COMMENT ON COLUMN "public"."knowledge_base"."title" IS '知识条目的标题，便于管理和识别';
COMMENT ON COLUMN "public"."knowledge_base"."content" IS '知识条目的原始文本内容';
COMMENT ON COLUMN "public"."knowledge_base"."embedding" IS '由文本内容生成的高维向量，用于相似度检索';
COMMENT ON COLUMN "public"."knowledge_base"."source" IS '知识来源，例如 PDF 文件名、Grok Search MCP 或手工资料包';
COMMENT ON COLUMN "public"."knowledge_base"."visibility" IS '知识可见性，当前支持 public/private';
COMMENT ON COLUMN "public"."knowledge_base"."status" IS '知识条目状态，例如 ready/failed/archived';
COMMENT ON COLUMN "public"."knowledge_base"."version" IS '同一 user_id + title + source 知识来源的上传批次版本号';
COMMENT ON COLUMN "public"."knowledge_base"."content_hash" IS '知识内容哈希，用于后续去重和重建索引';
COMMENT ON COLUMN "public"."knowledge_base"."created_at" IS '知识条目的创建时间戳';
COMMENT ON COLUMN "public"."knowledge_base"."updated_at" IS '知识条目的最近更新时间';


-- ----------------------------
-- 步骤 3: 升级 `vector_store` 表
-- ----------------------------
-- 3.1: 新增 user_id 字段 (允许为NULL)
ALTER TABLE "public"."vector_store" ADD COLUMN IF NOT EXISTS "user_id" BIGINT DEFAULT NULL;

-- 3.2: 将旧的 source_type 字段重命名为 doc_type (如果存在)
DO $$
    BEGIN
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vector_store' AND column_name='source_type')
           AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vector_store' AND column_name='doc_type') THEN
            ALTER TABLE "public"."vector_store" RENAME COLUMN "source_type" TO "doc_type";
        END IF;
    END $$;

-- 3.2.1: 为 doc_type 设置默认值
ALTER TABLE "public"."vector_store" ADD COLUMN IF NOT EXISTS "doc_type" VARCHAR(50);
UPDATE "public"."vector_store"
SET "doc_type" = COALESCE(NULLIF(btrim("doc_type"), ''), 'message')
WHERE "doc_type" IS NULL OR btrim("doc_type") = '';
ALTER TABLE "public"."vector_store" ALTER COLUMN "doc_type" SET DEFAULT 'message';
ALTER TABLE "public"."vector_store" ALTER COLUMN "doc_type" SET NOT NULL;


-- 3.3: 将 embedding 字段从 JSONB 转换为 vector
DO $$
    DECLARE
        embedding_udt text;
        array_rows bigint;
        string_rows bigint;
        other_rows bigint;
        existing_dims integer;
        distinct_dims integer;
    BEGIN
        SELECT udt_name
        INTO embedding_udt
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'vector_store'
          AND column_name = 'embedding';

        IF embedding_udt = 'jsonb' THEN
            SELECT COUNT(*) FILTER (WHERE jsonb_typeof(embedding) = 'array'),
                   COUNT(*) FILTER (WHERE jsonb_typeof(embedding) = 'string'),
                   COUNT(*) FILTER (WHERE jsonb_typeof(embedding) NOT IN ('array', 'string'))
            INTO array_rows, string_rows, other_rows
            FROM "public"."vector_store"
            WHERE embedding IS NOT NULL;

            IF array_rows = 0 AND string_rows = 0 THEN
                ALTER TABLE "public"."vector_store"
                    ALTER COLUMN "embedding" TYPE vector(1536) USING ("embedding"::text::vector);
            ELSIF array_rows > 0 AND string_rows = 0 AND other_rows = 0 THEN
                SELECT COUNT(DISTINCT jsonb_array_length(embedding)),
                       COALESCE(MAX(jsonb_array_length(embedding)), 1536)
                INTO distinct_dims, existing_dims
                FROM "public"."vector_store"
                WHERE embedding IS NOT NULL;

                IF distinct_dims > 1 THEN
                    RAISE NOTICE 'skip converting vector_store.embedding: inconsistent vector dimensions in existing jsonb array data';
                ELSIF existing_dims = 1536 THEN
                    ALTER TABLE "public"."vector_store"
                        ALTER COLUMN "embedding" TYPE vector(1536) USING ("embedding"::text::vector);
                ELSE
                    RAISE NOTICE 'skip converting vector_store.embedding: existing dimension % does not match expected 1536', existing_dims;
                END IF;
            ELSIF string_rows > 0 AND array_rows = 0 AND other_rows = 0 THEN
                BEGIN
                    SELECT COUNT(DISTINCT jsonb_array_length((convert_from(decode(trim(both '"' from embedding::text), 'base64'), 'UTF8'))::jsonb)),
                           COALESCE(MAX(jsonb_array_length((convert_from(decode(trim(both '"' from embedding::text), 'base64'), 'UTF8'))::jsonb)), 1536)
                    INTO distinct_dims, existing_dims
                    FROM "public"."vector_store"
                    WHERE embedding IS NOT NULL;

                    IF distinct_dims > 1 THEN
                        RAISE NOTICE 'skip converting vector_store.embedding: inconsistent vector dimensions in base64 jsonb string data';
                    ELSIF existing_dims = 1536 THEN
                        ALTER TABLE "public"."vector_store"
                            ALTER COLUMN "embedding" TYPE vector(1536)
                            USING ((convert_from(decode(trim(both '"' from embedding::text), 'base64'), 'UTF8'))::vector);
                    ELSE
                        RAISE NOTICE 'skip converting vector_store.embedding: existing base64 dimension % does not match expected 1536', existing_dims;
                    END IF;
                EXCEPTION
                    WHEN OTHERS THEN
                        RAISE NOTICE 'skip converting vector_store.embedding: unsupported jsonb string payload: %', SQLERRM;
                END;
            ELSE
                RAISE NOTICE 'skip converting vector_store.embedding: mixed jsonb payload shapes detected';
            END IF;
        END IF;
    END $$;

-- 3.4: 添加外键约束 (如果不存在)
DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_vs_user_id') THEN
            ALTER TABLE "public"."vector_store" ADD CONSTRAINT fk_vs_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        END IF;
    END $$;

-- 3.5: 创建新索引 (使用 DROP IF EXISTS + CREATE 模式)
-- 用于高效查询某登录用户的简历或对话历史
DROP INDEX IF EXISTS idx_vs_user_id_type;
CREATE INDEX idx_vs_user_id_type ON "public"."vector_store" (user_id, doc_type);
-- 用于快速加载和回放特定会话的上下文
DROP INDEX IF EXISTS idx_vs_chat_id;
CREATE INDEX idx_vs_chat_id ON "public"."vector_store" (chat_id);
DROP INDEX IF EXISTS idx_vs_chat_user_type;
CREATE INDEX idx_vs_chat_user_type ON "public"."vector_store" (chat_id, user_id, doc_type);
-- 【新增】用于按时间排序和高效地执行定期的数据清理任务
DROP INDEX IF EXISTS idx_vs_created_at;
CREATE INDEX idx_vs_created_at ON "public"."vector_store" (created_at DESC);
-- 用于高性能向量相似度搜索
DROP INDEX IF EXISTS idx_vs_embedding;
DO $$
    DECLARE
        embedding_udt text;
    BEGIN
        SELECT udt_name
        INTO embedding_udt
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'vector_store'
          AND column_name = 'embedding';

        IF embedding_udt = 'vector' THEN
            CREATE INDEX idx_vs_embedding ON "public"."vector_store" USING hnsw (embedding vector_l2_ops);
        ELSE
            RAISE NOTICE 'skip creating idx_vs_embedding: vector_store.embedding is still %', embedding_udt;
        END IF;
    END $$;


-- 3.6: 为 vector_store 表的所有字段添加/更新注释
COMMENT ON TABLE "public"."vector_store" IS '存储所有与具体对话会话相关的瞬时数据，包括对话历史和用户上传的简历';
COMMENT ON COLUMN "public"."vector_store"."id" IS '向量记录的唯一标识符 (主键)';
COMMENT ON COLUMN "public"."vector_store"."user_id" IS '所属用户的ID；对于匿名用户的会话，此字段为NULL。';
COMMENT ON COLUMN "public"."vector_store"."chat_id" IS '一次完整对话的唯一会话ID，用于关联上下文';
COMMENT ON COLUMN "public"."vector_store"."role" IS '消息发送者的角色 (例如: "user", "assistant")';
COMMENT ON COLUMN "public"."vector_store"."content" IS '消息或文档（如简历片段）的原始文本内容';
COMMENT ON COLUMN "public"."vector_store"."embedding" IS '文本内容的向量化表示';
COMMENT ON COLUMN "public"."vector_store"."doc_type" IS '文档类型，区分为 "message" (对话消息) 和 "resume" (简历内容)';
COMMENT ON COLUMN "public"."vector_store"."created_at" IS '记录的创建时间戳，可用于TTL清理';


-- ----------------------------
-- 步骤 4: 正式化 `chat_sessions` 表
-- ----------------------------
CREATE TABLE IF NOT EXISTS "public"."chat_sessions" (
                                                        "id" BIGSERIAL PRIMARY KEY,
                                                        "session_id" VARCHAR(64) UNIQUE NOT NULL,
                                                        "user_id" BIGINT NOT NULL,
                                                        "title" VARCHAR(200) NOT NULL DEFAULT '新对话',
                                                        "mode" VARCHAR(64) NOT NULL DEFAULT 'Interview',
                                                        "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
                                                        "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
                                                        "last_message_at" TIMESTAMPTZ,
                                                        "message_count" INTEGER NOT NULL DEFAULT 0,
                                                        "is_active" BOOLEAN NOT NULL DEFAULT true
);

DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_chat_sessions_user_id') THEN
            ALTER TABLE "public"."chat_sessions"
                ADD CONSTRAINT fk_chat_sessions_user_id
                    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        END IF;
    END $$;

DROP INDEX IF EXISTS idx_chat_sessions_session_id;
CREATE UNIQUE INDEX idx_chat_sessions_session_id ON "public"."chat_sessions" (session_id);
ALTER TABLE "public"."chat_sessions" ADD COLUMN IF NOT EXISTS "mode" VARCHAR(64);
UPDATE "public"."chat_sessions"
SET "mode" = 'Interview'
WHERE "mode" IS NULL OR btrim("mode") = '';
ALTER TABLE "public"."chat_sessions" ALTER COLUMN "mode" SET NOT NULL;
ALTER TABLE "public"."chat_sessions" ALTER COLUMN "mode" SET DEFAULT 'Interview';
ALTER TABLE "public"."chat_sessions" ADD COLUMN IF NOT EXISTS "direction_key" VARCHAR(64) NOT NULL DEFAULT 'go_backend';
ALTER TABLE "public"."chat_sessions" ADD COLUMN IF NOT EXISTS "direction_label" VARCHAR(80) NOT NULL DEFAULT 'Go 后端';
ALTER TABLE "public"."chat_sessions" ADD COLUMN IF NOT EXISTS "difficulty_level" INTEGER NOT NULL DEFAULT 3;
ALTER TABLE "public"."chat_sessions" ADD COLUMN IF NOT EXISTS "difficulty_label" VARCHAR(32) NOT NULL DEFAULT '中级';
ALTER TABLE "public"."chat_sessions" ADD COLUMN IF NOT EXISTS "interviewer_style" VARCHAR(64) NOT NULL DEFAULT 'senior';
ALTER TABLE "public"."chat_sessions" ADD COLUMN IF NOT EXISTS "interviewer_style_label" VARCHAR(80) NOT NULL DEFAULT '资深技术官';
ALTER TABLE "public"."chat_sessions" ADD COLUMN IF NOT EXISTS "focus_areas" JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE "public"."chat_sessions" ADD COLUMN IF NOT EXISTS "follow_up_depth" VARCHAR(16) NOT NULL DEFAULT 'N+3';
ALTER TABLE "public"."chat_sessions" ADD COLUMN IF NOT EXISTS "estimated_minutes" INTEGER NOT NULL DEFAULT 30;
ALTER TABLE "public"."chat_sessions" ADD COLUMN IF NOT EXISTS "progress_percent" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "public"."chat_sessions" ADD COLUMN IF NOT EXISTS "started_at" TIMESTAMPTZ;
ALTER TABLE "public"."chat_sessions" ADD COLUMN IF NOT EXISTS "completed_at" TIMESTAMPTZ;
ALTER TABLE "public"."chat_sessions" ADD COLUMN IF NOT EXISTS "duration_seconds" INTEGER NOT NULL DEFAULT 0;
DROP INDEX IF EXISTS idx_chat_sessions_user_last_message;
CREATE INDEX idx_chat_sessions_user_last_message ON "public"."chat_sessions" (user_id, last_message_at DESC);
DROP INDEX IF EXISTS idx_chat_sessions_user_active;
CREATE INDEX idx_chat_sessions_user_active ON "public"."chat_sessions" (user_id, is_active);
DROP INDEX IF EXISTS idx_chat_sessions_user_completed_at;
CREATE INDEX idx_chat_sessions_user_completed_at ON "public"."chat_sessions" (user_id, completed_at DESC);

COMMENT ON TABLE "public"."chat_sessions" IS '存储用户工作台中的会话元数据，用于会话列表、恢复与排序';
COMMENT ON COLUMN "public"."chat_sessions"."id" IS '会话元数据主键';
COMMENT ON COLUMN "public"."chat_sessions"."session_id" IS '对外暴露的会话ID，对应前端 chatId';
COMMENT ON COLUMN "public"."chat_sessions"."user_id" IS '会话所属用户ID';
COMMENT ON COLUMN "public"."chat_sessions"."title" IS '会话标题，默认由首条用户消息或默认值生成';
COMMENT ON COLUMN "public"."chat_sessions"."mode" IS '会话所属工作模式，当前规范值为 Interview/Research/Memory/Coach';
COMMENT ON COLUMN "public"."chat_sessions"."direction_key" IS '面试方向键，例如 go_backend/system_design/frontend_vue';
COMMENT ON COLUMN "public"."chat_sessions"."direction_label" IS '面试方向展示名';
COMMENT ON COLUMN "public"."chat_sessions"."difficulty_level" IS '面试难度等级，范围 1-5';
COMMENT ON COLUMN "public"."chat_sessions"."difficulty_label" IS '面试难度展示名';
COMMENT ON COLUMN "public"."chat_sessions"."interviewer_style" IS '面试官人格键，例如 senior/pressure/humorous';
COMMENT ON COLUMN "public"."chat_sessions"."interviewer_style_label" IS '面试官人格展示名';
COMMENT ON COLUMN "public"."chat_sessions"."focus_areas" IS '面试侧重点 JSON 数组，保存创建会话时的配置快照';
COMMENT ON COLUMN "public"."chat_sessions"."follow_up_depth" IS '追问深度标签，例如 N+3/N+5';
COMMENT ON COLUMN "public"."chat_sessions"."estimated_minutes" IS '预计面试时长，单位分钟';
COMMENT ON COLUMN "public"."chat_sessions"."progress_percent" IS '当前面试进度百分比，用于工作台继续入口';
COMMENT ON COLUMN "public"."chat_sessions"."created_at" IS '会话创建时间';
COMMENT ON COLUMN "public"."chat_sessions"."updated_at" IS '会话最近一次元数据更新时间';
COMMENT ON COLUMN "public"."chat_sessions"."last_message_at" IS '会话最近一条消息时间';
COMMENT ON COLUMN "public"."chat_sessions"."started_at" IS '面试开始时间';
COMMENT ON COLUMN "public"."chat_sessions"."completed_at" IS '面试完成时间';
COMMENT ON COLUMN "public"."chat_sessions"."duration_seconds" IS '面试持续时长，单位秒';
COMMENT ON COLUMN "public"."chat_sessions"."message_count" IS '会话消息条数，用于工作台统计';
COMMENT ON COLUMN "public"."chat_sessions"."is_active" IS '会话是否仍处于活跃状态';

-- ----------------------------
-- 步骤 4.5: 新增 `resume_documents` 表
-- ----------------------------
CREATE TABLE IF NOT EXISTS "public"."resume_documents" (
                                                          "id" BIGSERIAL PRIMARY KEY,
                                                          "user_id" BIGINT NOT NULL,
                                                          "session_id" VARCHAR(64) NOT NULL,
                                                          "version" BIGINT NOT NULL DEFAULT 1,
                                                          "title" VARCHAR(200) NOT NULL,
                                                          "filename" VARCHAR(255) NOT NULL,
                                                          "status" VARCHAR(32) NOT NULL DEFAULT 'ready',
                                                          "chunk_count" INTEGER NOT NULL DEFAULT 0,
                                                          "is_current" BOOLEAN NOT NULL DEFAULT true,
                                                          "uploaded_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
                                                          "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
                                                          UNIQUE ("user_id", "session_id", "version")
);

DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_resume_documents_user_id') THEN
            ALTER TABLE "public"."resume_documents"
                ADD CONSTRAINT fk_resume_documents_user_id
                    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        END IF;
    END $$;

DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_resume_documents_session_id') THEN
            ALTER TABLE "public"."resume_documents"
                ADD CONSTRAINT fk_resume_documents_session_id
                    FOREIGN KEY (session_id) REFERENCES "public"."chat_sessions"(session_id) ON DELETE CASCADE NOT VALID;
        END IF;
    END $$;

DO $$
    BEGIN
        IF EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE conname = 'fk_resume_documents_session_id'
              AND NOT convalidated
        ) THEN
            BEGIN
                ALTER TABLE "public"."resume_documents"
                    VALIDATE CONSTRAINT fk_resume_documents_session_id;
            EXCEPTION
                WHEN OTHERS THEN
                    RAISE NOTICE 'skip validating fk_resume_documents_session_id: %', SQLERRM;
            END;
        END IF;
    END $$;

DROP INDEX IF EXISTS idx_resume_documents_user_current;
CREATE INDEX idx_resume_documents_user_current ON "public"."resume_documents" (user_id, is_current, updated_at DESC);
DROP INDEX IF EXISTS idx_resume_documents_session_current;
CREATE INDEX idx_resume_documents_session_current ON "public"."resume_documents" (session_id, user_id, is_current);

COMMENT ON TABLE "public"."resume_documents" IS '存储用户简历资料的版本化元数据，向量分块仍保存在 vector_store 中';
COMMENT ON COLUMN "public"."resume_documents"."id" IS '简历资料版本主键';
COMMENT ON COLUMN "public"."resume_documents"."user_id" IS '简历所属用户ID';
COMMENT ON COLUMN "public"."resume_documents"."session_id" IS '该简历当前绑定的会话ID';
COMMENT ON COLUMN "public"."resume_documents"."version" IS '同一用户同一会话下的简历版本号';
COMMENT ON COLUMN "public"."resume_documents"."title" IS '简历资料展示标题';
COMMENT ON COLUMN "public"."resume_documents"."filename" IS '上传的原始文件名';
COMMENT ON COLUMN "public"."resume_documents"."status" IS '解析状态，例如 ready/failed';
COMMENT ON COLUMN "public"."resume_documents"."chunk_count" IS '本版本写入 vector_store 的简历分块数量';
COMMENT ON COLUMN "public"."resume_documents"."is_current" IS '是否为当前会话正在使用的简历版本';
COMMENT ON COLUMN "public"."resume_documents"."uploaded_at" IS '本版本上传时间';
COMMENT ON COLUMN "public"."resume_documents"."updated_at" IS '本版本最近更新时间';

-- ----------------------------
-- 步骤 5: 新增 `session_evaluations` 表
-- ----------------------------
CREATE TABLE IF NOT EXISTS "public"."session_evaluations" (
                                                              "id" BIGSERIAL PRIMARY KEY,
                                                              "session_id" VARCHAR(64) NOT NULL,
                                                              "user_id" BIGINT NOT NULL,
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
                                                              "first_generated_at" TIMESTAMPTZ,
                                                              "generated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
                                                              "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
                                                              UNIQUE ("session_id", "user_id")
);

DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_session_evaluations_user_id') THEN
            ALTER TABLE "public"."session_evaluations"
                ADD CONSTRAINT fk_session_evaluations_user_id
                    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        END IF;
    END $$;

DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_session_evaluations_session_id') THEN
            ALTER TABLE "public"."session_evaluations"
                ADD CONSTRAINT fk_session_evaluations_session_id
                    FOREIGN KEY (session_id) REFERENCES "public"."chat_sessions"(session_id) ON DELETE CASCADE NOT VALID;
        END IF;
    END $$;

DO $$
    BEGIN
        IF EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE conname = 'fk_session_evaluations_session_id'
              AND NOT convalidated
        ) THEN
            BEGIN
                ALTER TABLE "public"."session_evaluations"
                    VALIDATE CONSTRAINT fk_session_evaluations_session_id;
            EXCEPTION
                WHEN OTHERS THEN
                    RAISE NOTICE 'skip validating fk_session_evaluations_session_id: %', SQLERRM;
            END;
        END IF;
    END $$;

ALTER TABLE "public"."session_evaluations" ADD COLUMN IF NOT EXISTS "overall_score" DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE "public"."session_evaluations" ADD COLUMN IF NOT EXISTS "rubric_version" VARCHAR(32) NOT NULL DEFAULT 'rubric-v1';
ALTER TABLE "public"."session_evaluations" ADD COLUMN IF NOT EXISTS "score_source" VARCHAR(32) NOT NULL DEFAULT 'heuristic';
ALTER TABLE "public"."session_evaluations" ADD COLUMN IF NOT EXISTS "suggestions" JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE "public"."session_evaluations" ADD COLUMN IF NOT EXISTS "source_last_message_id" BIGINT;
ALTER TABLE "public"."session_evaluations" ADD COLUMN IF NOT EXISTS "source_last_message_at" TIMESTAMPTZ;
ALTER TABLE "public"."session_evaluations" ADD COLUMN IF NOT EXISTS "first_generated_at" TIMESTAMPTZ;
UPDATE "public"."session_evaluations"
SET "first_generated_at" = COALESCE("first_generated_at", "generated_at", "updated_at", now())
WHERE "first_generated_at" IS NULL;
ALTER TABLE "public"."session_evaluations" ALTER COLUMN "first_generated_at" SET NOT NULL;
ALTER TABLE "public"."session_evaluations" ALTER COLUMN "first_generated_at" SET DEFAULT now();
ALTER TABLE "public"."session_evaluations" ALTER COLUMN "status" SET DEFAULT 'draft';
ALTER TABLE "public"."session_evaluations" ALTER COLUMN "summary" SET DEFAULT '';

DROP INDEX IF EXISTS idx_session_evaluations_user_id;
CREATE INDEX idx_session_evaluations_user_id ON "public"."session_evaluations" (user_id);
DROP INDEX IF EXISTS idx_session_evaluations_session_id;
CREATE INDEX idx_session_evaluations_session_id ON "public"."session_evaluations" (session_id);

COMMENT ON TABLE "public"."session_evaluations" IS '存储会话维度的结构化评估结果，为工作台和报告中心提供稳定数据源';
COMMENT ON COLUMN "public"."session_evaluations"."id" IS '评估记录主键';
COMMENT ON COLUMN "public"."session_evaluations"."session_id" IS '关联的会话ID';
COMMENT ON COLUMN "public"."session_evaluations"."user_id" IS '评估所属用户ID';
COMMENT ON COLUMN "public"."session_evaluations"."status" IS '评估状态，如 draft/ready/insufficient_data';
COMMENT ON COLUMN "public"."session_evaluations"."summary" IS '评估摘要';
COMMENT ON COLUMN "public"."session_evaluations"."user_turns" IS '用户消息轮次';
COMMENT ON COLUMN "public"."session_evaluations"."assistant_turns" IS '助手消息轮次';
COMMENT ON COLUMN "public"."session_evaluations"."overall_score" IS '按 rubric 计算后的综合得分';
COMMENT ON COLUMN "public"."session_evaluations"."rubric_version" IS '当前评估所使用的 rubric 版本';
COMMENT ON COLUMN "public"."session_evaluations"."score_source" IS '综合评分来源，例如 llm / heuristic / mixed';
COMMENT ON COLUMN "public"."session_evaluations"."dimensions" IS '结构化评估维度 JSON';
COMMENT ON COLUMN "public"."session_evaluations"."strengths" IS '优势摘要 JSON 数组';
COMMENT ON COLUMN "public"."session_evaluations"."risks" IS '风险摘要 JSON 数组';
COMMENT ON COLUMN "public"."session_evaluations"."suggestions" IS '可执行建议 JSON 数组';
COMMENT ON COLUMN "public"."session_evaluations"."evidence" IS '评估证据片段 JSON 数组';
COMMENT ON COLUMN "public"."session_evaluations"."source_last_message_id" IS '当前评估结果实际覆盖到的最新消息ID水位';
COMMENT ON COLUMN "public"."session_evaluations"."source_last_message_at" IS '当前评估结果实际覆盖到的最新消息时间水位';
COMMENT ON COLUMN "public"."session_evaluations"."first_generated_at" IS '当前会话评估结果首次生成的时间';
COMMENT ON COLUMN "public"."session_evaluations"."generated_at" IS '最近一次生成当前评估结果的时间（接口兼容字段来源）';
COMMENT ON COLUMN "public"."session_evaluations"."updated_at" IS '当前会话评估结果最近一次刷新时间';

-- ----------------------------
-- 步骤 5.5: 新增 `session_evaluation_items` 表
-- ----------------------------
CREATE TABLE IF NOT EXISTS "public"."session_evaluation_items" (
                                                                  "id" BIGSERIAL PRIMARY KEY,
                                                                  "session_id" VARCHAR(64) NOT NULL,
                                                                  "user_id" BIGINT NOT NULL,
                                                                  "turn_index" INTEGER NOT NULL,
                                                                  "question" TEXT NOT NULL DEFAULT '',
                                                                  "answer" TEXT NOT NULL DEFAULT '',
                                                                  "ai_comment" TEXT NOT NULL DEFAULT '',
                                                                  "score" DOUBLE PRECISION NOT NULL DEFAULT 0,
                                                                  "max_score" DOUBLE PRECISION NOT NULL DEFAULT 5,
                                                                  "tags" JSONB NOT NULL DEFAULT '[]'::jsonb,
                                                                  "source_message_id" BIGINT,
                                                                  "source_message_at" TIMESTAMPTZ,
                                                                  "generated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
                                                                  "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
                                                                  UNIQUE ("session_id", "user_id", "turn_index")
);

DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_session_evaluation_items_user_id') THEN
            ALTER TABLE "public"."session_evaluation_items"
                ADD CONSTRAINT fk_session_evaluation_items_user_id
                    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        END IF;
    END $$;

DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_session_evaluation_items_session_id') THEN
            ALTER TABLE "public"."session_evaluation_items"
                ADD CONSTRAINT fk_session_evaluation_items_session_id
                    FOREIGN KEY (session_id) REFERENCES "public"."chat_sessions"(session_id) ON DELETE CASCADE NOT VALID;
        END IF;
    END $$;

DO $$
    BEGIN
        IF EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE conname = 'fk_session_evaluation_items_session_id'
              AND NOT convalidated
        ) THEN
            BEGIN
                ALTER TABLE "public"."session_evaluation_items"
                    VALIDATE CONSTRAINT fk_session_evaluation_items_session_id;
            EXCEPTION
                WHEN OTHERS THEN
                    RAISE NOTICE 'skip validating fk_session_evaluation_items_session_id: %', SQLERRM;
            END;
        END IF;
    END $$;

DROP INDEX IF EXISTS idx_session_evaluation_items_user_session;
CREATE INDEX idx_session_evaluation_items_user_session ON "public"."session_evaluation_items" (user_id, session_id, turn_index);

COMMENT ON TABLE "public"."session_evaluation_items" IS '存储会话评估的逐题卡片快照，用于报告详情复盘';
COMMENT ON COLUMN "public"."session_evaluation_items"."id" IS '逐题评估明细主键';
COMMENT ON COLUMN "public"."session_evaluation_items"."session_id" IS '关联的会话ID';
COMMENT ON COLUMN "public"."session_evaluation_items"."user_id" IS '评估明细所属用户ID';
COMMENT ON COLUMN "public"."session_evaluation_items"."turn_index" IS '用户回答轮次，从 1 开始';
COMMENT ON COLUMN "public"."session_evaluation_items"."question" IS '该轮回答对应的面试官问题';
COMMENT ON COLUMN "public"."session_evaluation_items"."answer" IS '用户回答摘要';
COMMENT ON COLUMN "public"."session_evaluation_items"."ai_comment" IS 'AI 对该轮回答的点评';
COMMENT ON COLUMN "public"."session_evaluation_items"."score" IS '该轮回答评分';
COMMENT ON COLUMN "public"."session_evaluation_items"."max_score" IS '该轮回答满分';
COMMENT ON COLUMN "public"."session_evaluation_items"."tags" IS '该轮回答标签 JSON 数组';
COMMENT ON COLUMN "public"."session_evaluation_items"."source_message_id" IS '对应的用户消息ID';
COMMENT ON COLUMN "public"."session_evaluation_items"."source_message_at" IS '对应的用户消息时间';
COMMENT ON COLUMN "public"."session_evaluation_items"."generated_at" IS '该明细生成时间';
COMMENT ON COLUMN "public"."session_evaluation_items"."updated_at" IS '该明细最近更新时间';

COMMIT; -- 提交事务，所有更改生效

