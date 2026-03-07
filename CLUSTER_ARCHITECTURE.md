# 代理集群管理系统 - 完整架构设计文档

## 目录
1. [系统概述](#系统概述)
2. [总体架构](#总体架构)
3. [数据库设计](#数据库设计)
4. [后端服务设计](#后端服务设计)
5. [API接口设计](#api接口设计)
6. [前端架构设计](#前端架构设计)
7. [备份恢复机制](#备份恢复机制)
8. [自动伸缩策略](#自动伸缩策略)
9. [部署方案](#部署方案)
10. [安全设计](#安全设计)

---

## 系统概述

### 1.1 系统定位

| 维度 | 说明 |
|------|------|
| **项目类型** | 分布式代理集群管理系统 |
| **访问模型** | 内部管理系统（局域网/私有网络） |
| **规模预估** | 中型（支持 10-100 台服务器集群） |
| **核心能力** | 自动化部署、集中管理、动态伸缩、实时监控 |

### 1.2 技术栈

| 层级 | 技术选型 |
|------|---------|
| **前端** | Vue 3 + TypeScript + Vite + Canvas |
| **后端** | Go 1.25+ + Gin + SQLite |
| **代理核心** | Xray-core v26.2.6+ |
| **远程连接** | SSH (golang.org/x/crypto/ssh) |
| **通信协议** | HTTP/REST + WebSocket (实时通信) |

---

## 总体架构

### 2.1 架构分层

```
┌─────────────────────────────────────────────────────────────────┐
│                        Client Layer                              │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Vue 3 Frontend                                          │   │
│  │  - 服务器列表视图                                          │   │
│  │  - 拓扑图可视化 (Canvas)                                   │   │
│  │  - 实时监控面板                                            │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Gateway Layer                              │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Gin Router + Middleware                                  │   │
│  │  - Auth / RateLimit / Logger / Recovery                   │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Service Layer                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌───────────┐ │
│  │   Scanner   │ │   Deployer  │ │   Scaler    │ │  Backup   │ │
│  │   Service   │ │   Service   │ │   Service   │ │  Service  │ │
│  │ 局域网扫描   │ │ 代理部署     │ │  自动伸缩    │ │  备份恢复  │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └───────────┘ │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐              │
│  │    SSH      │ │   Monitor   │ │   Cluster   │              │
│  │   Manager   │ │   Service   │ │   Service   │              │
│  │ 远程连接管理  │ │  监控服务    │ │  集群管理    │              │
│  └─────────────┘ └─────────────┘ └─────────────┘              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Data Layer                                │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐              │
│  │   SQLite    │ │    Redis    │ │  File Store │              │
│  │  (主数据)   │ │   (缓存)    │ │  (备份文件) │              │
│  └─────────────┘ └─────────────┘ └─────────────┘              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Infrastructure Layer                          │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐              │
│  │  Xray-core  │ │   SSH/TCP   │ │   WebSocket │              │
│  │  (代理内核) │ │  (远程连接) │ │  (实时通信) │              │
│  └─────────────┘ └─────────────┘ └─────────────┘              │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 核心设计原则

| 原则 | 说明 |
|------|------|
| **无状态化** | 服务层无状态，支持水平扩展 |
| **异步处理** | 扫描、部署等耗时任务异步执行 |
| **连接池** | SSH连接复用，减少连接开销 |
| **幂等性** | 所有API设计保证幂等 |
| **失败重试** | 网络操作自动重试机制 |

---

## 数据库设计

### 3.1 核心数据表

#### cluster_servers (集群服务器表)
```sql
CREATE TABLE cluster_servers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    ip TEXT NOT NULL UNIQUE,
    port INTEGER DEFAULT 22,
    username TEXT NOT NULL,
    password TEXT,                    -- AES加密存储
    private_key TEXT,                 -- AES加密存储
    os_type TEXT,
    os_version TEXT,
    arch TEXT,
    status TEXT DEFAULT 'offline',
    last_heartbeat DATETIME,
    
    proxy_enabled INTEGER DEFAULT 0,
    proxy_port INTEGER,
    proxy_type TEXT,
    proxy_config TEXT,                -- JSON存储
    
    cpu INTEGER,
    memory INTEGER,
    disk INTEGER,
    cpu_usage REAL,
    memory_usage REAL,
    bandwidth_up INTEGER,
    bandwidth_down INTEGER,
    connections INTEGER,
    
    tags TEXT,                        -- JSON数组
    group_id TEXT,
    
    created_at DATETIME,
    updated_at DATETIME,
    deployed_at DATETIME
);
```

#### server_groups (服务器分组表)
```sql
CREATE TABLE server_groups (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    auto_scale INTEGER DEFAULT 0,
    min_servers INTEGER DEFAULT 1,
    max_servers INTEGER DEFAULT 10,
    scale_policy TEXT DEFAULT 'cpu',
    scale_threshold REAL DEFAULT 80.0,
    created_at DATETIME,
    updated_at DATETIME
);
```

#### backup_records (备份记录表)
```sql
CREATE TABLE backup_records (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    server_id TEXT,
    file_path TEXT NOT NULL,
    file_size INTEGER,
    checksum TEXT,
    status TEXT DEFAULT 'creating',
    error_message TEXT,
    created_at DATETIME
);
```

#### scan_tasks (扫描任务表)
```sql
CREATE TABLE scan_tasks (
    id TEXT PRIMARY KEY,
    status TEXT DEFAULT 'pending',
    network_cidr TEXT NOT NULL,
    progress INTEGER DEFAULT 0,
    results TEXT,                     -- JSON存储
    started_at DATETIME,
    completed_at DATETIME,
    created_at DATETIME
);
```

#### deploy_tasks (部署任务表)
```sql
CREATE TABLE deploy_tasks (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    progress INTEGER DEFAULT 0,
    current_step TEXT,
    logs TEXT,                        -- JSON存储
    started_at DATETIME,
    completed_at DATETIME,
    created_at DATETIME
);
```

#### scale_events (伸缩事件表)
```sql
CREATE TABLE scale_events (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    type TEXT NOT NULL,
    reason TEXT,
    target_count INTEGER,
    status TEXT DEFAULT 'pending',
    created_at DATETIME,
    completed_at DATETIME
);
```

### 3.2 数据模型关系

```
server_groups (1) ────── (N) cluster_servers
                              │
                              ├─ (1) backup_records
                              ├─ (1) deploy_tasks
                              └─ (N) scale_events (通过group_id)
```

---

## 后端服务设计

### 4.1 SSH连接管理服务 (SSHManager)

**职责**: 管理与远程服务器的SSH连接，提供远程命令执行和文件传输能力

**核心功能**:
- 连接管理: 建立、断开、复用SSH连接
- 远程执行: 执行Shell命令，支持超时控制
- 文件传输: 上传/下载文件，支持SCP协议
- 系统探测: 自动检测操作系统类型和版本

**关键设计**:
```go
type SSHManager interface {
    Connect(server *ClusterServer) (*ssh.Client, error)
    Disconnect(serverID string) error
    Execute(serverID, cmd string) (stdout, stderr string, err error)
    UploadFile(serverID, remotePath string, content []byte) error
    DownloadFile(serverID, remotePath string) ([]byte, error)
    GetSystemInfo(serverID string) (*SystemInfo, error)
    DetectOS(serverID string) (osType, version string, err error)
}
```

**安全措施**:
- 密码和私钥AES加密存储
- 支持公钥认证（推荐）
- 连接超时控制
- 命令执行超时控制

### 4.2 局域网扫描服务 (ScannerService)

**职责**: 扫描局域网内的Linux服务器，发现可用主机

**核心功能**:
- CIDR扫描: 支持标准CIDR格式（如 192.168.1.0/24）
- 并发扫描: 多协程并发，提高扫描效率
- 端口检测: 检测SSH端口（默认22）可达性
- 系统识别: 通过SSH Banner识别操作系统类型
- 延迟测试: 测试网络延迟

**扫描流程**:
```
1. 解析CIDR → 生成IP列表
2. 创建扫描任务 → 记录到数据库
3. 启动Worker Pool → 并发扫描
4. TCP连接测试 → 判断可达性
5. 获取SSH Banner → 识别系统
6. 汇总结果 → 更新任务状态
```

**性能优化**:
- Worker Pool限制并发数（默认50）
- 连接超时控制（默认5秒）
- 扫描进度实时更新
- 支持取消扫描任务

### 4.3 代理部署服务 (DeployerService)

**职责**: 自动在远程服务器上部署Xray代理服务

**部署步骤**:
```
1. connect     - 建立SSH连接
2. check_env   - 检查系统环境
3. create_dirs - 创建服务目录
4. upload_binary - 上传Xray二进制
5. upload_config - 上传配置文件
6. install_service - 安装系统服务
7. start_service - 启动代理服务
8. verify      - 验证部署结果
```

**支持的系统**:
- Ubuntu/Debian (apt-get)
- CentOS/RHEL (yum)
- Alpine (apk)

**服务配置**:
```ini
[Unit]
Description=Xray Proxy Service
After=network.target

[Service]
Type=simple
User=root
ExecStart=/opt/proxy/bin/xray run -c /opt/proxy/config/config.json
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

**批量操作**:
- 批量部署: 多服务器并发部署
- 批量启动: 一键启动所有代理
- 批量停止: 一键停止所有代理

### 4.4 备份恢复服务 (BackupService)

**职责**: 提供配置备份和恢复功能，确保数据安全

**备份类型**:
- **全量备份**: 包含所有服务器配置、分组、系统设置
- **配置备份**: 仅备份服务器配置信息
- **代理配置备份**: 备份远程服务器的代理配置文件

**备份格式**:
```json
{
  "version": "1.0",
  "created_at": "2026-03-07T10:00:00Z",
  "servers": [...],
  "groups": [...],
  "settings": {...}
}
```

**备份流程**:
```
1. 收集数据 → 从数据库读取
2. 序列化 → JSON格式
3. 压缩 → GZIP压缩
4. 打包 → TAR归档
5. 计算校验和 → MD5
6. 存储 → 本地文件系统
```

**恢复流程**:
```
1. 验证备份 → 检查文件完整性
2. 解压解包 → 提取数据
3. 反序列化 → 解析JSON
4. 数据恢复 → 写入数据库
5. 服务重启 → 应用新配置
```

**定时备份**:
- 可配置备份间隔
- 自动清理过期备份
- 备份失败告警

### 4.5 自动伸缩服务 (ScalerService)

**职责**: 根据负载自动增减服务器数量

**伸缩策略**:
- **CPU策略**: 基于平均CPU使用率
- **内存策略**: 基于平均内存使用率
- **连接数策略**: 基于平均连接数

**扩容条件**:
```
if (metric > threshold && serverCount < maxServers) {
    scaleUp(1);
}
```

**缩容条件**:
```
if (metric < threshold/2 && serverCount > minServers) {
    scaleDown(1);
}
```

**服务器选择**:
- 扩容: 选择未分配的在线服务器
- 缩容: 选择负载最低的服务器

**伸缩事件**:
```
{
  "id": "scale-001",
  "group_id": "group-001",
  "type": "scale_up",
  "reason": "CPU使用率超过85%",
  "target_count": 5,
  "status": "completed"
}
```

---

## API接口设计

### 5.1 服务器管理API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/cluster/servers` | 获取服务器列表 |
| POST | `/api/cluster/servers` | 创建服务器 |
| GET | `/api/cluster/servers/:id` | 获取服务器详情 |
| PUT | `/api/cluster/servers/:id` | 更新服务器 |
| DELETE | `/api/cluster/servers/:id` | 删除服务器 |
| POST | `/api/cluster/servers/:id/test` | 测试连接 |
| POST | `/api/cluster/servers/:id/start` | 启动代理 |
| POST | `/api/cluster/servers/:id/stop` | 停止代理 |
| POST | `/api/cluster/servers/:id/restart` | 重启代理 |
| GET | `/api/cluster/servers/:id/status` | 获取代理状态 |

### 5.2 扫描管理API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/cluster/scan` | 启动扫描 |
| GET | `/api/cluster/scan/:id` | 获取扫描任务 |
| POST | `/api/cluster/scan/:id/cancel` | 取消扫描 |
| POST | `/api/cluster/scan/quick` | 快速扫描 |

### 5.3 部署管理API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/cluster/deploy` | 部署代理 |
| GET | `/api/cluster/deploy/:id` | 获取部署任务 |
| POST | `/api/cluster/deploy/batch` | 批量部署 |

### 5.4 备份管理API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/cluster/backups` | 创建备份 |
| GET | `/api/cluster/backups` | 获取备份列表 |
| POST | `/api/cluster/backups/restore` | 恢复备份 |
| DELETE | `/api/cluster/backups/:id` | 删除备份 |

### 5.5 伸缩管理API

| 方法 | 路径 | 说明 |
|------|------|------|
| PUT | `/api/cluster/scale/policy` | 更新伸缩策略 |
| POST | `/api/cluster/scale/up` | 扩容 |
| POST | `/api/cluster/scale/down` | 缩容 |
| GET | `/api/cluster/scale/events` | 获取伸缩事件 |

### 5.6 拓扑图API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/cluster/topology` | 获取集群拓扑 |

---

## 前端架构设计

### 6.1 页面结构

```
Cluster.vue (集群管理主页面)
├── 服务器列表视图
│   ├── 服务器卡片
│   ├── 资源使用监控
│   └── 操作按钮
├── 拓扑图视图
│   ├── Canvas渲染
│   ├── 交互操作
│   └── 节点详情
└── 分组管理视图
    ├── 分组列表
    └── 伸缩策略配置
```

### 6.2 拓扑图组件 (TopologyGraph)

**技术实现**:
- 使用HTML5 Canvas绘制
- 支持缩放、拖拽、选择
- 实时更新节点状态
- 动画效果（连接指示器）

**节点类型**:
- **主节点**: 管理节点，蓝色星形图标
- **服务器节点**: 代理节点，矩形图标
- **状态指示**: 在线/离线/部署中

**交互功能**:
- 鼠标拖拽平移视图
- 滚轮缩放
- 点击节点查看详情
- 悬停高亮相关连接

### 6.3 实时通信

**WebSocket连接**:
```typescript
const ws = new WebSocket('ws://localhost:8080/ws/cluster')
ws.onmessage = (event) => {
  const data = JSON.parse(event.data)
  // 更新拓扑图
  updateTopology(data)
}
```

**推送事件**:
- 服务器状态变更
- 部署进度更新
- 扫描进度更新
- 资源使用更新

---

## 备份恢复机制

### 7.1 备份策略

| 备份类型 | 频率 | 保留期 | 存储位置 |
|---------|------|--------|---------|
| 全量备份 | 每日 | 30天 | 本地文件系统 |
| 配置备份 | 每次变更 | 90天 | 本地文件系统 |
| 代理配置 | 每日 | 7天 | 本地文件系统 |

### 7.2 备份文件结构

```
./backups/
├── full_backup_20260307_100000.tar.gz
├── config_backup_20260307_100000.json
├── proxy_backup_server1_20260307_100000.json
├── exports/
│   └── full_backup_20260307_100000.tar.gz
└── .metadata.json
```

### 7.3 恢复流程

**场景1: 完整系统恢复**
```
1. 停止所有服务
2. 选择全量备份
3. 执行恢复
4. 验证数据完整性
5. 重启服务
```

**场景2: 单服务器配置恢复**
```
1. 选择配置备份
2. 指定目标服务器
3. 执行恢复
4. 重启代理服务
5. 验证配置生效
```

---

## 自动伸缩策略

### 8.1 伸缩策略配置

```json
{
  "auto_scale": true,
  "min_servers": 2,
  "max_servers": 10,
  "scale_policy": "cpu",
  "scale_threshold": 80.0
}
```

### 8.2 伸缩决策逻辑

```
每30秒检查一次:
1. 获取分组内所有服务器指标
2. 计算平均值
3. 判断是否需要伸缩
4. 执行伸缩操作
5. 记录伸缩事件
```

### 8.3 冷却时间

- 扩容后冷却: 5分钟
- 缩容后冷却: 10分钟
- 避免频繁伸缩

---

## 部署方案

### 9.1 单机部署

```bash
# 1. 克隆代码
git clone <repository>
cd proxy_server

# 2. 启动后端
cd backend
go run main.go

# 3. 启动前端
cd frontend
npm install
npm run dev
```

### 9.2 Docker部署

```yaml
# docker-compose.yml
version: '3.8'
services:
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
      - ./backups:/app/backups
  
  frontend:
    build: ./frontend
    ports:
      - "3000:80"
    depends_on:
      - backend
```

### 9.3 生产部署建议

| 组件 | 配置建议 |
|------|---------|
| **后端** | 2核4G，SSD存储 |
| **前端** | Nginx反向代理 |
| **数据库** | SQLite（小规模）/ PostgreSQL（大规模） |
| **备份** | 定期备份到对象存储 |
| **监控** | Prometheus + Grafana |

---

## 安全设计

### 10.1 认证与授权

- **API认证**: JWT Token
- **SSH认证**: 公钥认证（推荐）/ 密码认证
- **权限模型**: RBAC（基于角色的访问控制）

### 10.2 数据安全

- **密码加密**: AES-256加密存储
- **传输加密**: HTTPS/WSS
- **连接安全**: SSH密钥对管理

### 10.3 审计日志

- 用户操作日志
- 部署任务日志
- 伸缩事件日志
- 访问日志

---

## 附录

### A. 错误码定义

| 错误码 | 说明 |
|--------|------|
| 5000 | 服务器不存在 |
| 5001 | 服务器已存在 |
| 5002 | 连接失败 |
| 5003 | 连接不存在 |
| 5004 | 认证失败 |
| 5005 | 会话创建失败 |
| 5006 | 文件传输失败 |
| 5007 | 部署失败 |
| 5008 | 备份失败 |
| 5009 | 恢复失败 |
| 5010 | 伸缩失败 |
| 5011 | 分组不存在 |

### B. 性能指标

| 指标 | 目标值 |
|------|--------|
| API响应时间 | < 200ms |
| 扫描速度 | > 100 IP/s |
| 部署时间 | < 5分钟/台 |
| 备份速度 | > 10 MB/s |
| 内存占用 | < 500MB |

### C. 扩展性设计

- **水平扩展**: 无状态服务，支持多实例
- **插件化**: 支持自定义部署脚本
- **多租户**: 支持多租户隔离（未来）
- **云原生**: 支持K8s部署（未来）

---

## 总结

本架构设计提供了完整的代理集群管理解决方案，包括：

1. **自动化**: 局域网扫描、自动部署、自动伸缩
2. **可视化**: 实时拓扑图、资源监控、部署进度
3. **可靠性**: 备份恢复、故障重试、连接池
4. **安全性**: 加密存储、认证授权、审计日志
5. **扩展性**: 模块化设计、插件支持、水平扩展

该架构可满足中小规模代理集群的管理需求，支持未来扩展到更大规模。
