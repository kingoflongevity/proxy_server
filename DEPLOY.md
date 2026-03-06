# Proxy Server - 部署指南

## 环境要求

- Go 1.21+
- Node.js 18+
- SQLite3

## 快速部署

### 1. 编译后端

```bash
cd backend
go build -o proxy-server .
```

### 2. 构建前端

```bash
cd frontend
npm install
npm run build
```

### 3. 运行

```bash
# 直接运行
./proxy-server

# 或指定端口
PORT=8080 ./proxy-server
```

## Linux Systemd 服务

创建服务文件 `/etc/systemd/system/proxy-server.service`:

```ini
[Unit]
Description=Proxy Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/proxy-server
ExecStart=/opt/proxy-server/proxy-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启用并启动服务:

```bash
sudo systemctl daemon-reload
sudo systemctl enable proxy-server
sudo systemctl start proxy-server
```

## Docker 部署

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY backend .
RUN go build -o proxy-server .

FROM node:18-alpine AS frontend-builder

WORKDIR /app
COPY frontend .
RUN npm install && npm run build

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/proxy-server .
COPY --from=frontend-builder /app/dist ./static

EXPOSE 8000 10808 10809 10810

CMD ["./proxy-server"]
```

### 构建和运行

```bash
docker build -t proxy-server .
docker run -d -p 8000:8000 -p 10808:10808 -p 10809:10809 -p 10810:10810 proxy-server
```

## 配置说明

### 端口说明

| 端口 | 用途 |
|------|------|
| 8000 | Web 管理界面 API |
| 10808 | SOCKS5 代理 |
| 10809 | HTTP 代理 |
| 10810 | 混合代理 (HTTP+SOCKS5) |

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| PORT | API 服务端口 | 8000 |
| DB_PATH | 数据库路径 | ./data/proxy.db |

## 内核安装

首次使用需要安装 Xray 内核：

1. **自动更新**: 在设置页面点击"从官方更新"
2. **手动上传**: 下载对应平台的 Xray 内核，在设置页面上传

内核下载地址: https://github.com/XTLS/Xray-core/releases

## 代理使用

### 浏览器代理设置

设置浏览器代理为：
- HTTP: 127.0.0.1:10809
- SOCKS5: 127.0.0.1:10808
- 混合: 127.0.0.1:10810

### 系统代理设置

Windows:
```
netsh winhttp set proxy 127.0.0.1:10809
```

Linux:
```bash
export http_proxy=http://127.0.0.1:10809
export https_proxy=http://127.0.0.1:10809
export all_proxy=socks5://127.0.0.1:10808
```

## 安全建议

1. 修改默认端口
2. 启用防火墙，仅开放必要端口
3. 定期更新内核版本
4. 不要在公网暴露管理界面

## 故障排除

### 查看日志

```bash
# Systemd 服务日志
journalctl -u proxy-server -f

# Docker 日志
docker logs -f <container_id>
```

### 常见问题

1. **内核无法启动**: 检查内核文件是否有执行权限
2. **端口被占用**: 使用 `netstat -tlnp | grep <port>` 检查端口占用
3. **连接失败**: 检查节点配置和代理模式设置
