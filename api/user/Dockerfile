# ==========================================
# GoZero-AI User API Docker镜像构建文件
# ==========================================
# 功能: 构建用户认证与工作台接口服务容器镜像
# 架构: 多阶段构建，分离编译环境和运行环境

FROM golang:1.24.6 AS builder

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /build
COPY . .

WORKDIR /build/api/user
RUN CGO_ENABLED=0 GOOS=linux go build -o /output/user .

FROM alpine:latest

RUN mkdir -p /api/user/etc

COPY --from=builder /build/api/user/etc/docker.yaml /api/user/etc/user-api.yaml
COPY --from=builder /output/user /api/user/user

WORKDIR /api/user
CMD ["./user", "-f", "etc/user-api.yaml"]
