package model

import "time"

// Subscription 订阅模型
type Subscription struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	URL            string    `json:"url"`
	Type           string    `json:"type"`
	Status         string    `json:"status"`
	LastUpdate     time.Time `json:"lastUpdate"`
	NodeCount      int       `json:"nodeCount"`
	AutoUpdate     bool      `json:"autoUpdate"`
	UpdateInterval int       `json:"updateInterval"` // 更新间隔（小时）
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// SubscriptionCreateRequest 创建订阅请求
type SubscriptionCreateRequest struct {
	Name           string `json:"name" binding:"required"`
	URL            string `json:"url" binding:"required"`
	Type           string `json:"type"`
	AutoUpdate     bool   `json:"autoUpdate"`
	UpdateInterval int    `json:"updateInterval"` // 默认24小时
}

// SubscriptionUpdateRequest 更新订阅请求
type SubscriptionUpdateRequest struct {
	Name           string `json:"name"`
	URL            string `json:"url"`
	Type           string `json:"type"`
	AutoUpdate     bool   `json:"autoUpdate"`
	UpdateInterval int    `json:"updateInterval"`
}
