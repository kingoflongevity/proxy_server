package model

import "time"

// ClusterServer 集群服务器
type ClusterServer struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	IP            string    `json:"ip"`
	Port          int       `json:"port"`
	Username      string    `json:"username"`
	Password      string    `json:"password"` // 加密存储
	OSType        string    `json:"osType"`      // ubuntu, debian, centos, alpine
	Arch          string    `json:"arch"`        // amd64, arm64
	Status        string    `json:"status"`      // idle, connecting, active, error, deploying
	ProxyType     string    `json:"proxyType"`   // vmess, vless, trojan, ss
	ProxyConfig   string    `json:"proxyConfig"` // JSON配置
	ProxyEnabled  bool      `json:"proxyEnabled"`
	ProxyPort     int       `json:"proxyPort"`
	GroupID       string    `json:"groupId"`
	Tags          []string  `json:"tags"`
	PrivateKey    string    `json:"privateKey"`
	OSVersion     string    `json:"osVersion"`
	LastHeartbeat time.Time `json:"lastHeartbeat"`
	CPU           float64   `json:"cpu"`
	Memory        uint64    `json:"memory"`     // MB
	Disk          uint64    `json:"disk"`       // GB
	CPUUsage      float64   `json:"cpuUsage"`
	MemoryUsage   float64   `json:"memoryUsage"`
	BandwidthUp   uint64    `json:"bandwidthUp"`
	BandwidthDown uint64    `json:"bandwidthDown"`
	Connections   int       `json:"connections"`
	Latency       int       `json:"latency"`      // ms
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	DeployedAt    time.Time `json:"deployedAt"`
	LastDeployAt  time.Time `json:"lastDeployAt"`
}

// BackupType 备份类型
type BackupType string

const (
	BackupTypeFull   BackupType = "full"
	BackupTypeConfig BackupType = "config"
	BackupTypeProxy  BackupType = "proxy"
)

// ServerListQuery 服务器列表查询参数
type ServerListQuery struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"pageSize" json:"pageSize"`
	GroupID  string `form:"groupId" json:"groupId"`
	Status   string `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
	SortBy   string `form:"sortBy" json:"sortBy"`
	SortDesc bool   `form:"sortDesc" json:"sortDesc"`
}

// ServerGroup 服务器分组
type ServerGroup struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ServerCount    int       `json:"serverCount"`
	AutoScale      bool      `json:"autoScale"`
	MinServers     int       `json:"minServers"`
	MaxServers     int       `json:"maxServers"`
	ScalePolicy    string    `json:"scalePolicy"`
	ScaleThreshold float64   `json:"scaleThreshold"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// ScalePolicyUpdateRequest 伸缩策略更新请求
type ScalePolicyUpdateRequest struct {
	GroupID        string  `json:"groupId"`
	AutoScale      bool    `json:"autoScale"`
	MinServers     int     `json:"minServers"`
	MaxServers     int     `json:"maxServers"`
	ScalePolicy    string  `json:"scalePolicy"`
	ScaleThreshold float64 `json:"scaleThreshold"`
}

// ServerStatus 服务器状态常量
const (
	ServerStatusIdle       = "idle"
	ServerStatusConnecting = "connecting"
	ServerStatusOnline     = "active"
	ServerStatusError      = "error"
	ServerStatusDeploying  = "deploying"
)

// BackupRecord 备份记录
type BackupRecord struct {
	ID           string      `json:"id"`
	Type         BackupType  `json:"type"`         // full, config, proxy
	ServerID     string      `json:"serverId"`     // 空表示全局备份
	Name         string      `json:"name"`
	Size         int64       `json:"size"`         // bytes
	MD5          string      `json:"md5"`
	FilePath     string      `json:"filePath"`
	FileSize     int64       `json:"fileSize"`
	Checksum     string      `json:"checksum"`
	Status       string      `json:"status"`
	ErrorMessage string      `json:"errorMessage"`
	CreatedAt    time.Time   `json:"createdAt"`
}

// ScanTask 扫描任务
type ScanTask struct {
	ID          string      `json:"id"`
	CIDR        string      `json:"cidr"`        // 如 192.168.1.0/24
	NetworkCIDR string      `json:"networkCidr"` // 网络CIDR
	Status      string      `json:"status"`      // pending, running, completed, failed
	Found       int         `json:"found"`       // 发现的服务器数
	Progress    float64     `json:"progress"`    // 0-100
	Results     []string    `json:"results"`     // 扫描结果
	StartedAt   *time.Time  `json:"startedAt"`
	CompletedAt *time.Time  `json:"completedAt"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

// DeployTask 部署任务
type DeployTask struct {
	ID          string       `json:"id"`
	ServerID    string       `json:"serverId"`
	Steps       []DeployStep `json:"steps"`
	CurrentStep int          `json:"currentStep"`
	Status      string       `json:"status"`     // pending, running, success, failed
	Progress    float64      `json:"progress"`
	Logs        []string     `json:"logs"`
	StartedAt   *time.Time   `json:"startedAt"`
	CompletedAt *time.Time   `json:"completedAt"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
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
	ID          string     `json:"id"`
	ServerID    string     `json:"serverId"`
	GroupID     string     `json:"groupId"`
	Action      string     `json:"action"`      // scale-up, scale-down
	Type        string     `json:"type"`        // cpu, memory, connections
	Reason      string     `json:"reason"`      // 触发原因
	Metric      float64    `json:"metric"`      // 触发阈值
	TargetCount int        `json:"targetCount"` // 目标数量
	Status      string     `json:"status"`      // pending, success, failed
	CompletedAt *time.Time `json:"completedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
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
