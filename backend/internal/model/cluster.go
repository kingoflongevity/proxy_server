package model

import "time"

// ClusterServer 集群服务器
type ClusterServer struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	IP          string    `json:"ip"`
	Port        int       `json:"port"`
	Username    string    `json:"username"`
	Password    string    `json:"password"` // 加密存储
	OSType      string    `json:"osType"`      // ubuntu, debian, centos, alpine
	Arch        string    `json:"arch"`        // amd64, arm64
	Status      string    `json:"status"`      // idle, connecting, active, error, deploying
	ProxyType   string    `json:"proxyType"`   // vmess, vless, trojan, ss
	ProxyConfig string    `json:"proxyConfig"` // JSON配置
	GroupID     string    `json:"groupId"`
	CPU         float64   `json:"cpu"`
	Memory      uint64    `json:"memory"`     // MB
	Connections int       `json:"connections"`
	Latency     int       `json:"latency"`      // ms
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	LastDeployAt time.Time `json:"lastDeployAt"`
}

// ServerGroup 服务器分组
type ServerGroup struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ServerCount int       `json:"serverCount"`
	CreatedAt   time.Time `json:"createdAt"`
}

// BackupRecord 备份记录
type BackupRecord struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`        // full, config, proxy
	ServerID    string    `json:"serverId"`    // 空表示全局备份
	Name        string    `json:"name"`
	Size        int64     `json:"size"`        // bytes
	MD5         string    `json:"md5"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ScanTask 扫描任务
type ScanTask struct {
	ID        string    `json:"id"`
	CIDR      string    `json:"cidr"`        // 如 192.168.1.0/24
	Status    string    `json:"status"`      // pending, running, completed, failed
	Found     int       `json:"found"`       // 发现的服务器数
	Progress  float64   `json:"progress"`    // 0-100
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// DeployTask 部署任务
type DeployTask struct {
	ID          string    `json:"id"`
	ServerID    string    `json:"serverId"`
	Steps       []DeployStep `json:"steps"`
	CurrentStep int          `json:"currentStep"`
	Status      string      `json:"status"`     // pending, running, success, failed
	Progress    float64    `json:"progress"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// DeployStep 部署步骤
type DeployStep struct {
	Name     string `json:"name"`
	Status   string `json:"status"`   // pending, running, success, failed
	Message  string `json:"message"`
	Duration int    `json:"duration"` // seconds
}

// ScaleEvent 伸缩事件
type ScaleEvent struct {
	ID        string    `json:"id"`
	ServerID  string    `json:"serverId"`
	Action    string    `json:"action"`    // scale-up, scale-down
	Reason    string    `json:"reason"`    // cpu, memory, connections
	Metric    float64   `json:"metric"`    // 触发阈值
	CreatedAt time.Time `json:"createdAt"`
}

// ClusterTopology 集群拓扑
type ClusterTopology struct {
	Servers      []ClusterServer `json:"servers"`
	Groups       []ServerGroup   `json:"groups"`
	Connections  []Connection     `json:"connections"`
}

// Connection 连接关系
type Connection struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // proxy, backup, sync
}
