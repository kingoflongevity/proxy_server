package model

import (
	"time"
)

type RequestLog struct {
	ID           string            `json:"id"`
	Timestamp    time.Time         `json:"timestamp"`
	ClientIP     string            `json:"clientIp"`
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Path         string            `json:"path"`
	QueryString  string            `json:"queryString,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	Body         string            `json:"body,omitempty"`
	BodySize     int64             `json:"bodySize"`
	UserAgent    string            `json:"userAgent,omitempty"`
	Protocol     string            `json:"protocol"`
	StatusCode   int               `json:"statusCode"`
	ResponseTime int64             `json:"responseTimeMs"`
	ResponseSize int64             `json:"responseSize"`
	Error        string            `json:"error,omitempty"`
	UserID       string            `json:"userId,omitempty"`
	// 额外详细信息
	ClientPort    string `json:"clientPort,omitempty"`
	Scheme        string `json:"scheme,omitempty"`
	Host          string `json:"host,omitempty"`
	ContentType   string `json:"contentType,omitempty"`
	ContentLength int64  `json:"contentLength"`
	Referrer      string `json:"referrer,omitempty"`
	Origin        string `json:"origin,omitempty"`
	XForwardedFor string `json:"xForwardedFor,omitempty"`
	XRealIP       string `json:"xRealIp,omitempty"`
	TLS           bool   `json:"tls"`
}

type TrafficLog struct {
	ID            string    `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	ClientIP      string    `json:"clientIp"`
	ServerIP      string    `json:"serverIp"`
	Domain        string    `json:"domain"`
	Method        string    `json:"method"`
	Path          string    `json:"path"`
	UploadBytes   int64     `json:"uploadBytes"`
	DownloadBytes int64     `json:"downloadBytes"`
	Duration      int64     `json:"durationMs"`
	StatusCode    int       `json:"statusCode"`
	Protocol      string    `json:"protocol"`
	UserID        string    `json:"userId,omitempty"`
}

type LogQuery struct {
	StartTime  *time.Time `form:"start_time" json:"startTime"`
	EndTime    *time.Time `form:"end_time" json:"endTime"`
	ClientIP   string     `form:"client_ip" json:"clientIp"`
	Method     string     `form:"method" json:"method"`
	StatusCode int        `form:"status_code" json:"statusCode"`
	URL        string     `form:"url" json:"url"`
	Keyword    string     `form:"keyword" json:"keyword"`
	UserID     string     `form:"user_id" json:"userId"`
	Limit      int        `form:"limit" json:"limit"`
	Offset     int        `form:"offset" json:"offset"`
}

type TrafficQuery struct {
	StartTime time.Time `form:"start_time" json:"startTime"`
	EndTime   time.Time `form:"end_time" json:"endTime"`
	ClientIP  string    `form:"client_ip" json:"clientIp"`
	Domain    string    `form:"domain" json:"domain"`
	Limit     int       `form:"limit" json:"limit"`
	Offset    int       `form:"offset" json:"offset"`
}

type LogStats struct {
	TotalRequests   int64            `json:"totalRequests"`
	TotalTraffic    int64            `json:"totalTraffic"`
	UploadBytes     int64            `json:"uploadBytes"`
	DownloadBytes   int64            `json:"downloadBytes"`
	AvgResponseTime int64            `json:"avgResponseTimeMs"`
	StatusCounts    map[string]int64 `json:"statusCounts"`
	TopDomains      []DomainStats    `json:"topDomains"`
	TopClients      []ClientStats    `json:"topClients"`
}

type DomainStats struct {
	Domain        string `json:"domain"`
	RequestCount  int64  `json:"requestCount"`
	UploadBytes   int64  `json:"uploadBytes"`
	DownloadBytes int64  `json:"downloadBytes"`
}

type ClientStats struct {
	ClientIP      string `json:"clientIp"`
	RequestCount  int64  `json:"requestCount"`
	UploadBytes   int64  `json:"uploadBytes"`
	DownloadBytes int64  `json:"downloadBytes"`
}

type LogConfig struct {
	Enabled        bool     `json:"enabled"`
	LogLevel       string   `json:"logLevel"`
	MaxFileSize    int64    `json:"maxFileSize"`
	MaxBackups     int      `json:"maxBackups"`
	MaxAge         int      `json:"maxAge"`
	Compress       bool     `json:"compress"`
	Directory      string   `json:"directory"`
	SensitiveWords []string `json:"sensitiveWords"`
}

func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Enabled:     true,
		LogLevel:    "info",
		MaxFileSize: 100 * 1024 * 1024,
		MaxBackups:  30,
		MaxAge:      90,
		Compress:    true,
		Directory:   "./logs/traffic",
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
