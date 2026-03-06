package model

import "time"

// NodeType 节点类型
type NodeType string

const (
	NodeTypeVLESS      NodeType = "vless"
	NodeTypeVMess      NodeType = "vmess"
	NodeTypeTrojan     NodeType = "trojan"
	NodeTypeShadowsocks NodeType = "shadowsocks"
	NodeTypeSSR        NodeType = "ssr"
)

// NetworkType 传输层网络类型
type NetworkType string

const (
	NetworkTCP    NetworkType = "tcp"
	NetworkWS     NetworkType = "ws"
	NetworkGRPC   NetworkType = "grpc"
	NetworkHTTP   NetworkType = "http"
	NetworkHTTP2  NetworkType = "h2"
	NetworkXHTTP  NetworkType = "xhttp" // Xray-core v26.2.6新增
	NetworkKCP    NetworkType = "mkcp"
)

// SecurityType 安全类型
type SecurityType string

const (
	SecurityNone    SecurityType = "none"
	SecurityTLS     SecurityType = "tls"
	SecurityReality SecurityType = "reality"
)

// FinalmaskConfig Finalmask配置（Xray-core v26.2.6新增）
type FinalmaskConfig struct {
	XICMP  bool `json:"xicmp,omitempty"`  // XICMP伪装
	XDNS   bool `json:"xdns,omitempty"`   // XDNS伪装
	Header map[string]bool `json:"header,omitempty"` // header-* 伪装
	MKCP   map[string]bool `json:"mkcp,omitempty"`   // mkcp-* 伪装
}

// XHTTPConfig XHTTP传输配置（Xray-core v26.2.6新增）
type XHTTPConfig struct {
	EnableCDNBypass bool   `json:"enable_cdn_bypass,omitempty"` // CDN绕过选项
	Mode           string `json:"mode,omitempty"`             // stream或packet
	DownloadSettings *struct {
		Address string `json:"address,omitempty"`
		Port    int    `json:"port,omitempty"`
	} `json:"download_settings,omitempty"`
}

// Node 节点模型
type Node struct {
	ID             string    `json:"id"`
	SubscriptionID string    `json:"subscription_id"`
	Name           string    `json:"name"`
	Type           NodeType  `json:"type"`
	Server         string    `json:"server"`
	Port           int       `json:"port"`
	UUID           string    `json:"uuid,omitempty"`
	Password       string    `json:"password,omitempty"`
	Method         string    `json:"method,omitempty"` // Shadowsocks加密方式
	Network        string    `json:"network,omitempty"` // tcp, ws, grpc, http, xhttp
	Security       string    `json:"security,omitempty"` // tls, reality, none
	SNI            string    `json:"sni,omitempty"`
	Host           string    `json:"host,omitempty"`
	Path           string    `json:"path,omitempty"`
	Alpn           []string  `json:"alpn,omitempty"`
	
	// REALITY配置 (Xray-core v25+)
	RealityPublicKey string `json:"reality_public_key,omitempty"`
	RealityShortID   string `json:"reality_short_id,omitempty"`
	RealityFingerprint string `json:"reality_fingerprint,omitempty"`
	RealitySpiderX   string `json:"reality_spider_x,omitempty"` // SpiderX参数
	
	// TLS配置 (Xray-core v26.2.6重要变更)
	PinnedPeerCertSha256 string `json:"pinned_peer_cert_sha256,omitempty"` // 替代allowInsecure
	VerifyPeerCertByName string `json:"verify_peer_cert_by_name,omitempty"` // 证书名称验证
	Fingerprint         string `json:"fingerprint,omitempty"` // TLS指纹（chrome, firefox等）
	
	// XHTTP配置 (Xray-core v26.2.6新增)
	XHTTPConfig *XHTTPConfig `json:"xhttp_config,omitempty"`
	
	// Finalmask配置 (Xray-core v26.2.6新增)
	Finalmask *FinalmaskConfig `json:"finalmask,omitempty"`
	
	// 其他配置
	Flow          string   `json:"flow,omitempty"` // VLESS flow控制
	ServiceName   string   `json:"service_name,omitempty"` // gRPC service name
	Headers       map[string]string `json:"headers,omitempty"`
	
	// VMess特殊配置
	AlterID       int    `json:"alter_id,omitempty"` // VMess AlterID（v26.2.6建议设为0）
	
	// 性能指标
	Latency       int       `json:"latency"` // 延迟（毫秒）
	Speed         int       `json:"speed"` // 速度（MB/s）
	Score         int       `json:"score"` // 综合评分
	LastTest      time.Time `json:"last_test"`
	
	// 状态
	Enabled       bool      `json:"enabled"`
	Connected     bool      `json:"connected"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NodeUpdateRequest 更新节点请求
type NodeUpdateRequest struct {
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
}

// NodeTestRequest 测试节点请求
type NodeTestRequest struct {
	NodeID string `json:"node_id" binding:"required"`
}

// NodeConnectRequest 连接节点请求
type NodeConnectRequest struct {
	NodeID string `json:"node_id" binding:"required"`
}

// NodeListQuery 节点列表查询参数
type NodeListQuery struct {
	Page           int    `form:"page" json:"page"`
	PageSize       int    `form:"page_size" json:"page_size"`
	SubscriptionID string `form:"subscription_id" json:"subscription_id"`
	Type           string `form:"type" json:"type"`
	Enabled        *bool  `form:"enabled" json:"enabled"`
	SortBy         string `form:"sort_by" json:"sort_by"` // latency, speed, score
	SortOrder      string `form:"sort_order" json:"sort_order"` // asc, desc
}
