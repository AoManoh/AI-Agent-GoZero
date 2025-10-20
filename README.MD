# GoZero-AI 快速启动指南

## 项目概述

GoZero-AI 是一个基于 **GoZero 微服务框架**构建的企业级 AI 对话服务系统，深度集成大语言模型（OpenAI/千问/Gemini 等）与本地 **RAG（检索增强生成）**技术，提供智能对话、知识库管理、文档解析等核心功能。

### 核心特性
- **微服务架构**: 基于 GoZero 框架，支持服务注册发现、负载均衡、链路追踪
- **多模型支持**: 兼容 OpenAI API 标准，支持千问、Gemini、GPT 等多种大模型
- **RAG 知识库**: 集成 PostgreSQL + pgvector，支持向量检索和语义搜索
- **文档处理**: 独立 MCP 微服务，支持 PDF 文档解析和文本提取
- **用户认证**: JWT + RefreshToken 双令牌机制（开发中）
- **会话状态机**: Redis 实现有限状态机，管理对话状态转移

### 技术栈
- **后端**: Go 1.24+ | GoZero | gRPC | PostgreSQL | Redis | ETCD
- **前端**: Vue 3 | Vite | TypeScript
- **AI**: OpenAI API（gemini-balance 轮询池） | 向量数据库 | UniPDF
- **部署**: Docker | Docker Compose

### 开发笔记
如果你想获取详细的开发笔记和技术分享，请访问我的**[个人博客专栏](https://aomanoh.com/categories/2)**查阅文档资料。

---

## 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         客户端层 (Client)                        │
│                     Vue3 + Vite + TypeScript                    │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTP/HTTPS
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                        API 网关层 (Gateway)                      │
│                  Nginx / GoZero Gateway (可选)                  │
└─────┬───────────────────────────────────┬───────────────────────┘
      │                                   │
      │ HTTP REST API                     │ HTTP REST API
      ▼                                   ▼
┌─────────────────────────┐   ┌─────────────────────────────────┐
│   Chat API 服务          │   │   User API 服务 (开发中)        │
│   端口: 8123             │   │   端口: 8124                    │
│   ─────────────────────  │   │   ─────────────────────────────  │
│   • AI 对话接口          │   │   • 用户注册/登录               │
│   • 知识库管理           │   │   • JWT 认证                    │
│   • 流式响应             │   │   • RefreshToken                │
│   • RAG 检索增强         │   │   • 会话管理                    │
└──────┬──────────────────┘   └─────────┬───────────────────────┘
       │ gRPC                            │
       │                                 │
       ▼                                 ▼
┌─────────────────────────┐   ┌─────────────────────────────────┐
│   MCP 微服务             │   │   ETCD 服务注册中心              │
│   端口: 8321             │   │   端口: 2379                    │
│   ─────────────────────  │   │   ─────────────────────────────  │
│   • PDF 文档解析         │   │   • 服务注册与发现               │
│   • 文本提取             │   │   • 配置管理                    │
│   • 高性能 gRPC          │   │   • 健康检查                    │
└─────────────────────────┘   └─────────────────────────────────┘
                             
┌─────────────────────────────────────────────────────────────────┐
│                        基础设施层 (Infrastructure)                │
├─────────────────────────┬──────────────────┬────────────────────┤
│  PostgreSQL + pgvector  │   Redis 6.x      │   OpenAI API       │
│  端口: 5432             │   端口: 6379     │   外部服务          │
│  ───────────────────    │   ─────────────  │   ───────────────  │
│  • 用户数据存储         │   • 会话状态机   │   • LLM 推理       │
│  • 向量数据存储         │   • RefreshToken │   • Embedding      │
│  • 知识库内容           │   • 异步任务队列 │   • 文本生成       │
└─────────────────────────┴──────────────────┴────────────────────┘
```

### 服务说明

| 服务名称 | 端口 | 协议 | 状态 | 说明 |
|---------|------|------|------|------|
| **Chat API** | 8123 | HTTP | 运行中 | AI 对话服务，支持流式输出和 RAG 检索 |
| **User API** | 8124 | HTTP | 开发中 | 用户认证服务，JWT + RefreshToken 机制 |
| **MCP Service** | 8321 | gRPC | 运行中 | PDF 文档处理微服务 |
| **ETCD** | 2379 | HTTP | 运行中 | 微服务注册发现中心 |
| **PostgreSQL** | 5432 | TCP | 运行中 | 主数据库 + 向量存储 |
| **Redis** | 6379 | TCP | 运行中 | 会话状态机与异步任务队列 |
| **Frontend** | 5173/80 | HTTP | 运行中 | Vue3 前端应用 |

### 数据流向

AI 对话流程：

```
用户输入 → Chat API → 检索向量数据库 (RAG) → 调用 OpenAI API → 流式返回 → 用户
                  ↓
         Redis (会话状态机)
```

文档上传流程：

```
用户上传 PDF → Chat API → MCP 服务 (gRPC) → 文本提取 → 向量化 → PostgreSQL
```

用户认证流程（开发中）：

```
登录请求 → User API → 验证密码 → 生成 JWT + RefreshToken → Redis (白名单) → 返回令牌
```

异步任务流程（规划中）：

```
定时器触发 → Redis 任务队列 → 清理匿名用户对话数据 → PostgreSQL
```

> **说明**: User 模块开发完成后，将支持异步清理匿名用户的对话历史，优化存储空间。

---

## 快速开始


### 配置文件说明：

- `xxx.yaml`：开发环境的配置文件
- `docker.yaml`：生产环境专用配置文件

| 文件名                             | 用途                | 说明                                                         |
| :--------------------------------- | ------------------- | :----------------------------------------------------------- |
| api/chat/etc/chat.yaml             | 本地开发环境配置    | 用于本地调试，数据库和服务地址指向 localhost 或 127.0.0.1    |
| api/chat/etc/docker.yaml           | Docker 容器环境配置 | 用于 Docker Compose 部署，服务地址使用 Docker 服务名（如 postgres、redis、mcp） |
| api/chat/etc/chat.example.yaml     | 本地开发配置模板    | 脱敏的示例文件，需复制为 chat.yaml 并填入真实密钥            |
| api/chat/etc/docker.example.yaml   | Docker 环境配置模板 | 脱敏的示例文件，需复制为 docker.yaml 并填入真实密钥          |
| mcp/etc/mcp.yaml                   | 本地开发环境配置    | ETCD 地址指向 127.0.0.1:2379                                 |
| mcp/etc/mcp.example.yaml           | 本地开发配置模板    | 脱敏的示例文件，需复制为 mcp.yaml 并填入 UniPDF 许可证       |
| mcp/etc/docker.example.yaml        | Docker 环境配置模板 | 脱敏的示例文件，需复制为 docker.yaml                         |
| api/user/etc/user-api.yaml         | 本地开发环境配置    | 包含 JWT 密钥和数据库连接信息                                |
| api/user/etc/user-api.example.yaml | 配置模板            | 脱敏的示例文件                                               |
| client/.env.development            | 开发环境变量        | API 地址通常为 `http://localhost:8123/api`                   |
| client/.env.production             | 生产环境变量        | API 地址通常为相对路径/api（通过 Nginx 代理）                |
| client/.env.example                | 环境变量模板        | 脱敏的示例文件，根据文件内容变更即可                         |
| docker-compose.yml                 | 容器编排配置        | 定义所有服务的 Docker                                        |
| docker-compose.example.yml         | 配置模板            | 脱敏的示例文件，需复制为 docker-compose.yml 并修改密码       |

---

### 第一步：配置文件准备

克隆项目后，需要根据 **example** 配置文件创建实际的配置文件，并填入你的密钥信息。

必需的配置文件（将项目中的配置文件中 **example** 关键字移除，然后更改秘钥、密码等配置）：

```bash
# 1. Chat API 配置（本地开发）
cp api/chat/etc/chat.example.yaml api/chat/etc/chat.yaml

# 2. Chat API 配置（Docker 环境）
cp api/chat/etc/docker.example.yaml api/chat/etc/docker.yaml

# 3. MCP 服务配置（本地开发）
cp mcp/etc/mcp.example.yaml mcp/etc/mcp.yaml

# 4. MCP 服务配置（Docker 环境）
cp mcp/etc/docker.example.yaml mcp/etc/docker.yaml

# 5. 前端环境变量
cp client/.env.example client/.env.development
cp client/.env.example client/.env.production

# 6. Docker Compose 配置
cp docker-compose.example.yml docker-compose.yml
```

关键配置项，修改以下配置文件，填入你的密钥信息：

**`api/chat/etc/chat.yaml`** 或 **`api/chat/etc/docker.yaml`**:

```yaml
OpenAI:
  ApiKey: "your-openai-api-key"        # 替换为你的 OpenAI API Key
  BaseURL: "https://api.openai.com/v1"
  Model: "gpt-4"

VectorDB:
  Password: "your-database-password"   # 替换为你的数据库密码
```

**`mcp/etc/mcp.yaml`** 或 **`mcp/etc/docker.yaml`**:
```yaml
# 请访问 https://unidoc.io 获取你自己的许可证
# 如果不知道如何获取，请访问我的个人博客查看详细步骤
UniPDFLicense: "your-unipdf-license"  # 替换为你的 UniPDF 许可证（可选）
```

**`docker-compose.yml`**:
```yaml
postgres:
  environment:
    POSTGRES_PASSWORD: your-database-password  # 替换为你的数据库密码
```

---

### 第二步：选择启动方式

方式一：完全 Docker 启动（适合快速体验），使用 Docker Compose 一键启动所有服务：

```bash
# 启动所有服务（ETCD、Redis、PostgreSQL、MCP、API）
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看服务日志，如
docker-compose logs -f api
docker-compose logs -f mcp
```

**说明**:

- 详细的 Docker 配置和故障排查请参考 Docker 官方文档或社区教程

---

#### 方式二：混合启动（Docker 基础服务 + 本地应用，推荐）

使用 Docker 运行 PostgreSQL 和 Redis，本地运行 Go 服务。这是笔者当前使用的方式。

**步骤 1**: 启动 Docker 基础服务，如果不会就用 **docker desktop** 软件启动。不过需要配置一下国内代理。我的博客有写明配置操作。如果不用那个，就可以参考下面的操作（不保证正确）

```bash
# 启动 PostgreSQL（带 pgvector 扩展）
docker run -d \
  --name postgres-vector \
  -e POSTGRES_DB=gozero_ai_agent \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=your_password \
  -p 5432:5432 \
  pgvector/pgvector:pg15

# 启动 Redis
docker run -d \
  --name my-redis \
  -p 6379:6379 \
  redis:alpine

# 初始化数据库（首次运行）
psql -h localhost -U postgres -d gozero_ai_agent -f db/init.sql
```

> 详细的 PostgreSQL + pgvector 和 Redis Docker 配置请查看**[博客文章](https://aomanoh.com/categories/2)**

**步骤 2**: 本地启动应用服务

```bash
# 终端 0：启动 ETCD 服务，这是启动微服务的前提（需要你先安装ETCD服务）
etcd

# 终端 1：启动 MCP 微服务
cd mcp
go run mcp.go -f etc/mcp.yaml

# 终端 2：启动 Chat API 服务
cd api/chat
go run chat.go -f etc/chat.yaml

# 终端 3：启动前端（可选）
cd client
npm install  # 首次运行
npm run dev
```

**访问地址**:

- Chat API: http://localhost:8123
- 前端: http://localhost:5000

---

#### 方式三：完全本地启动（适合深度调试）

> 由于笔者没有下载过 PostgreSQL，所以笔者这里的说明只供参考。

需要本地安装 Redis 和 PostgreSQL（含 pgvector 扩展）。

**步骤 1**: 启动本地服务和 Docker ETCD

```bash
# 启动 Redis（根据你的 Redis 版本选择启动方式）
# 微软维护的旧版本（3.x）: redis-server.exe redis.windows.conf
# 高版本（3.x+）: redis-server.exe redis.conf

# 启动 PostgreSQL（需预先安装 pgvector）
# 创建数据库并初始化
psql -U postgres -c "CREATE DATABASE gozero_ai_agent;"
psql -U postgres -d gozero_ai_agent -c "CREATE EXTENSION vector;"
psql -U postgres -d gozero_ai_agent -f db/init.sql
```

**步骤 2**: 确保配置文件指向本地服务（127.0.0.1）

**步骤 3**: 启动应用服务（同方式二）

---

### 启动方式对比

| 启动方式 | 优点 | 缺点 | 适用场景 |
|---------|------|------|----------|
| **完全 Docker** | 环境一致，一键启动 | 调试困难，资源占用高 | 快速体验、演示 |
| **混合启动** | 易于调试，环境隔离 | 需要管理多个终端 | 日常开发（推荐）|
| **完全本地** | 调试方便，资源占用少 | 环境配置复杂 | 深度调试 |

> 作者使用混合模式，因为 PostgreSQL 和 Redis 配置稳定，使用 Docker 更方便。

---

### 服务启动顺序

无论选择哪种方式，都应按以下顺序启动：

1. **ETCD** (端口 2379) - 服务注册中心（必需，MCP 服务依赖）
2. **Redis** (端口 6379) - 会话状态机
3. **PostgreSQL** (端口 5432) - 数据库 + 向量存储
4. **MCP 微服务** (端口 8321) - PDF 处理（依赖 ETCD）
5. **Chat API** (端口 8123) - 主服务
6. **前端** (端口 5173) - 可选

**重要提示**: ETCD 是 GoZero 微服务的服务注册中心，MCP 服务启动时需要向 ETCD 注册，因此 ETCD 必须先于 MCP 服务启动。如果没有本地 ETCD，建议使用 Docker 运行。

---

## Docker-Compose 部署

完整服务部署：

```bash
# 启动完整服务栈
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f [service_name]
```

精简前端部署

```bash
# 仅启动前端服务
docker-compose -f docker-compose.client.slim.yml up -d
```

---

## 项目更新时重新部署注意事项

### 前端热更新

+ 删除服务器中的 dist 目录
+ 本地执行 `npm run build` 生成新的 dist，然后压缩文件
+ 在本地终端里上传新的 dist.zip 到服务器中并解压，如：
    + `scp 你想上传的本地文件路径 你的服务器用户名@你的服务器地址:你想上传的服务器文件指定位置`
        + 如：`scp ./dist.zip "root@8.137.22.79:/opt/项目根目录/client"`
    + 如果输入密码后提示找不到文件，那就两步走
        + 先本地执行执行：`scp ./dist.zip root@120.79.123.45:~`
        + 上传成功后在服务器终端里执行：`mv ~/dist.zip /opt/项目根目录/client/`
    + 解压文件：`unzip dist.zip`
+ 重启前端容器：`docker restart xxx-client`
+ 注意：如果出现 **520/521** 的报错，可能是因为修改部署了后端后，重新挂载前端时，前端无法找到后端的 `compose` 文件导致。此时你可以执行：
    + `docker compose -f docker-compose.yml -f docker-compose.client.yml up client`


---

### 后端更新

+ 本地修改代码并提交到 github 仓库中
+ 服务器执行 `docker-compose pull`，或者本地打包上传服务器后，替换掉旧文件。
+ 服务器执行 `docker-compose build api（你的 docker-compose SERVICE）`，重新生成镜像文件。
+ 服务器执行 `docker-compose up -d --no-deps api（你的 docker-compose SERVICE）`，重启并替换旧的服务容器，其他容器不会中断。

+ 注意，不能删除容器，否则数据会丢失

---

## 常见问题排查

### 服务无法启动
1. 检查端口是否被占用: `netstat -tulpn | grep [端口号]`
2. 检查配置文件路径是否正确
3. 验证依赖服务（Redis、PostgreSQL）是否正常运行

---

### 数据库连接失败

1. 确认 PostgreSQL 服务已启动
2. 检查数据库连接参数
3. 验证 pgvector 扩展是否已安装

---

### API 调用失败
1. 检查 OpenAI API Key 配置
2. 验证网络连接
3. 确认请求格式是否正确

---

### 前端无法访问后端
查看日志 `docker-compose -f docker-compose.yml -f docker-compose.client.slim.yml logs client` 定位并分析问题

---

## 注意事项

- 确保你的 key 足额。或者用自己搭建 gemini 轮询池服务
- 更多详情请前往我的博客中查询有关信息