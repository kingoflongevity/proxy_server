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
)

// SubscriptionParser 订阅解析器
type SubscriptionParser struct{}

// NewSubscriptionParser 创建订阅解析器
func NewSubscriptionParser() *SubscriptionParser {
	return &SubscriptionParser{}
}

// Parse 解析订阅内容
// 参数：
//   - content: 订阅内容（通常是base64编码）
// 返回：
//   - []*model.Node: 解析出的节点列表
//   - error: 错误信息
func (p *SubscriptionParser) Parse(content string) ([]*model.Node, error) {
	// 尝试base64解码
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		// 如果不是base64，尝试URL安全的base64
		decoded, err = base64.RawURLEncoding.DecodeString(content)
		if err != nil {
			// 可能是明文，直接使用
			decoded = []byte(content)
		}
	}

	// 按行分割
	lines := strings.Split(string(decoded), "\n")
	var nodes []*model.Node

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		node, err := p.parseLine(line)
		if err != nil {
			continue
		}

		if node != nil {
			nodes = append(nodes, node)
		}
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("未解析到有效节点")
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

	return nil, fmt.Errorf("不支持的协议: %s", line[:20])
}

// parseVLESS 解析VLESS链接
// 格式: vless://uuid@server:port?参数#名称
func (p *SubscriptionParser) parseVLESS(link string) (*model.Node, error) {
	// 移除协议前缀
	link = strings.TrimPrefix(link, "vless://")

	// 分离名称
	parts := strings.SplitN(link, "#", 2)
	mainPart := parts[0]
	name := ""
	if len(parts) > 1 {
		name, _ = url.QueryUnescape(parts[1])
	}

	// 分离参数
	paramParts := strings.SplitN(mainPart, "?", 2)
	userServer := paramParts[0]
	params := ""
	if len(paramParts) > 1 {
		params = paramParts[1]
	}

	// 解析uuid@server:port
	userServerParts := strings.SplitN(userServer, "@", 2)
	if len(userServerParts) != 2 {
		return nil, fmt.Errorf("无效的VLESS链接格式")
	}
	uuid := userServerParts[0]
	serverPort := userServerParts[1]

	// 解析server:port
	lastColon := strings.LastIndex(serverPort, ":")
	if lastColon == -1 {
		return nil, fmt.Errorf("无效的服务器地址")
	}
	server := serverPort[:lastColon]
	port, err := strconv.Atoi(serverPort[lastColon+1:])
	if err != nil {
		return nil, fmt.Errorf("无效的端口: %w", err)
	}

	// 解析参数
	node := &model.Node{
		ID:         generateID(),
		Name:       name,
		Type:       model.NodeTypeVLESS,
		Server:     server,
		Port:       port,
		UUID:       uuid,
		Network:    "tcp",
		Security:   "none",
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if params != "" {
		query, err := url.ParseQuery(params)
		if err == nil {
			// 传输层类型
			if network := query.Get("type"); network != "" {
				node.Network = network
			}

			// 安全类型
			if security := query.Get("security"); security != "" {
				node.Security = security
			}

			// TLS参数
			if sni := query.Get("sni"); sni != "" {
				node.SNI = sni
			}
			if host := query.Get("host"); host != "" {
				node.Host = host
			}
			if path := query.Get("path"); path != "" {
				node.Path = path
			}

			// Flow控制
			if flow := query.Get("flow"); flow != "" {
				node.Flow = flow
			}

			// REALITY参数
			if pbk := query.Get("pbk"); pbk != "" {
				node.RealityPublicKey = pbk
			}
			if sid := query.Get("sid"); sid != "" {
				node.RealityShortID = sid
			}
			if fp := query.Get("fp"); fp != "" {
				node.RealityFingerprint = fp
			}
			if spx := query.Get("spx"); spx != "" {
				node.RealitySpiderX = spx
			}

			// XHTTP参数 (v26.2.6)
			if xhttpMode := query.Get("xhttpMode"); xhttpMode != "" {
				node.XHTTPConfig = &model.XHTTPConfig{
					Mode: xhttpMode,
				}
			}

			// ALPN
			if alpn := query.Get("alpn"); alpn != "" {
				node.Alpn = strings.Split(alpn, ",")
			}

			// 服务名称 (gRPC)
			if serviceName := query.Get("serviceName"); serviceName != "" {
				node.ServiceName = serviceName
			}

			// v26.2.6: pinnedPeerCertSha256
			if pcs := query.Get("pcs"); pcs != "" {
				node.PinnedPeerCertSha256 = pcs
			}

			// v26.2.6: verifyPeerCertByName
			if vcn := query.Get("vcn"); vcn != "" {
				node.VerifyPeerCertByName = vcn
			}
		}
	}

	return node, nil
}

// parseVMess 解析VMess链接
// 格式: vmess://base64编码的JSON
func (p *SubscriptionParser) parseVMess(link string) (*model.Node, error) {
	link = strings.TrimPrefix(link, "vmess://")

	// Base64解码
	decoded, err := base64.StdEncoding.DecodeString(link)
	if err != nil {
		decoded, err = base64.RawURLEncoding.DecodeString(link)
		if err != nil {
			return nil, fmt.Errorf("VMess链接解码失败: %w", err)
		}
	}

	// 解析JSON
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

	// 处理端口（可能是字符串或数字）
	var port int
	switch v := vmessData.Port.(type) {
	case float64:
		port = int(v)
	case string:
		port, _ = strconv.Atoi(v)
	}

	// 处理AlterID
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
// 格式: trojan://password@server:port?参数#名称
func (p *SubscriptionParser) parseTrojan(link string) (*model.Node, error) {
	link = strings.TrimPrefix(link, "trojan://")

	// 分离名称
	parts := strings.SplitN(link, "#", 2)
	mainPart := parts[0]
	name := ""
	if len(parts) > 1 {
		name, _ = url.QueryUnescape(parts[1])
	}

	// 分离参数
	paramParts := strings.SplitN(mainPart, "?", 2)
	userServer := paramParts[0]
	params := ""
	if len(paramParts) > 1 {
		params = paramParts[1]
	}

	// 解析password@server:port
	atIndex := strings.Index(userServer, "@")
	if atIndex == -1 {
		return nil, fmt.Errorf("无效的Trojan链接格式")
	}
	password := userServer[:atIndex]
	serverPort := userServer[atIndex+1:]

	// 解析server:port
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
// 格式: ss://base64(method:password)@server:port#名称
// 或: ss://base64(method:password@server:port)#名称
func (p *SubscriptionParser) parseShadowsocks(link string) (*model.Node, error) {
	link = strings.TrimPrefix(link, "ss://")

	// 分离名称
	parts := strings.SplitN(link, "#", 2)
	mainPart := parts[0]
	name := ""
	if len(parts) > 1 {
		name, _ = url.QueryUnescape(parts[1])
	}

	var method, password, server string
	var port int

	// 尝试解析SIP002格式: base64(method:password)@server:port
	if atIndex := strings.Index(mainPart, "@"); atIndex != -1 {
		userPart := mainPart[:atIndex]
		serverPart := mainPart[atIndex+1:]

		// 解码用户部分
		decoded, err := base64.StdEncoding.DecodeString(userPart)
		if err != nil {
			decoded, err = base64.RawURLEncoding.DecodeString(userPart)
			if err != nil {
				// 可能是URL编码的
				decoded = []byte(userPart)
			}
		}

		// 解析method:password
		colonIndex := strings.Index(string(decoded), ":")
		if colonIndex != -1 {
			method = string(decoded[:colonIndex])
			password = string(decoded[colonIndex+1:])
		}

		// 解析server:port
		lastColon := strings.LastIndex(serverPart, ":")
		if lastColon != -1 {
			server = serverPart[:lastColon]
			port, _ = strconv.Atoi(serverPart[lastColon+1:])
		}
	} else {
		// 旧格式: base64(method:password@server:port)
		decoded, err := base64.StdEncoding.DecodeString(mainPart)
		if err != nil {
			decoded, err = base64.RawURLEncoding.DecodeString(mainPart)
			if err != nil {
				return nil, fmt.Errorf("Shadowsocks链接解码失败: %w", err)
			}
		}

		// 解析method:password@server:port
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
// 格式: ssr://base64编码的内容
func (p *SubscriptionParser) parseSSR(link string) (*model.Node, error) {
	link = strings.TrimPrefix(link, "ssr://")

	decoded, err := base64.StdEncoding.DecodeString(link)
	if err != nil {
		decoded, err = base64.RawURLEncoding.DecodeString(link)
		if err != nil {
			return nil, fmt.Errorf("SSR链接解码失败: %w", err)
		}
	}

	// SSR格式: server:port:protocol:method:obfs:password_base64/?params
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

	// 解码密码
	passwordDecoded, _ := base64.StdEncoding.DecodeString(passwordEncoded)
	password := string(passwordDecoded)

	name := ""
	if len(parts) > 1 {
		// 解析参数
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
