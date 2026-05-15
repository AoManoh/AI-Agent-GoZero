# GoZero-AI

GoZero-AI 是一个基于 GoZero 微服务架构实现的 AI 面试与知识库增强对话系统。项目把大语言模型对话、RAG 检索增强、PDF 文档解析、用户认证、简历分析、面试会话管理和报告复盘整合到同一个可本地运行的工程中。

当前公开仓库包含三个后端服务和一个前端应用：

- `api/chat`：面试对话与知识库 RAG 服务，提供 SSE 流式对话、知识库上传、文本导入、目录管理和检索测试接口。
- `api/user`：用户域服务，提供注册登录、JWT/RefreshToken、工作台、题库、面试会话、简历上传与报告中心接口。
- `mcp`：PDF 文本提取 gRPC 服务，供知识库和简历解析链路调用。
- `client`：Vue 3 + Vite + JavaScript 前端，包含首页、登录注册、聊天面试、工作台、题库、知识库、简历管理和报告中心页面。

## 核心能力

- AI 面试官：基于会话状态、面试方向、难度、简历和侧重点组织追问。
- 流式对话：Chat API 通过 SSE 返回面试官回复，并记录会话过程。
- RAG 知识库：支持 PDF 上传、文本导入、知识库目录、文档 chunks 与检索测试。
- 简历管理：支持 PDF 简历上传、文本抽取、结构化分析和面试追问素材沉淀。
- 题库浏览：按方向、难度、能力维度浏览面试题，并为面试计划提供候选题。
- 工作台：聚合最近面试、能力雷达、面试入口、题库、知识库、简历和报告中心。
- 报告中心：汇总已完成面试的评分、模式表现、单场报告和复盘建议。
- 用户认证：User API 提供注册、登录、刷新 token、登出和 profile 能力。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.24、GoZero、gRPC |
| 前端 | Vue 3、Vite、JavaScript、Element Plus |
| 数据 | PostgreSQL、pgvector、Redis |
| 服务发现 | ETCD |
| AI 接口 | OpenAI 兼容接口、DashScope 兼容接口 |
| 文档解析 | UniPDF、MCP gRPC |
| 部署 | Docker、Docker Compose、Nginx |

## 架构概览

```text
Browser / Vite / Nginx
        |
        +-- /api/users/* --> User API (:8124)
        |                       |
        |                       +-- PostgreSQL
        |                       +-- Redis
        |                       +-- MCP gRPC
        |
        +-- /api/ai/* ----> Chat API (:8123)
                                |
                                +-- PostgreSQL + pgvector
                                +-- Redis
                                +-- MCP gRPC
                                +-- OpenAI-compatible LLM / Embedding

MCP gRPC (:8080 local, :8321 docker)
        |
        +-- PDF text extraction
        +-- ETCD service discovery
```

## 目录结构

```text
.
├── api/
│   ├── chat/                  # Chat API：SSE 面试对话、RAG 知识库
│   └── user/                  # User API：认证、工作台、简历、报告、题库
├── client/                    # Vue 3 + Vite 前端
├── cmd/pdfparse/              # PDF 解析调试命令
├── db/                        # PostgreSQL 初始化脚本
├── internal/                  # 后端共享包
├── mcp/                       # PDF 文本提取 gRPC 服务
├── docker-compose.example.yml # 完整服务编排示例
└── docker-compose.client.slim.yml
```

## 服务与端口

| 服务 | 默认端口 | 说明 |
|------|----------|------|
| Frontend dev server | `3000` | Vite 开发服务器 |
| Chat API | `8123` | `/api/ai/*`，SSE 对话与知识库 |
| User API | `8124` | `/api/users/*`，认证、工作台、简历、报告 |
| MCP local | `8080` | 本地 `mcp/etc/mcp.yaml` 默认 gRPC 端口 |
| MCP docker | `8321` | `mcp/etc/docker.yaml` 默认 gRPC 端口 |
| PostgreSQL | `5432` | 主库与 pgvector |
| Redis | `6379` | token、会话和状态数据 |
| ETCD | `2379` | GoZero 服务注册发现 |

## 配置准备

仓库只提交示例配置。首次运行前复制示例文件并填写数据库密码、JWT secret、模型 key、UniPDF license 等敏感配置。

```bash
cp docker-compose.example.yml docker-compose.yml

cp api/chat/etc/chat.example.yaml api/chat/etc/chat.yaml
cp api/chat/etc/docker.example.yaml api/chat/etc/docker.yaml

cp api/user/etc/user-api.example.yaml api/user/etc/user-api.yaml
cp api/user/etc/docker.example.yaml api/user/etc/docker.yaml

cp mcp/etc/mcp.example.yaml mcp/etc/mcp.yaml
cp mcp/etc/docker.example.yaml mcp/etc/docker.yaml

cp client/.env.example client/.env.development
```

请确认 `docker-compose.yml` 中的 `POSTGRES_USER`、`POSTGRES_PASSWORD`、`POSTGRES_DB` 与 `api/chat/etc/*.yaml`、`api/user/etc/*.yaml` 中的数据库连接配置保持一致。

前端开发环境默认通过 Vite 代理访问后端：

- `/api/users/*` 转发到 `http://127.0.0.1:8124`
- `/api/ai/*` 转发到 `http://127.0.0.1:8123`

如需修改代理目标，编辑 `client/.env.development`：

```env
VITE_API_BASE_URL=/api
VITE_USER_API_BASE_URL=/api
VITE_CHAT_API_BASE_URL=/api
VITE_DEV_USER_PROXY_TARGET=http://127.0.0.1:8124
VITE_DEV_CHAT_PROXY_TARGET=http://127.0.0.1:8123
```

## 本地开发启动

### 1. 启动基础依赖

复制并确认 `docker-compose.yml` 后，可以只启动基础设施：

```bash
docker compose up -d etcd redis postgres
```

### 2. 启动 MCP 服务

```bash
go run mcp/mcp.go -f mcp/etc/mcp.yaml
```

### 3. 启动 Chat API

```bash
go run api/chat/chat.go -f api/chat/etc/chat.yaml
```

### 4. 启动 User API

```bash
go run api/user/user.go -f api/user/etc/user-api.yaml
```

### 5. 启动前端

```bash
cd client
npm install
npm run dev
```

默认访问地址：

```text
http://localhost:3000
```

## Docker Compose 启动

准备好 `docker-compose.yml` 和各服务的 `docker.yaml` 后，可以直接启动完整服务栈：

```bash
docker compose up -d
```

查看状态与日志：

```bash
docker compose ps
docker compose logs -f api
docker compose logs -f user
docker compose logs -f mcp
```

如果只想叠加前端 Nginx 静态服务：

```bash
cd client
npm install
npm run build
cd ..
docker compose -f docker-compose.yml -f docker-compose.client.slim.yml up -d client
```

`docker-compose.client.slim.yml` 默认把前端发布到宿主机 `8080` 端口。

## 主要 HTTP 接口

接口契约以 `api/chat/chat.api` 和 `api/user/user.api` 为准。

### Chat API

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/ai/interview_app/chat/sse` | 面试聊天 SSE |
| `POST` | `/api/ai/knowledge/upload` | 上传知识库 PDF |
| `POST` | `/api/ai/knowledge/text` | 导入知识库文本 |
| `GET` | `/api/ai/knowledge/folders` | 查询知识库目录 |
| `GET` | `/api/ai/knowledge/documents` | 查询知识库文档 |
| `GET` | `/api/ai/knowledge/documents/:id/chunks` | 查询文档分块 |
| `POST` | `/api/ai/knowledge/test-query` | RAG 检索测试 |

### User API

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/users/register` | 注册 |
| `POST` | `/api/users/login` | 登录 |
| `POST` | `/api/users/refresh` | 刷新 token |
| `GET` | `/api/users/workbench/bootstrap` | 工作台聚合数据 |
| `GET` | `/api/users/interview/questions` | 面试题列表 |
| `POST` | `/api/users/sessions` | 创建面试会话 |
| `GET` | `/api/users/sessions/:id/bootstrap` | 会话启动数据 |
| `POST` | `/api/users/sessions/:id/finish` | 结束面试 |
| `GET` | `/api/users/sessions/:id/report` | 单场报告 |
| `POST` | `/api/users/resume/upload` | 上传简历 PDF |
| `GET` | `/api/users/report-center/bootstrap` | 报告中心聚合数据 |

## 前端路由

| 路由 | 页面 |
|------|------|
| `/` | 首页 |
| `/login` | 登录 |
| `/register` | 注册 |
| `/chat` | AI 面试聊天 |
| `/workbench` | 工作台 |
| `/workbench/new` | 新建面试 |
| `/workbench/bank` | 题库浏览 |
| `/workbench/knowledge` | 知识库 |
| `/workbench/resume` | 简历管理 |
| `/workbench/reports` | 报告中心 |

## 常用验证

```bash
go test ./...

cd client
npm run build
```

## 安全说明

- 不要把真实 API Key、数据库密码、JWT secret、UniPDF license 提交到仓库。
- `*.example.yaml` 和 `.env.example` 只用于说明配置结构。
- 生产环境应使用环境变量、密钥管理服务或受控配置文件注入敏感信息。

## 许可证

本项目采用 MIT License，详见 `LICENSE`。
