# GoZero-AI 快速启动指南

## 项目概述

GoZero-AI 是一个基于 GoZero 框架构建的 AI 对话服务系统，结合了大模型（如 OpenAI）与本地 RAG（Retrieval-Augmented Generation）能力，提供智能对话、知识上传与处理等功能。

## 系统架构

- **API 服务**: 主要的 AI 对话接口服务
- **MCP 微服务**: PDF 文档处理微服务
- **Redis**: 会话状态管理和缓存
- **PostgreSQL**: 向量数据库存储
- **Vue3 前端**: 用户交互界面

## 本地开发环境启动方式对比

### 方式一：完全 Docker 启动（推荐）

适用于：快速环境搭建，避免本地环境冲突

#### 1. 启动基础服务
```bash
# 启动 Redis 服务
docker run -d --name my-redis -p 127.0.0.1:6379:6379 redis

# 启动 PostgreSQL 数据库（带 pgvector 扩展）
docker run -d \
  --name postgres-vector \
  -e POSTGRES_DB=gozero_ai \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=your_password \
  -p 5432:5432 \
  pgvector/pgvector:pg15

# 初始化数据库结构
psql -h localhost -U postgres -d gozero_ai -f db/init.sql
```

#### 2. 启动应用服务
```bash
# 构建并启动所有服务
docker-compose up -d

# 或者分别启动
docker-compose up -d postgres redis
docker-compose up -d mcp-service
docker-compose up -d api-service
docker-compose up -d frontend
```

#### 3. 验证服务状态
```bash
# 查看所有服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f api-service
docker-compose logs -f mcp-service
```

---

### 方式二：混合启动（Docker基础服务 + 本地应用）

适用于：需要调试Go代码或前端代码

#### 1. 启动基础服务（Docker）
```bash
# 启动 Redis 和 PostgreSQL
docker run -d --name my-redis -p 127.0.0.1:6379:6379 redis

docker run -d \
  --name postgres-vector \
  -e POSTGRES_DB=gozero_ai \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=123456 \
  -p 5432:5432 \
  pgvector/pgvector:pg15

# 初始化数据库
psql -h localhost -U postgres -d gozero_ai -f db/init.sql
```

#### 2. 启动应用服务（本地）
```bash
# 启动 MCP 微服务
cd mcp
go run mcp.go -f etc/mcp.yaml
# 或构建后启动: go build -o mcp-server mcp.go && ./mcp-server -f etc/mcp.yaml

# 新终端：启动 API 主服务
cd api
go run chat.go -f etc/chat.yaml
# 或构建后启动: go build -o chat-server chat.go && ./chat-server -f etc/chat.yaml

# 新终端：启动前端开发服务器
cd client
npm install  # 首次运行
npm run dev
```

---

### 方式三：完全本地启动

适用于：有本地Redis和PostgreSQL环境，或需要深度调试

#### 1. 启动本地数据库服务
```bash
# 启动本地 Redis（需要预先安装）
redis-server /usr/local/etc/redis.conf
# Windows: redis-server.exe redis.windows.conf
# 如果你是用的最新的Windows-redis服务，那么请执行：redis-server.exe redis.conf

# 启动本地 PostgreSQL（需要预先安装并配置pgvector）
psql -U postgres -c "CREATE DATABASE gozero_ai;"
psql -U postgres -d gozero_ai -c "CREATE EXTENSION vector;"
psql -U postgres -d gozero_ai -f db/init.sql
```

#### 2. 修改配置文件
```bash
# 确保配置文件指向本地服务
# api/etc/chat.yaml 和 mcp/etc/mcp.yaml 中的数据库连接配置
```

#### 3. 启动应用服务
```bash
# 启动 MCP 微服务
cd mcp
go run mcp.go -f etc/mcp.yaml

# 启动 API 主服务
cd api  
go run chat.go -f etc/chat.yaml

# 启动前端服务
cd client
npm run dev
```

---

## 启动方式选择指南

| 启动方式       | 优点                           | 缺点                       | 适用场景             |
| -------------- | ------------------------------ | -------------------------- | -------------------- |
| **完全Docker** | 环境一致，快速部署，无本地依赖 | 调试相对困难，资源占用较高 | 演示、测试、生产环境 |
| **混合启动**   | 易于调试Go代码，环境相对隔离   | 需要管理多个进程           | 后端开发、API调试    |
| **完全本地**   | 调试最方便，资源占用最少       | 环境配置复杂，版本冲突风险 | 深度开发、性能调优   |

## 服务启动顺序

**重要**: 无论选择哪种启动方式，都请按照以下顺序启动服务，确保依赖关系正确：

### 启动顺序说明
1. **Redis** (端口: 6379) - 会话状态缓存
2. **PostgreSQL** (端口: 5432) - 向量数据库存储
3. **MCP 微服务** (端口: 8082) - PDF处理服务
4. **API 主服务** (端口: 8888) - 核心业务逻辑
5. **前端服务** (端口: 5173/80) - 用户界面

### 依赖关系检查
```bash
# 检查服务启动顺序
echo "检查 Redis..." && redis-cli ping
echo "检查 PostgreSQL..." && pg_isready -h localhost -p 5432
echo "检查 MCP 服务..." && curl -f http://localhost:8082/health || echo "MCP未启动"
echo "检查 API 服务..." && curl -f http://localhost:8888/health || echo "API未启动"
echo "检查前端服务..." && curl -f http://localhost:5173 || echo "前端未启动"
```

## 快速启动脚本

为了方便开发者快速启动服务，提供以下脚本：

### Docker 启动脚本
创建 `start-docker.sh`（Linux/Mac）或 `start-docker.bat`（Windows）：

```bash
#!/bin/bash
# start-docker.sh
echo "启动 GoZero-AI Docker 环境..."

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo "错误: Docker 未运行，请先启动 Docker"
    exit 1
fi

# 清理旧容器（可选）
echo "清理旧容器..."
docker-compose down

# 启动所有服务
echo "启动所有服务..."
docker-compose up -d

# 等待服务启动
echo "等待服务启动..."
sleep 10

# 初始化数据库
echo "初始化数据库..."
docker-compose exec postgres psql -U postgres -d gozero_ai -f /docker-entrypoint-initdb.d/init.sql

echo "服务启动完成！"
echo "访问地址: http://localhost:80"
docker-compose ps
```

### 本地开发脚本
创建 `start-local.sh`（Linux/Mac）或 `start-local.bat`（Windows）：

```bash
#!/bin/bash
# start-local.sh
echo "启动 GoZero-AI 本地开发环境..."

# 检查依赖
check_dependency() {
    if ! command -v $1 &> /dev/null; then
        echo "错误: $1 未安装或不在 PATH 中"
        exit 1
    fi
}

check_dependency "go"
check_dependency "node"
check_dependency "docker"

# 启动基础服务（Docker）
echo "启动基础服务 (Redis & PostgreSQL)..."
docker run -d --name my-redis -p 127.0.0.1:6379:6379 redis 2>/dev/null || echo "Redis 容器已存在"
docker run -d --name postgres-vector \
  -e POSTGRES_DB=gozero_ai \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=123456 \
  -p 5432:5432 \
  pgvector/pgvector:pg15 2>/dev/null || echo "PostgreSQL 容器已存在"

# 等待数据库启动
echo "等待数据库启动..."
sleep 5

# 初始化数据库
echo "初始化数据库..."
psql -h localhost -U postgres -d gozero_ai -f db/init.sql 2>/dev/null || echo "数据库已初始化"

echo "基础服务启动完成！"
echo "现在请手动在不同终端中启动："
echo "1. cd mcp && go run mcp.go -f etc/mcp.yaml"
echo "2. cd api && go run chat.go -f etc/chat.yaml"
echo "3. cd client && npm run dev"
```

### Windows 批处理脚本
创建 `start-docker.bat`：

```batch
@echo off
echo 启动 GoZero-AI Docker 环境...

REM 检查 Docker 是否运行
docker info >nul 2>&1
if errorlevel 1 (
    echo 错误: Docker 未运行，请先启动 Docker
    pause
    exit /b 1
)

REM 清理旧容器
echo 清理旧容器...
docker-compose down

REM 启动服务
echo 启动所有服务...
docker-compose up -d

REM 等待服务启动
echo 等待服务启动...
timeout /t 10 /nobreak

echo 服务启动完成！
echo 访问地址: http://localhost:80
docker-compose ps
pause
```

## 服务验证

### 通用验证脚本
创建 `check-services.sh` 或 `check-services.bat` 来验证所有服务状态：

```bash
#!/bin/bash
# check-services.sh
echo "==== GoZero-AI 服务状态检查 ===="

# 检查 Redis
echo -n "Redis (6379): "
if redis-cli ping >/dev/null 2>&1; then
    echo "✓ 运行中"
else
    echo "✗ 未运行"
fi

# 检查 PostgreSQL
echo -n "PostgreSQL (5432): "
if pg_isready -h localhost -p 5432 >/dev/null 2>&1; then
    echo "✓ 运行中"
else
    echo "✗ 未运行"
fi

# 检查 MCP 服务
echo -n "MCP 服务 (8082): "
if curl -f http://localhost:8082/health >/dev/null 2>&1; then
    echo "✓ 运行中"
else
    echo "✗ 未运行"
fi

# 检查 API 服务
echo -n "API 服务 (8888): "
if curl -f http://localhost:8888/health >/dev/null 2>&1; then
    echo "✓ 运行中"
else
    echo "✗ 未运行"
fi

# 检查前端服务
echo -n "前端服务 (5173/80): "
if curl -f http://localhost:5173 >/dev/null 2>&1 || curl -f http://localhost:80 >/dev/null 2>&1; then
    echo "✓ 运行中"
else
    echo "✗ 未运行"
fi

echo "==== 检查完成 ===="
```

### API 服务验证
```bash
# 检查 API 服务健康状态
curl http://localhost:8888/health

# 测试聊天接口
curl -X POST http://localhost:8888/chat \
  -H "Content-Type: application/json" \
  -d '{"chatId":"test123", "message":"你好"}'
```

### MCP 服务验证
```bash
# 检查 MCP 服务状态
grpc_health_probe -addr=localhost:8082
```

### 数据库验证
```bash
# 验证 PostgreSQL 连接
psql -h localhost -U postgres -d gozero_ai -c "SELECT version();"

# 验证 pgvector 扩展
psql -h localhost -U postgres -d gozero_ai -c "SELECT * FROM pg_extension WHERE extname='vector';"
```

## Docker Compose 部署

### 完整服务部署
```bash
# 启动完整服务栈
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f [service_name]
```

### 精简前端部署
```bash
# 仅启动前端服务
docker-compose -f docker-compose.client.slim.yml up -d
```

## 配置文件说明

### API 服务配置
- **开发环境**: `api/etc/chat.yaml`
- **Docker 环境**: `api/etc/docker.yaml`

### MCP 服务配置
- **开发环境**: `mcp/etc/mcp.yaml`
- **Docker 环境**: `mcp/etc/docker.yaml`

## 端口映射

| 服务       | 开发端口 | Docker 端口 | 说明             |
| ---------- | -------- | ----------- | ---------------- |
| API 服务   | 8888     | 8888        | 主要 AI 对话接口 |
| MCP 服务   | 8082     | 8082        | PDF 处理微服务   |
| Redis      | 6379     | 6379        | 状态管理缓存     |
| PostgreSQL | 5432     | 5432        | 向量数据库       |
| 前端服务   | 5173     | 80          | Vue3 用户界面    |

## 环境变量配置

### OpenAI API 配置
```bash
export OPENAI_API_KEY="your-openai-api-key"
export OPENAI_BASE_URL="https://api.openai.com/v1"
```

### 数据库配置
```bash
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_NAME="gozero_ai"
export DB_USER="postgres"
export DB_PASSWORD="123456"
```

## 常见问题排查

### 服务无法启动
1. 检查端口是否被占用: `netstat -tulpn | grep [端口号]`
2. 检查配置文件路径是否正确
3. 验证依赖服务（Redis、PostgreSQL）是否正常运行

### 数据库连接失败
1. 确认 PostgreSQL 服务已启动
2. 检查数据库连接参数
3. 验证 pgvector 扩展是否已安装

### API 调用失败
1. 检查 OpenAI API Key 配置
2. 验证网络连接
3. 确认请求格式是否正确

## 开发工具推荐

- **Go 版本**: 1.24.6
- **Node.js**: 最新 LTS 版本
- **Redis 客户端**: redis-cli 或 RedisInsight
- **数据库客户端**: pgAdmin 或 DBeaver
- **API 测试**: Postman 或 curl

## 性能优化建议

1. **生产环境部署时调整连接池大小**
2. **配置 Redis 持久化策略**
3. **优化 PostgreSQL 向量索引**
4. **启用 Gzip 压缩**
5. **配置反向代理（Nginx）**

## 注意事项

- 确保 OpenAI API Key 有足够的配额
- 生产环境请修改默认密码
- 定期备份数据库数据
- 监控服务资源使用情况
- 及时更新依赖库版本