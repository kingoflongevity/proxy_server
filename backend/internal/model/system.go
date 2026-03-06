package model

import "time"

// SystemStatus 系统状态
type SystemStatus struct {
	Connected      bool      `json:"connected"`
	CurrentNode    *Node     `json:"current_node,omitempty"`
	StartTime      time.Time `json:"start_time"`
	Uptime         int64     `json:"uptime"` // 运行时间（秒）
	GoroutineCount int       `json:"goroutine_count"`
	Version        string    `json:"version"`
}

// TrafficStats 流量统计
type TrafficStats struct {
	Upload       uint64    `json:"upload"` // 上传流量（字节）
	Download     uint64    `json:"download"` // 下载流量（字节）
	UploadSpeed  uint64    `json:"upload_speed"` // 上传速度（字节/秒）
	DownloadSpeed uint64   `json:"download_speed"` // 下载速度（字节/秒）
	Connections  int       `json:"connections"` // 当前连接数
	Timestamp    time.Time `json:"timestamp"`
}

// LogEntry 日志条目
type LogEntry struct {
	Level     string    `json:"level"` // INFO, WARN, ERROR, DEBUG
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Node      string    `json:"node,omitempty"`
}

// LogQuery 日志查询参数
type LogQuery struct {
	Level   string `form:"level" json:"level"`
	Limit   int    `form:"limit" json:"limit"`
	Offset  int    `form:"offset" json:"offset"`
	Node    string `form:"node" json:"node"`
	Keyword string `form:"keyword" json:"keyword"`
}
