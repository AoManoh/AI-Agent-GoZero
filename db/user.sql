-- =================================================================
-- 数据库安全升级脚本 v4 (非破坏性, 已修复索引兼容性问题)
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
            INSERT INTO "public"."users" (id, username, password_hash) VALUES (1, 'your_username', MD5('your_password')); -- 请替换为真实的哈希密码
        END IF;
    END $$;


-- ----------------------------
-- 步骤 2: 升级 `knowledge_base` 表
-- ----------------------------
-- 2.1: 新增 user_id 字段，并假设所有现有知识都属于管理员(id=1)
ALTER TABLE "public"."knowledge_base" ADD COLUMN IF NOT EXISTS "user_id" BIGINT NOT NULL DEFAULT 1;

-- 2.2: 将 embedding 字段从 JSONB 转换为 vector
ALTER TABLE "public"."knowledge_base" ALTER COLUMN "embedding" TYPE vector(1536) USING ("embedding"::text::vector);


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
CREATE INDEX idx_kb_user_id ON "public"."knowledge_base" (user_id);
DROP INDEX IF EXISTS idx_kb_embedding;
CREATE INDEX idx_kb_embedding ON "public"."knowledge_base" USING hnsw (embedding vector_l2_ops);

-- 2.5: 为 knowledge_base 表的所有字段添加/更新注释
COMMENT ON TABLE "public"."knowledge_base" IS '存储可复用的RAG数据源，包括公共和私有知识';
COMMENT ON COLUMN "public"."knowledge_base"."id" IS '知识条目的唯一标识符 (主键)';
COMMENT ON COLUMN "public"."knowledge_base"."user_id" IS '知识的所有者ID。业务约定 user_id = 1 为管理员上传的公共知识。';
COMMENT ON COLUMN "public"."knowledge_base"."title" IS '知识条目的标题，便于管理和识别';
COMMENT ON COLUMN "public"."knowledge_base"."content" IS '知识条目的原始文本内容';
COMMENT ON COLUMN "public"."knowledge_base"."embedding" IS '由文本内容生成的高维向量，用于相似度检索';
COMMENT ON COLUMN "public"."knowledge_base"."created_at" IS '知识条目的创建时间戳';


-- ----------------------------
-- 步骤 3: 升级 `vector_store` 表
-- ----------------------------
-- 3.1: 新增 user_id 字段 (允许为NULL)
ALTER TABLE "public"."vector_store" ADD COLUMN IF NOT EXISTS "user_id" BIGINT DEFAULT NULL;

-- 3.2: 将旧的 source_type 字段重命名为 doc_type (如果存在)
DO $$
    BEGIN
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vector_store' AND column_name='source_type') THEN
            ALTER TABLE "public"."vector_store" RENAME COLUMN "source_type" TO "doc_type";
        END IF;
    END $$;

-- 3.2.1: 为 doc_type 设置默认值
ALTER TABLE "public"."vector_store" ALTER COLUMN "doc_type" SET DEFAULT 'message';


-- 3.3: 将 embedding 字段从 JSONB 转换为 vector
ALTER TABLE "public"."vector_store" ALTER COLUMN "embedding" TYPE vector(1536) USING ("embedding"::text::vector);

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
-- 【新增】用于按时间排序和高效地执行定期的数据清理任务
DROP INDEX IF EXISTS idx_vs_created_at;
CREATE INDEX idx_vs_created_at ON "public"."vector_store" (created_at DESC);
-- 用于高性能向量相似度搜索
DROP INDEX IF EXISTS idx_vs_embedding;
CREATE INDEX idx_vs_embedding ON "public"."vector_store" USING hnsw (embedding vector_l2_ops);


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

COMMIT; -- 提交事务，所有更改生效

