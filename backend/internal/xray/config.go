package xray

import (
	"encoding/json"
	"fmt"

	"proxy_server/internal/model"
)

// XrayConfig Xray核心配置结构
type XrayConfig struct {
	Log       *LogConfig       `json:"log,omitempty"`
	API       *APIConfig       `json:"api,omitempty"`
	Inbounds  []InboundConfig  `json:"inbounds"`
	Outbounds []OutboundConfig `json:"outbounds"`
	Routing   *RoutingConfig   `json:"routing,omitempty"`
	Policy    *PolicyConfig    `json:"policy,omitempty"`
	Stats     *StatsConfig     `json:"stats,omitempty"`
	DNS       *DNSConfig       `json:"dns,omitempty"`
}

// LogConfig 日志配置
type LogConfig struct {
	Loglevel string `json:"loglevel"` // debug, info, warning, error, none
	Access   string `json:"access"`   // 访问日志路径
	Error    string `json:"error"`    // 错误日志路径
	DNSLog   bool   `json:"dnsLog"`   // DNS日志
}

// APIConfig API配置
type APIConfig struct {
	Tag      string   `json:"tag"`
	Services []string `json:"services"`
}

// InboundConfig 入站配置
type InboundConfig struct {
	Tag            string          `json:"tag"`
	Port           int             `json:"port"`
	Listen         string          `json:"listen,omitempty"`
	Protocol       string          `json:"protocol"`
	Settings       json.RawMessage `json:"settings,omitempty"`
	StreamSettings json.RawMessage `json:"streamSettings,omitempty"`
	Sniffing       *SniffingConfig `json:"sniffing,omitempty"`
}

// SniffingConfig 嗅探配置
type SniffingConfig struct {
	Enabled        bool     `json:"enabled"`
	DestOverride   []string `json:"destOverride"`
	MetadataOnly   bool     `json:"metadataOnly,omitempty"`
	RoutesOnly     bool     `json:"routesOnly,omitempty"`
	Domains        []string `json:"domains,omitempty"`
	ExcludeDomains []string `json:"excludeDomains,omitempty"`
}

// OutboundConfig 出站配置
type OutboundConfig struct {
	Tag            string          `json:"tag"`
	Protocol       string          `json:"protocol"`
	Settings       json.RawMessage `json:"settings,omitempty"`
	StreamSettings json.RawMessage `json:"streamSettings,omitempty"`
	ProxySettings  *ProxySettings  `json:"proxySettings,omitempty"`
	Mux            *MuxConfig      `json:"mux,omitempty"`
}

// ProxySettings 代理设置
type ProxySettings struct {
	Tag string `json:"tag"`
}

// MuxConfig Mux配置
type MuxConfig struct {
	Enabled     bool `json:"enabled"`
	Concurrency int  `json:"concurrency,omitempty"`
}

// RoutingConfig 路由配置
type RoutingConfig struct {
	DomainStrategy string       `json:"domainStrategy"`
	Rules          []RuleConfig `json:"rules"`
}

// RuleConfig 路由规则
type RuleConfig struct {
	Type        string   `json:"type"`
	OutboundTag string   `json:"outboundTag"`
	InboundTag  []string `json:"inboundTag,omitempty"`
	Domain      []string `json:"domain,omitempty"`
	IP          []string `json:"ip,omitempty"`
	Port        string   `json:"port,omitempty"`
	Network     string   `json:"network,omitempty"`
	Source      []string `json:"source,omitempty"`
	User        []string `json:"user,omitempty"`
	Protocol    []string `json:"protocol,omitempty"`
}

// PolicyConfig 策略配置
type PolicyConfig struct {
	Levels map[int]PolicyLevel `json:"levels"`
	System *PolicySystem       `json:"system,omitempty"`
}

// PolicyLevel 策略级别
type PolicyLevel struct {
	Handshake         uint32 `json:"handshake"`
	ConnIdle          uint32 `json:"connIdle"`
	UplinkOnly        uint32 `json:"uplinkOnly"`
	DownlinkOnly      uint32 `json:"downlinkOnly"`
	StatsUserUplink   bool   `json:"statsUserUplink"`
	StatsUserDownlink bool   `json:"statsUserDownlink"`
	BufferSize        int    `json:"bufferSize"`
}

// PolicySystem 系统策略
type PolicySystem struct {
	StatsInboundUplink    bool `json:"statsInboundUplink"`
	StatsInboundDownlink  bool `json:"statsInboundDownlink"`
	StatsOutboundUplink   bool `json:"statsOutboundUplink"`
	StatsOutboundDownlink bool `json:"statsOutboundDownlink"`
}

// StatsConfig 统计配置
type StatsConfig struct{}

// DNSConfig DNS配置
type DNSConfig struct {
	Servers []DNSServer `json:"servers"`
	Tag     string      `json:"tag"`
}

// DNSServer DNS服务器
type DNSServer struct {
	Address      string   `json:"address"`
	Port         int      `json:"port,omitempty"`
	Domains      []string `json:"domains,omitempty"`
	SkipFallback bool     `json:"skipFallback,omitempty"`
}

// ConfigGenerator Xray配置生成器
type ConfigGenerator struct {
	localPort int
	proxyMode string
}

// NewConfigGenerator 创建配置生成器
func NewConfigGenerator(localPort int) *ConfigGenerator {
	if localPort == 0 {
		localPort = 10808 // 默认SOCKS5端口
	}
	return &ConfigGenerator{
		localPort: localPort,
		proxyMode: "rule",
	}
}

// SetProxyMode 设置代理模式
func (g *ConfigGenerator) SetProxyMode(mode string) {
	g.proxyMode = mode
}

// GenerateConfig 为节点生成Xray配置
// 参数：
//   - node: 节点信息
//
// 返回：
//   - *XrayConfig: 生成的配置
//   - error: 错误信息
func (g *ConfigGenerator) GenerateConfig(node *model.Node) (*XrayConfig, error) {
	if node == nil {
		return nil, fmt.Errorf("节点不能为空")
	}

	config := &XrayConfig{
		Log: &LogConfig{
			Loglevel: "warning",
		},
		Inbounds: []InboundConfig{
			g.generateSOCKSInbound(),
			g.generateHTTPInbound(),
		},
		Outbounds: []OutboundConfig{
			g.generateOutbound(node),
			g.generateDirectOutbound(),
		},
		Routing: g.generateRouting(),
		DNS:     g.generateDNS(),
	}

	return config, nil
}

// generateSOCKSInbound 生成SOCKS5入站配置
func (g *ConfigGenerator) generateSOCKSInbound() InboundConfig {
	settings, _ := json.Marshal(map[string]interface{}{
		"udp": true,
		"ip":  "127.0.0.1",
	})

	return InboundConfig{
		Tag:      "socks-in",
		Port:     g.localPort,
		Listen:   "127.0.0.1",
		Protocol: "socks",
		Settings: settings,
		Sniffing: &SniffingConfig{
			Enabled:      true,
			DestOverride: []string{"http", "tls", "quic"},
		},
	}
}

// generateHTTPInbound 生成HTTP入站配置
func (g *ConfigGenerator) generateHTTPInbound() InboundConfig {
	settings, _ := json.Marshal(map[string]interface{}{
		"allowTransparent": false,
	})

	return InboundConfig{
		Tag:      "http-in",
		Port:     g.localPort + 1, // HTTP端口为SOCKS5端口+1
		Listen:   "127.0.0.1",
		Protocol: "http",
		Settings: settings,
		Sniffing: &SniffingConfig{
			Enabled:      true,
			DestOverride: []string{"http", "tls", "quic"},
		},
	}
}

// generateOutbound 根据节点类型生成出站配置
func (g *ConfigGenerator) generateOutbound(node *model.Node) OutboundConfig {
	var settings json.RawMessage
	var streamSettings json.RawMessage

	// 根据节点类型生成设置
	switch node.Type {
	case model.NodeTypeVLESS:
		settings = g.generateVLESSSettings(node)
		streamSettings = g.generateStreamSettings(node)
	case model.NodeTypeVMess:
		settings = g.generateVMessSettings(node)
		streamSettings = g.generateStreamSettings(node)
	case model.NodeTypeTrojan:
		settings = g.generateTrojanSettings(node)
		streamSettings = g.generateStreamSettings(node)
	case model.NodeTypeShadowsocks:
		settings = g.generateShadowsocksSettings(node)
		streamSettings = g.generateStreamSettings(node)
	}

	return OutboundConfig{
		Tag:            "proxy",
		Protocol:       string(node.Type),
		Settings:       settings,
		StreamSettings: streamSettings,
	}
}

// generateVLESSSettings 生成VLESS协议设置
func (g *ConfigGenerator) generateVLESSSettings(node *model.Node) json.RawMessage {
	vnext := []map[string]interface{}{
		{
			"address": node.Server,
			"port":    node.Port,
			"users": []map[string]interface{}{
				{
					"id":         node.UUID,
					"flow":       node.Flow,
					"encryption": "none",
				},
			},
		},
	}

	settings, _ := json.Marshal(map[string]interface{}{
		"vnext": vnext,
	})
	return settings
}

// generateVMessSettings 生成VMess协议设置
func (g *ConfigGenerator) generateVMessSettings(node *model.Node) json.RawMessage {
	// v26.2.6建议AlterID设为0
	alterID := node.AlterID
	if alterID == 0 {
		alterID = 0 // 使用AEAD加密
	}

	vnext := []map[string]interface{}{
		{
			"address": node.Server,
			"port":    node.Port,
			"users": []map[string]interface{}{
				{
					"id":       node.UUID,
					"alterId":  alterID,
					"security": "auto",
				},
			},
		},
	}

	settings, _ := json.Marshal(map[string]interface{}{
		"vnext": vnext,
	})
	return settings
}

// generateTrojanSettings 生成Trojan协议设置
func (g *ConfigGenerator) generateTrojanSettings(node *model.Node) json.RawMessage {
	servers := []map[string]interface{}{
		{
			"address":  node.Server,
			"port":     node.Port,
			"password": node.Password,
		},
	}

	settings, _ := json.Marshal(map[string]interface{}{
		"servers": servers,
	})
	return settings
}

// generateShadowsocksSettings 生成Shadowsocks协议设置
func (g *ConfigGenerator) generateShadowsocksSettings(node *model.Node) json.RawMessage {
	servers := []map[string]interface{}{
		{
			"address":  node.Server,
			"port":     node.Port,
			"method":   node.Method,
			"password": node.Password,
		},
	}

	settings, _ := json.Marshal(map[string]interface{}{
		"servers": servers,
	})
	return settings
}

// generateStreamSettings 生成传输层配置
// 支持WebSocket、gRPC、HTTP/2、XHTTP等传输方式
func (g *ConfigGenerator) generateStreamSettings(node *model.Node) json.RawMessage {
	stream := map[string]interface{}{
		"network": node.Network,
	}

	// 根据网络类型配置传输层
	switch node.Network {
	case string(model.NetworkWS):
		stream["wsSettings"] = g.generateWSSettings(node)
	case string(model.NetworkGRPC):
		stream["grpcSettings"] = g.generateGRPCSettings(node)
	case string(model.NetworkHTTP2):
		stream["httpSettings"] = g.generateHTTP2Settings(node)
	case string(model.NetworkXHTTP):
		stream["xhttpSettings"] = g.generateXHTTPSettings(node)
	case string(model.NetworkKCP):
		stream["kcpSettings"] = g.generateKCPSettings(node)
	}

	// 配置安全层
	if node.Security == string(model.SecurityTLS) {
		stream["security"] = "tls"
		stream["tlsSettings"] = g.generateTLSSettings(node)
	} else if node.Security == string(model.SecurityReality) {
		stream["security"] = "reality"
		stream["realitySettings"] = g.generateRealitySettings(node)
	}

	settings, _ := json.Marshal(stream)
	return settings
}

// generateWSSettings 生成WebSocket传输配置
func (g *ConfigGenerator) generateWSSettings(node *model.Node) map[string]interface{} {
	wsSettings := map[string]interface{}{
		"path": node.Path,
	}

	if node.Host != "" {
		wsSettings["headers"] = map[string]interface{}{
			"Host": node.Host,
		}
	}

	return wsSettings
}

// generateGRPCSettings 生成gRPC传输配置
func (g *ConfigGenerator) generateGRPCSettings(node *model.Node) map[string]interface{} {
	grpcSettings := map[string]interface{}{
		"serviceName": node.ServiceName,
	}

	if node.Host != "" {
		grpcSettings["authority"] = node.Host
	}

	return grpcSettings
}

// generateHTTP2Settings 生成HTTP/2传输配置
func (g *ConfigGenerator) generateHTTP2Settings(node *model.Node) map[string]interface{} {
	h2Settings := map[string]interface{}{
		"path": node.Path,
		"host": []string{node.Host},
	}

	return h2Settings
}

// generateXHTTPSettings 生成XHTTP传输配置（Xray-core v26.2.6新增）
func (g *ConfigGenerator) generateXHTTPSettings(node *model.Node) map[string]interface{} {
	xhttpSettings := map[string]interface{}{
		"mode": "stream", // 默认stream模式
	}

	if node.XHTTPConfig != nil {
		if node.XHTTPConfig.Mode != "" {
			xhttpSettings["mode"] = node.XHTTPConfig.Mode
		}

		// CDN绕过选项（v26.2.6新增）
		if node.XHTTPConfig.EnableCDNBypass {
			xhttpSettings["enableCdnBypass"] = true
		}

		// 下载设置
		if node.XHTTPConfig.DownloadSettings != nil {
			xhttpSettings["downloadSettings"] = map[string]interface{}{
				"address": node.XHTTPConfig.DownloadSettings.Address,
				"port":    node.XHTTPConfig.DownloadSettings.Port,
			}
		}
	}

	return xhttpSettings
}

// generateKCPSettings 生成KCP传输配置
func (g *ConfigGenerator) generateKCPSettings(node *model.Node) map[string]interface{} {
	kcpSettings := map[string]interface{}{
		"header": map[string]interface{}{
			"type": "none",
		},
		"mtu":              1350,
		"tti":              20,
		"uplinkCapacity":   5,
		"downlinkCapacity": 20,
		"congestion":       false,
		"readBufferSize":   1,
		"writeBufferSize":  1,
	}

	return kcpSettings
}

// generateTLSSettings 生成TLS配置（Xray-core v26.2.6重要变更）
// 移除allowInsecure，使用pinnedPeerCertSha256和verifyPeerCertByName
func (g *ConfigGenerator) generateTLSSettings(node *model.Node) map[string]interface{} {
	tlsSettings := map[string]interface{}{
		"serverName":    node.SNI,
		"allowInsecure": false, // v26.2.6强制为false
	}

	// v26.2.6新增：使用pinnedPeerCertSha256替代allowInsecure
	if node.PinnedPeerCertSha256 != "" {
		tlsSettings["pinnedPeerCertSha256"] = node.PinnedPeerCertSha256
	}

	// v26.2.6新增：证书名称验证
	if node.VerifyPeerCertByName != "" {
		tlsSettings["verifyPeerCertByName"] = node.VerifyPeerCertByName
	}

	// TLS指纹（v26.2.6使用动态Chrome）
	if node.Fingerprint != "" {
		tlsSettings["fingerprint"] = node.Fingerprint
	} else {
		tlsSettings["fingerprint"] = "chrome" // 默认使用Chrome指纹
	}

	// ALPN
	if len(node.Alpn) > 0 {
		tlsSettings["alpn"] = node.Alpn
	}

	return tlsSettings
}

// generateRealitySettings 生成REALITY配置
func (g *ConfigGenerator) generateRealitySettings(node *model.Node) map[string]interface{} {
	realitySettings := map[string]interface{}{
		"serverName":  node.SNI,
		"publicKey":   node.RealityPublicKey,
		"shortId":     node.RealityShortID,
		"fingerprint": node.RealityFingerprint,
	}

	// SpiderX参数
	if node.RealitySpiderX != "" {
		realitySettings["spiderX"] = node.RealitySpiderX
	}

	return realitySettings
}

// generateDirectOutbound 生成直连出站配置
func (g *ConfigGenerator) generateDirectOutbound() OutboundConfig {
	return OutboundConfig{
		Tag:      "direct",
		Protocol: "freedom",
	}
}

// generateRouting 生成路由配置
func (g *ConfigGenerator) generateRouting() *RoutingConfig {
	return &RoutingConfig{
		DomainStrategy: "IPIfNonMatch",
		Rules: []RuleConfig{
			{
				Type:        "field",
				OutboundTag: "proxy",
				InboundTag:  []string{"socks-in", "http-in"},
			},
		},
	}
}

// generateDNS 生成DNS配置
func (g *ConfigGenerator) generateDNS() *DNSConfig {
	return &DNSConfig{
		Tag: "dns-in",
		Servers: []DNSServer{
			{
				Address: "https://dns.google/dns-query",
			},
			{
				Address: "1.1.1.1",
			},
			{
				Address: "8.8.8.8",
			},
		},
	}
}

// ToJSON 将配置转换为JSON字符串
func (c *XrayConfig) ToJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化配置失败: %w", err)
	}
	return string(data), nil
}
