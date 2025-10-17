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
DROP TABLE IF EXISTS "public"."vector_store";
DROP TABLE IF EXISTS "public"."knowledge_base";

-- 创建新表
CREATE TABLE "public"."vector_store" (
                                         "id" BIGSERIAL PRIMARY KEY,
                                         "chat_id" varchar(255) NOT NULL,
                                         "role" varchar(50) NOT NULL,   -- 新增角色字段
                                         "content" TEXT NOT NULL,
                                         "embedding" JSONB NOT NULL,
                                         "source_type" VARCHAR(50) NOT NULL DEFAULT 'message',
                                         "created_at" TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 创建索引
CREATE INDEX idx_vector_store_chat_id ON vector_store (chat_id);
CREATE INDEX idx_vector_store_created_at ON vector_store (created_at DESC);

-- 创建知识库内容表
CREATE TABLE "public"."knowledge_base" (
                                           "id" BIGSERIAL PRIMARY KEY,
                                           "title" VARCHAR(255) NOT NULL,
                                           "content" TEXT NOT NULL,
                                           "embedding" JSONB NOT NULL,
                                           "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 创建索引
CREATE INDEX idx_knowledge_base_title ON knowledge_base (title);