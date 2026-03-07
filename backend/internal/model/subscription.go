package model

import "time"

// ParseFormat 解析格式类型
type ParseFormat string

const (
	ParseFormatAuto      ParseFormat = "auto"      // 自动检测
	ParseFormatBase64    ParseFormat = "base64"    // Base64编码节点列表
	ParseFormatClash     ParseFormat = "clash"     // Clash配置文件
	ParseFormatSurge     ParseFormat = "surge"     // Surge配置文件
	ParseFormatQuantumult ParseFormat = "quantumult" // Quantumult格式
	ParseFormatSSD       ParseFormat = "ssd"       // SSD格式
)

// Subscription 订阅模型
type Subscription struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	URL            string      `json:"url"`
	Type           string      `json:"type"`
	ParseFormat    ParseFormat `json:"parseFormat"`    // 解析格式
	Status         string      `json:"status"`
	LastUpdate     time.Time   `json:"lastUpdate"`
	NodeCount      int         `json:"nodeCount"`
	AutoUpdate     bool        `json:"autoUpdate"`
	UpdateInterval int         `json:"updateInterval"` // 更新间隔（小时）
	CreatedAt      time.Time   `json:"createdAt"`
	UpdatedAt      time.Time   `json:"updatedAt"`
}

// SubscriptionCreateRequest 创建订阅请求
type SubscriptionCreateRequest struct {
	Name           string      `json:"name" binding:"required"`
	URL            string      `json:"url" binding:"required"`
	Type           string      `json:"type"`
	ParseFormat    ParseFormat `json:"parseFormat"`    // 解析格式
	AutoUpdate     bool        `json:"autoUpdate"`
	UpdateInterval int         `json:"updateInterval"` // 默认24小时
}

// SubscriptionUpdateRequest 更新订阅请求
type SubscriptionUpdateRequest struct {
	Name           string      `json:"name"`
	URL            string      `json:"url"`
	Type           string      `json:"type"`
	ParseFormat    ParseFormat `json:"parseFormat"`    // 解析格式
	AutoUpdate     bool        `json:"autoUpdate"`
	UpdateInterval int         `json:"updateInterval"`
}
