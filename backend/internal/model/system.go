package model

import "time"

type SystemStatus struct {
	Connected      bool      `json:"connected"`
	CurrentNode    *Node     `json:"currentNode,omitempty"`
	StartTime      time.Time `json:"startTime"`
	Uptime         int64     `json:"uptime"`
	GoroutineCount int       `json:"goroutineCount"`
	Version        string    `json:"version"`
	Mode           string    `json:"mode"`
}

type TrafficStats struct {
	Upload        uint64    `json:"upload"`
	Download      uint64    `json:"download"`
	UploadSpeed   uint64    `json:"uploadSpeed"`
	DownloadSpeed uint64    `json:"downloadSpeed"`
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
	Theme          string   `json:"theme"`
	Language       string   `json:"language"`
	AutoStart      bool     `json:"autoStart"`
	SilentStart    bool     `json:"silentStart"`
	AllowLan       bool     `json:"allowLan"`
	BindAddress    string   `json:"bindAddress"`
	Port           int      `json:"port"`
	SocksPort      int      `json:"socksPort"`
	HttpPort       int      `json:"httpPort"`
	MixedPort      int      `json:"mixedPort"`
	LogLevel       string   `json:"logLevel"`
	ProxyMode      string   `json:"proxyMode"`
	// 高级设置
	DNSServers     []string `json:"dnsServers"`
	EnableMux      bool     `json:"enableMux"`
	EnableIpv6     bool     `json:"enableIpv6"`
	DomainStrategy string   `json:"domainStrategy"`
	TunMode        bool     `json:"tunMode"`
}

type UpdateSettingsRequest struct {
	Theme          string   `json:"theme"`
	Language       string   `json:"language"`
	AutoStart      bool     `json:"autoStart"`
	SilentStart    bool     `json:"silentStart"`
	AllowLan       bool     `json:"allowLan"`
	BindAddress    string   `json:"bindAddress"`
	Port           int      `json:"port"`
	SocksPort      int      `json:"socksPort"`
	HttpPort       int      `json:"httpPort"`
	MixedPort      int      `json:"mixedPort"`
	LogLevel       string   `json:"logLevel"`
	ProxyMode      string   `json:"proxyMode"`
	// 高级设置
	DNSServers     []string `json:"dnsServers"`
	EnableMux      bool     `json:"enableMux"`
	EnableIpv6     bool     `json:"enableIpv6"`
	DomainStrategy string   `json:"domainStrategy"`
	TunMode        bool     `json:"tunMode"`
}

type ConnectionStatus struct {
	Connected       bool   `json:"connected"`
	CurrentMode     string `json:"currentMode"`
	UploadSpeed     uint64 `json:"uploadSpeed"`
	DownloadSpeed   uint64 `json:"downloadSpeed"`
	UploadTotal     uint64 `json:"uploadTotal"`
	DownloadTotal   uint64 `json:"downloadTotal"`
	ConnectionCount int    `json:"connectionCount"`
}

type SystemInfo struct {
	Version       string `json:"version"`
	GoVersion     string `json:"goVersion"`
	Os            string `json:"os"`
	Arch          string `json:"arch"`
	NumCPU        int    `json:"numCpu"`
	GoroutineNum  int    `json:"goroutineNum"`
	Uptime        int64  `json:"uptime"`
}

func (s *SystemSettings) ExportConfig() string {
	return "exported_config"
}

func (s *SystemSettings) ImportConfig(config string) error {
	return nil
}
