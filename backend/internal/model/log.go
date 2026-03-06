package model

import (
	"time"
)

type RequestLog struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	ClientIP      string                 `json:"client_ip"`
	Method        string                 `json:"method"`
	URL           string                 `json:"url"`
	Path          string                 `json:"path"`
	QueryString   string                 `json:"query_string,omitempty"`
	Headers       map[string]string      `json:"headers,omitempty"`
	Body          string                 `json:"body,omitempty"`
	BodySize      int64                  `json:"body_size"`
	UserAgent     string                 `json:"user_agent,omitempty"`
	Protocol      string                 `json:"protocol"`
	StatusCode    int                     `json:"status_code"`
	ResponseTime int64                   `json:"response_time_ms"`
	ResponseSize int64                  `json:"response_size"`
	Error         string                 `json:"error,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
}

type TrafficLog struct {
	ID            string    `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	ClientIP      string    `json:"client_ip"`
	ServerIP      string    `json:"server_ip"`
	Domain        string    `json:"domain"`
	Method        string    `json:"method"`
	Path          string    `json:"path"`
	UploadBytes   int64     `json:"upload_bytes"`
	DownloadBytes int64     `json:"download_bytes"`
	Duration      int64     `json:"duration_ms"`
	StatusCode    int       `json:"status_code"`
	Protocol      string    `json:"protocol"`
	UserID        string    `json:"user_id,omitempty"`
}

type LogQuery struct {
	StartTime   *time.Time `form:"start_time" json:"start_time"`
	EndTime     *time.Time `form:"end_time" json:"end_time"`
	ClientIP    string     `form:"client_ip" json:"client_ip"`
	Method      string     `form:"method" json:"method"`
	StatusCode  int        `form:"status_code" json:"status_code"`
	URL         string     `form:"url" json:"url"`
	Keyword     string     `form:"keyword" json:"keyword"`
	UserID      string     `form:"user_id" json:"user_id"`
	Limit       int        `form:"limit" json:"limit"`
	Offset      int        `form:"offset" json:"offset"`
}

type TrafficQuery struct {
	StartTime time.Time  `form:"start_time" json:"start_time"`
	EndTime   time.Time  `form:"end_time" json:"end_time"`
	ClientIP  string     `form:"client_ip" json:"client_ip"`
	Domain    string     `form:"domain" json:"domain"`
	Limit     int        `form:"limit" json:"limit"`
	Offset    int        `form:"offset" json:"offset"`
}

type LogStats struct {
	TotalRequests   int64            `json:"total_requests"`
	TotalTraffic    int64            `json:"total_traffic"`
	UploadBytes     int64            `json:"upload_bytes"`
	DownloadBytes   int64            `json:"download_bytes"`
	AvgResponseTime int64            `json:"avg_response_time_ms"`
	StatusCounts    map[string]int64  `json:"status_counts"`
	TopDomains      []DomainStats    `json:"top_domains"`
	TopClients      []ClientStats    `json:"top_clients"`
}

type DomainStats struct {
	Domain        string `json:"domain"`
	RequestCount  int64  `json:"request_count"`
	UploadBytes   int64  `json:"upload_bytes"`
	DownloadBytes int64  `json:"download_bytes"`
}

type ClientStats struct {
	ClientIP      string `json:"client_ip"`
	RequestCount  int64  `json:"request_count"`
	UploadBytes   int64  `json:"upload_bytes"`
	DownloadBytes int64  `json:"download_bytes"`
}

type LogConfig struct {
	Enabled         bool   `json:"enabled"`
	LogLevel        string `json:"log_level"`
	MaxFileSize     int64  `json:"max_file_size"`
	MaxBackups      int    `json:"max_backups"`
	MaxAge          int    `json:"max_age"`
	Compress        bool   `json:"compress"`
	Directory       string `json:"directory"`
	SensitiveWords []string `json:"sensitive_words"`
}

func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Enabled:      true,
		LogLevel:     "info",
		MaxFileSize:  100 * 1024 * 1024,
		MaxBackups:   30,
		MaxAge:       90,
		Compress:     true,
		Directory:    "./logs/traffic",
		SensitiveWords: []string{
			"password",
			"token",
			"authorization",
			"cookie",
			"secret",
			"api_key",
			"apikey",
			"access_token",
			"refresh_token",
		},
	}
}
