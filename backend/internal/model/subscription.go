package model

import "time"

// Subscription 订阅模型
type Subscription struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	LastUpdate  time.Time `json:"last_update"`
	NodeCount   int       `json:"node_count"`
	AutoUpdate  bool      `json:"auto_update"`
	UpdateInterval int    `json:"update_interval"` // 更新间隔（小时）
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SubscriptionCreateRequest 创建订阅请求
type SubscriptionCreateRequest struct {
	Name           string `json:"name" binding:"required"`
	URL            string `json:"url" binding:"required"`
	AutoUpdate     bool   `json:"auto_update"`
	UpdateInterval int    `json:"update_interval"` // 默认24小时
}

// SubscriptionUpdateRequest 更新订阅请求
type SubscriptionUpdateRequest struct {
	Name           string `json:"name"`
	URL            string `json:"url"`
	AutoUpdate     bool   `json:"auto_update"`
	UpdateInterval int    `json:"update_interval"`
}
