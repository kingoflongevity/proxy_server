# 阶段1: 构建前端
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm install

COPY frontend/ ./
RUN npm run build

# 阶段2: 运行后端
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o proxy-server .

# 最终镜像
FROM alpine:3.19

RUN apk add --no-cache ca-certificates xray

WORKDIR /app

COPY --from=backend-builder /app/backend/proxy-server ./
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

RUN mkdir -p /app/data /app/logs

EXPOSE 3000 8000

ENV PORT=8000
ENV FRONTEND_PORT=3000

ENTRYPOINT ["./proxy-server"]
