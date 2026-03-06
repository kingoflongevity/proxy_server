# 构建后端
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend .
RUN CGO_ENABLED=1 GOOS=linux go build -o proxy-server .

# 构建前端
FROM node:18-alpine AS frontend-builder

WORKDIR /app
COPY frontend/package*.json ./
RUN npm install

COPY frontend .
RUN npm run build

# 运行镜像
FROM alpine:latest

RUN apk add --no-cache ca-certificates sqlite

WORKDIR /app

# 复制后端
COPY --from=backend-builder /app/proxy-server .

# 复制前端静态文件
COPY --from=frontend-builder /app/dist ./static

# 创建数据目录
RUN mkdir -p /app/data

# 暴露端口
EXPOSE 8000 10808 10809 10810

# 设置环境变量
ENV GIN_MODE=release

# 启动服务
CMD ["./proxy-server"]
