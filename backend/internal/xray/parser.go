package xray

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"proxy_server/internal/model"
	"proxy_server/pkg/logger"
)

// SubscriptionParser 订阅解析器
// 支持多种订阅格式，参考subconverter项目
type SubscriptionParser struct{}

// NewSubscriptionParser 创建订阅解析器
func NewSubscriptionParser() *SubscriptionParser {
	return &SubscriptionParser{}
}

// Parse 解析订阅内容
// 支持多种格式：
// - Base64编码的节点列表
// - Clash配置文件
// - Quantumult/Quantumult X格式
// - Surge配置文件
// - Surfboard配置文件
// - SSD格式
func (p *SubscriptionParser) Parse(content string) ([]*model.Node, error) {
	// 首先尝试解析为Clash配置
	if nodes, err := p.parseClashConfig(content); err == nil && len(nodes) > 0 {
		logger.Info("成功解析Clash配置，节点数: %d", len(nodes))
		return nodes, nil
	}

	// 尝试base64解码
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		decoded, err = base64.RawURLEncoding.DecodeString(content)
		if err != nil {
			// 可能是明文，直接使用
			decoded = []byte(content)
		}
	}

	decodedStr := string(decoded)

	// 尝试解析为Surge配置
	if nodes, err := p.parseSurgeConfig(decodedStr); err == nil && len(nodes) > 0 {
		logger.Info("成功解析Surge配置，节点数: %d", len(nodes))
		return nodes, nil
	}

	// 尝试解析为Quantumult格式
	if nodes, err := p.parseQuantumultConfig(decodedStr); err == nil && len(nodes) > 0 {
		logger.Info("成功解析Quantumult配置，节点数: %d", len(nodes))
		return nodes, nil
	}

	// 尝试解析为SSD格式
	if nodes, err := p.parseSSDConfig(decodedStr); err == nil && len(nodes) > 0 {
		logger.Info("成功解析SSD配置，节点数: %d", len(nodes))
		return nodes, nil
	}

	// 按行分割，解析节点链接
	lines := strings.Split(decodedStr, "\n")
	var nodes []*model.Node

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		node, err := p.parseLine(line)
		if err != nil {
			logger.Debug("解析行失败: %v", err)
			continue
		}

		if node != nil {
			nodes = append(nodes, node)
		}
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("未解析到有效节点")
	}

	logger.Info("成功解析节点链接，节点数: %d", len(nodes))
	return nodes, nil
}

// parseClashConfig 解析Clash配置文件
func (p *SubscriptionParser) parseClashConfig(content string) ([]*model.Node, error) {
	// 检查是否是Clash配置格式
	if !strings.Contains(content, "proxies:") && !strings.Contains(content, "\"proxies\"") {
		return nil, fmt.Errorf("不是Clash配置格式")
	}

	var clashConfig struct {
		Proxies []map[string]interface{} `json:"proxies"`
	}

	// 尝试YAML解析（简化处理，使用JSON解析）
	// 实际项目中应该使用yaml库
	if err := json.Unmarshal([]byte(content), &clashConfig); err != nil {
		// 尝试简单的正则匹配
		return p.parseClashProxiesSimple(content)
	}

	if len(clashConfig.Proxies) == 0 {
		return nil, fmt.Errorf("Clash配置中没有节点")
	}

	var nodes []*model.Node
	for _, proxy := range clashConfig.Proxies {
		node, err := p.parseClashProxy(proxy)
		if err != nil {
			continue
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

// parseClashProxiesSimple 简单解析Clash配置中的节点
func (p *SubscriptionParser) parseClashProxiesSimple(content string) ([]*model.Node, error) {
	// 使用正则匹配节点定义
	// 匹配格式: - {name: xxx, type: xxx, server: xxx, port: xxx, ...}
	re := regexp.MustCompile(`(?s)-\s*\{[^}]*name:\s*([^,}]+)[^}]*type:\s*([^,}]+)[^}]*server:\s*([^,}]+)[^}]*port:\s*(\d+)[^}]*\}`)
	matches := re.FindAllStringSubmatch(content, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("未找到Clash节点")
	}

	var nodes []*model.Node
	for _, match := range matches {
		if len(match) < 5 {
			continue
		}

		name := strings.TrimSpace(match[1])
		proxyType := strings.TrimSpace(match[2])
		server := strings.TrimSpace(match[3])
		port, _ := strconv.Atoi(strings.TrimSpace(match[4]))

		node := &model.Node{
			ID:        generateID(),
			Name:      name,
			Server:    server,
			Port:      port,
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		switch proxyType {
		case "ss":
			node.Type = model.NodeTypeShadowsocks
		case "ssr":
			node.Type = model.NodeTypeSSR
		case "vmess":
			node.Type = model.NodeTypeVMess
		case "trojan":
			node.Type = model.NodeTypeTrojan
		case "vless":
			node.Type = model.NodeTypeVLESS
		default:
			continue
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// parseClashProxy 解析单个Clash代理配置
func (p *SubscriptionParser) parseClashProxy(proxy map[string]interface{}) (*model.Node, error) {
	proxyType, ok := proxy["type"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少type字段")
	}

	name, _ := proxy["name"].(string)
	server, _ := proxy["server"].(string)
	port := 0
	switch v := proxy["port"].(type) {
	case float64:
		port = int(v)
	case int:
		port = v
	case string:
		port, _ = strconv.Atoi(v)
	}

	if server == "" || port == 0 {
		return nil, fmt.Errorf("缺少server或port字段")
	}

	node := &model.Node{
		ID:        generateID(),
		Name:      name,
		Server:    server,
		Port:      port,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	switch proxyType {
	case "ss":
		node.Type = model.NodeTypeShadowsocks
		node.Method, _ = proxy["cipher"].(string)
		if node.Method == "" {
			node.Method, _ = proxy["method"].(string)
		}
		node.Password, _ = proxy["password"].(string)

	case "ssr":
		node.Type = model.NodeTypeSSR
		node.Method, _ = proxy["cipher"].(string)
		if node.Method == "" {
			node.Method, _ = proxy["method"].(string)
		}
		node.Password, _ = proxy["password"].(string)

	case "vmess":
		node.Type = model.NodeTypeVMess
		node.UUID, _ = proxy["uuid"].(string)
		node.AlterID = 0
		if aid, ok := proxy["alterId"]; ok {
			switch v := aid.(type) {
			case float64:
				node.AlterID = int(v)
			case int:
				node.AlterID = v
			}
		}
		node.Network, _ = proxy["network"].(string)
		if node.Network == "" {
			node.Network = "tcp"
		}
		node.Security, _ = proxy["tls"].(string)
		if node.Security == "" {
			if tls, ok := proxy["tls"].(bool); ok && tls {
				node.Security = "tls"
			}
		}

	case "trojan":
		node.Type = model.NodeTypeTrojan
		node.Password, _ = proxy["password"].(string)
		node.Network, _ = proxy["network"].(string)
		if node.Network == "" {
			node.Network = "tcp"
		}
		node.SNI, _ = proxy["sni"].(string)

	case "vless":
		node.Type = model.NodeTypeVLESS
		node.UUID, _ = proxy["uuid"].(string)
		node.Network, _ = proxy["network"].(string)
		if node.Network == "" {
			node.Network = "tcp"
		}
		node.Security, _ = proxy["tls"].(string)
		node.Flow, _ = proxy["flow"].(string)

	default:
		return nil, fmt.Errorf("不支持的代理类型: %s", proxyType)
	}

	return node, nil
}

// parseSurgeConfig 解析Surge配置文件
func (p *SubscriptionParser) parseSurgeConfig(content string) ([]*model.Node, error) {
	// 检查是否是Surge配置格式
	if !strings.Contains(content, "[Proxy]") && !strings.Contains(content, "[Proxy]") {
		return nil, fmt.Errorf("不是Surge配置格式")
	}

	var nodes []*model.Node
	lines := strings.Split(content, "\n")
	inProxySection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "[Proxy]" {
			inProxySection = true
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inProxySection = false
			continue
		}

		if !inProxySection || line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Surge格式: name = type, server, port, ...
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		config := strings.TrimSpace(parts[1])
		configParts := strings.Split(config, ",")

		if len(configParts) < 3 {
			continue
		}

		proxyType := strings.TrimSpace(configParts[0])
		server := strings.TrimSpace(configParts[1])
		port, _ := strconv.Atoi(strings.TrimSpace(configParts[2]))

		node := &model.Node{
			ID:        generateID(),
			Name:      name,
			Server:    server,
			Port:      port,
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		switch proxyType {
		case "ss":
			node.Type = model.NodeTypeShadowsocks
			if len(configParts) > 3 {
				node.Method = strings.TrimSpace(configParts[3])
			}
			if len(configParts) > 4 {
				node.Password = strings.TrimSpace(configParts[4])
			}
		case "vmess":
			node.Type = model.NodeTypeVMess
			// 解析VMess特定参数
		case "trojan":
			node.Type = model.NodeTypeTrojan
			if len(configParts) > 3 {
				node.Password = strings.TrimSpace(configParts[3])
			}
		default:
			continue
		}

		nodes = append(nodes, node)
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("未找到Surge节点")
	}

	return nodes, nil
}

// parseQuantumultConfig 解析Quantumult配置
func (p *SubscriptionParser) parseQuantumultConfig(content string) ([]*model.Node, error) {
	var nodes []*model.Node
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Quantumult格式: name=type,server,port,...
		if strings.Contains(line, "=") && strings.Contains(line, ",") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			name := strings.TrimSpace(parts[0])
			config := strings.TrimSpace(parts[1])
			configParts := strings.Split(config, ",")

			if len(configParts) < 3 {
				continue
			}

			proxyType := strings.TrimSpace(configParts[0])
			server := strings.TrimSpace(configParts[1])
			port, _ := strconv.Atoi(strings.TrimSpace(configParts[2]))

			node := &model.Node{
				ID:        generateID(),
				Name:      name,
				Server:    server,
				Port:      port,
				Enabled:   true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			switch proxyType {
			case "shadowsocks", "ss":
				node.Type = model.NodeTypeShadowsocks
				if len(configParts) > 3 {
					node.Method = strings.TrimSpace(configParts[3])
				}
				if len(configParts) > 4 {
					node.Password = strings.TrimSpace(configParts[4])
				}
			case "vmess":
				node.Type = model.NodeTypeVMess
			case "trojan":
				node.Type = model.NodeTypeTrojan
			default:
				continue
			}

			nodes = append(nodes, node)
		}
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("未找到Quantumult节点")
	}

	return nodes, nil
}

// parseSSDConfig 解析SSD配置
func (p *SubscriptionParser) parseSSDConfig(content string) ([]*model.Node, error) {
	// SSD格式通常是base64编码的JSON
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		decoded, err = base64.RawURLEncoding.DecodeString(content)
		if err != nil {
			return nil, fmt.Errorf("SSD解码失败")
		}
	}

	var ssdConfig struct {
		Servers []struct {
			Server   string `json:"server"`
			Port     int    `json:"port"`
			Method   string `json:"encryption"`
			Password string `json:"password"`
			Remarks  string `json:"remarks"`
		} `json:"servers"`
	}

	if err := json.Unmarshal(decoded, &ssdConfig); err != nil {
		return nil, fmt.Errorf("SSD JSON解析失败")
	}

	if len(ssdConfig.Servers) == 0 {
		return nil, fmt.Errorf("SSD配置中没有服务器")
	}

	var nodes []*model.Node
	for _, server := range ssdConfig.Servers {
		node := &model.Node{
			ID:        generateID(),
			Name:      server.Remarks,
			Type:      model.NodeTypeShadowsocks,
			Server:    server.Server,
			Port:      server.Port,
			Method:    server.Method,
			Password:  server.Password,
			Network:   "tcp",
			Security:  "none",
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// parseLine 解析单行订阅内容
func (p *SubscriptionParser) parseLine(line string) (*model.Node, error) {
	// 判断协议类型
	if strings.HasPrefix(line, "vless://") {
		return p.parseVLESS(line)
	} else if strings.HasPrefix(line, "vmess://") {
		return p.parseVMess(line)
	} else if strings.HasPrefix(line, "trojan://") {
		return p.parseTrojan(line)
	} else if strings.HasPrefix(line, "ss://") {
		return p.parseShadowsocks(line)
	} else if strings.HasPrefix(line, "ssr://") {
		return p.parseSSR(line)
	}

	// 过滤掉分类节点和其他非代理节点
	ignoredProtocols := []string{
		"selector://",
		"urltest://",
		"fallback://",
		"loadbalance://",
		"shadowsocksr://",
		"http://",
		"https://",
		"socks://",
		"socks5://",
	}
	
	for _, ignored := range ignoredProtocols {
		if strings.HasPrefix(line, ignored) {
			return nil, fmt.Errorf("忽略非代理节点: %s", ignored)
		}
	}

	return nil, fmt.Errorf("不支持的协议: %s", line[:min(20, len(line))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseVLESS 解析VLESS链接
func (p *SubscriptionParser) parseVLESS(link string) (*model.Node, error) {
	link = strings.TrimPrefix(link, "vless://")

	parts := strings.SplitN(link, "#", 2)
	mainPart := parts[0]
	name := ""
	if len(parts) > 1 {
		name, _ = url.QueryUnescape(parts[1])
	}

	paramParts := strings.SplitN(mainPart, "?", 2)
	userServer := paramParts[0]
	params := ""
	if len(paramParts) > 1 {
		params = paramParts[1]
	}

	userServerParts := strings.SplitN(userServer, "@", 2)
	if len(userServerParts) != 2 {
		return nil, fmt.Errorf("无效的VLESS链接格式")
	}
	uuid := userServerParts[0]
	serverPort := userServerParts[1]

	lastColon := strings.LastIndex(serverPort, ":")
	if lastColon == -1 {
		return nil, fmt.Errorf("无效的服务器地址")
	}
	server := serverPort[:lastColon]
	port, err := strconv.Atoi(serverPort[lastColon+1:])
	if err != nil {
		return nil, fmt.Errorf("无效的端口: %w", err)
	}

	node := &model.Node{
		ID:        generateID(),
		Name:      name,
		Type:      model.NodeTypeVLESS,
		Server:    server,
		Port:      port,
		UUID:      uuid,
		Network:   "tcp",
		Security:  "none",
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if params != "" {
		query, err := url.ParseQuery(params)
		if err == nil {
			if network := query.Get("type"); network != "" {
				node.Network = network
			}
			if security := query.Get("security"); security != "" {
				node.Security = security
			}
			if sni := query.Get("sni"); sni != "" {
				node.SNI = sni
			}
			if host := query.Get("host"); host != "" {
				node.Host = host
			}
			if path := query.Get("path"); path != "" {
				node.Path = path
			}
			if flow := query.Get("flow"); flow != "" {
				node.Flow = flow
			}
			if pbk := query.Get("pbk"); pbk != "" {
				node.RealityPublicKey = pbk
			}
			if sid := query.Get("sid"); sid != "" {
				node.RealityShortID = sid
			}
			if fp := query.Get("fp"); fp != "" {
				node.RealityFingerprint = fp
			}
			if alpn := query.Get("alpn"); alpn != "" {
				node.Alpn = strings.Split(alpn, ",")
			}
			if serviceName := query.Get("serviceName"); serviceName != "" {
				node.ServiceName = serviceName
			}
		}
	}

	return node, nil
}

// parseVMess 解析VMess链接
func (p *SubscriptionParser) parseVMess(link string) (*model.Node, error) {
	link = strings.TrimPrefix(link, "vmess://")

	decoded, err := base64.StdEncoding.DecodeString(link)
	if err != nil {
		decoded, err = base64.RawURLEncoding.DecodeString(link)
		if err != nil {
			return nil, fmt.Errorf("VMess链接解码失败: %w", err)
		}
	}

	var vmessData struct {
		V    string `json:"v"`
		Ps   string `json:"ps"`
		Add  string `json:"add"`
		Port any    `json:"port"`
		ID   string `json:"id"`
		Aid  any    `json:"aid"`
		Net  string `json:"net"`
		Type string `json:"type"`
		Host string `json:"host"`
		Path string `json:"path"`
		TLS  string `json:"tls"`
		Sni  string `json:"sni"`
		Alpn string `json:"alpn"`
		Fp   string `json:"fp"`
	}

	if err := json.Unmarshal(decoded, &vmessData); err != nil {
		return nil, fmt.Errorf("VMess JSON解析失败: %w", err)
	}

	var port int
	switch v := vmessData.Port.(type) {
	case float64:
		port = int(v)
	case string:
		port, _ = strconv.Atoi(v)
	}

	var alterID int
	switch v := vmessData.Aid.(type) {
	case float64:
		alterID = int(v)
	case string:
		alterID, _ = strconv.Atoi(v)
	}

	node := &model.Node{
		ID:        generateID(),
		Name:      vmessData.Ps,
		Type:      model.NodeTypeVMess,
		Server:    vmessData.Add,
		Port:      port,
		UUID:      vmessData.ID,
		AlterID:   alterID,
		Network:   vmessData.Net,
		Security:  vmessData.TLS,
		SNI:       vmessData.Sni,
		Host:      vmessData.Host,
		Path:      vmessData.Path,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if vmessData.Alpn != "" {
		node.Alpn = strings.Split(vmessData.Alpn, ",")
	}

	if vmessData.Fp != "" {
		node.Fingerprint = vmessData.Fp
	}

	return node, nil
}

// parseTrojan 解析Trojan链接
func (p *SubscriptionParser) parseTrojan(link string) (*model.Node, error) {
	link = strings.TrimPrefix(link, "trojan://")

	parts := strings.SplitN(link, "#", 2)
	mainPart := parts[0]
	name := ""
	if len(parts) > 1 {
		name, _ = url.QueryUnescape(parts[1])
	}

	paramParts := strings.SplitN(mainPart, "?", 2)
	userServer := paramParts[0]
	params := ""
	if len(paramParts) > 1 {
		params = paramParts[1]
	}

	atIndex := strings.Index(userServer, "@")
	if atIndex == -1 {
		return nil, fmt.Errorf("无效的Trojan链接格式")
	}
	password := userServer[:atIndex]
	serverPort := userServer[atIndex+1:]

	lastColon := strings.LastIndex(serverPort, ":")
	if lastColon == -1 {
		return nil, fmt.Errorf("无效的服务器地址")
	}
	server := serverPort[:lastColon]
	port, err := strconv.Atoi(serverPort[lastColon+1:])
	if err != nil {
		return nil, fmt.Errorf("无效的端口: %w", err)
	}

	node := &model.Node{
		ID:        generateID(),
		Name:      name,
		Type:      model.NodeTypeTrojan,
		Server:    server,
		Port:      port,
		Password:  password,
		Network:   "tcp",
		Security:  "tls",
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if params != "" {
		query, err := url.ParseQuery(params)
		if err == nil {
			if network := query.Get("type"); network != "" {
				node.Network = network
			}
			if sni := query.Get("sni"); sni != "" {
				node.SNI = sni
			}
			if host := query.Get("host"); host != "" {
				node.Host = host
			}
			if path := query.Get("path"); path != "" {
				node.Path = path
			}
		}
	}

	return node, nil
}

// parseShadowsocks 解析Shadowsocks链接
func (p *SubscriptionParser) parseShadowsocks(link string) (*model.Node, error) {
	link = strings.TrimPrefix(link, "ss://")

	parts := strings.SplitN(link, "#", 2)
	mainPart := parts[0]
	name := ""
	if len(parts) > 1 {
		name, _ = url.QueryUnescape(parts[1])
	}

	var method, password, server string
	var port int

	if atIndex := strings.Index(mainPart, "@"); atIndex != -1 {
		userPart := mainPart[:atIndex]
		serverPart := mainPart[atIndex+1:]

		decoded, err := base64.StdEncoding.DecodeString(userPart)
		if err != nil {
			decoded, err = base64.RawURLEncoding.DecodeString(userPart)
			if err != nil {
				decoded = []byte(userPart)
			}
		}

		colonIndex := strings.Index(string(decoded), ":")
		if colonIndex != -1 {
			method = string(decoded[:colonIndex])
			password = string(decoded[colonIndex+1:])
		}

		lastColon := strings.LastIndex(serverPart, ":")
		if lastColon != -1 {
			server = serverPart[:lastColon]
			port, _ = strconv.Atoi(serverPart[lastColon+1:])
		}
	} else {
		decoded, err := base64.StdEncoding.DecodeString(mainPart)
		if err != nil {
			decoded, err = base64.RawURLEncoding.DecodeString(mainPart)
			if err != nil {
				return nil, fmt.Errorf("Shadowsocks链接解码失败: %w", err)
			}
		}

		re := regexp.MustCompile(`^(.+?):(.+)@(.+):(\d+)$`)
		matches := re.FindStringSubmatch(string(decoded))
		if len(matches) == 5 {
			method = matches[1]
			password = matches[2]
			server = matches[3]
			port, _ = strconv.Atoi(matches[4])
		}
	}

	if method == "" || password == "" || server == "" || port == 0 {
		return nil, fmt.Errorf("无效的Shadowsocks链接格式")
	}

	return &model.Node{
		ID:        generateID(),
		Name:      name,
		Type:      model.NodeTypeShadowsocks,
		Server:    server,
		Port:      port,
		Method:    method,
		Password:  password,
		Network:   "tcp",
		Security:  "none",
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// parseSSR 解析ShadowsocksR链接
func (p *SubscriptionParser) parseSSR(link string) (*model.Node, error) {
	link = strings.TrimPrefix(link, "ssr://")

	decoded, err := base64.StdEncoding.DecodeString(link)
	if err != nil {
		decoded, err = base64.RawURLEncoding.DecodeString(link)
		if err != nil {
			return nil, fmt.Errorf("SSR链接解码失败: %w", err)
		}
	}

	parts := strings.Split(string(decoded), "/?")
	if len(parts) < 1 {
		return nil, fmt.Errorf("无效的SSR链接格式")
	}

	mainPart := parts[0]
	mainParts := strings.Split(mainPart, ":")
	if len(mainParts) < 6 {
		return nil, fmt.Errorf("无效的SSR链接格式")
	}

	server := mainParts[0]
	port, _ := strconv.Atoi(mainParts[1])
	protocol := mainParts[2]
	method := mainParts[3]
	obfs := mainParts[4]
	passwordEncoded := mainParts[5]

	passwordDecoded, _ := base64.StdEncoding.DecodeString(passwordEncoded)
	password := string(passwordDecoded)

	name := ""
	if len(parts) > 1 {
		params, _ := url.ParseQuery(parts[1])
		if remarks := params.Get("remarks"); remarks != "" {
			nameDecoded, _ := base64.StdEncoding.DecodeString(remarks)
			name = string(nameDecoded)
		}
	}

	return &model.Node{
		ID:        generateID(),
		Name:      name,
		Type:      model.NodeTypeSSR,
		Server:    server,
		Port:      port,
		Method:    method,
		Password:  password,
		Network:   "tcp",
		Security:  "none",
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Headers: map[string]string{
			"protocol": protocol,
			"obfs":     obfs,
		},
	}, nil
}

// generateID 生成唯一ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
