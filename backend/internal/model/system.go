package model

import "time"

type SystemStatus struct {
	Connected      bool      `json:"connected"`
	CurrentNode    *Node     `json:"current_node,omitempty"`
	StartTime      time.Time `json:"start_time"`
	Uptime         int64     `json:"uptime"`
	GoroutineCount int       `json:"goroutine_count"`
	Version        string    `json:"version"`
	Mode           string    `json:"mode"`
}

type TrafficStats struct {
	Upload        uint64    `json:"upload"`
	Download      uint64    `json:"download"`
	UploadSpeed   uint64    `json:"upload_speed"`
	DownloadSpeed uint64    `json:"download_speed"`
	Connections   int       `json:"connections"`
	Timestamp     time.Time `json:"timestamp"`
}

type LogEntry struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Node      string    `json:"node,omitempty"`
}

type SystemLogQuery struct {
	Level   string `form:"level" json:"level"`
	Limit   int    `form:"limit" json:"limit"`
	Offset  int    `form:"offset" json:"offset"`
	Node    string `form:"node" json:"node"`
	Keyword string `form:"keyword" json:"keyword"`
}

type SystemSettings struct {
	Theme       string `json:"theme"`
	Language    string `json:"language"`
	AutoStart   bool   `json:"auto_start"`
	SilentStart bool   `json:"silent_start"`
	AllowLan    bool   `json:"allow_lan"`
	BindAddress string `json:"bind_address"`
	Port        int    `json:"port"`
	SocksPort   int    `json:"socks_port"`
	HttpPort    int    `json:"http_port"`
	MixedPort   int    `json:"mixed_port"`
	LogLevel    string `json:"log_level"`
	ProxyMode   string `json:"proxy_mode"`
}

type UpdateSettingsRequest struct {
	Theme       string `json:"theme"`
	Language    string `json:"language"`
	AutoStart   bool   `json:"auto_start"`
	SilentStart bool   `json:"silent_start"`
	AllowLan    bool   `json:"allow_lan"`
	BindAddress string `json:"bind_address"`
	Port        int    `json:"port"`
	SocksPort   int    `json:"socks_port"`
	HttpPort    int    `json:"http_port"`
	MixedPort   int    `json:"mixed_port"`
	LogLevel    string `json:"log_level"`
	ProxyMode   string `json:"proxy_mode"`
}

type ConnectionStatus struct {
	Connected      bool   `json:"connected"`
	CurrentMode    string `json:"current_mode"`
	UploadSpeed    uint64 `json:"upload_speed"`
	DownloadSpeed  uint64 `json:"download_speed"`
	UploadTotal    uint64 `json:"upload_total"`
	DownloadTotal  uint64 `json:"download_total"`
	ConnectionCount int   `json:"connection_count"`
}

type SystemInfo struct {
	Version       string `json:"version"`
	GoVersion     string `json:"go_version"`
	Os            string `json:"os"`
	Arch          string `json:"arch"`
	NumCPU        int    `json:"num_cpu"`
	GoroutineNum  int    `json:"goroutine_num"`
	Uptime        int64  `json:"uptime"`
}

func (s *SystemSettings) ExportConfig() string {
	return "exported_config"
}

func (s *SystemSettings) ImportConfig(config string) error {
	return nil
}
