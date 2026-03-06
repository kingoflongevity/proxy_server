package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"proxy_server/internal/model"
	"proxy_server/internal/repository"
	"proxy_server/pkg/errors"
	"proxy_server/pkg/logger"
	"proxy_server/pkg/utils"
)

// SubscriptionService 订阅服务接口
type SubscriptionService interface {
	Create(req *model.SubscriptionCreateRequest) (*model.Subscription, error)
	GetByID(id string) (*model.Subscription, error)
	GetAll() ([]*model.Subscription, error)
	Update(id string, req *model.SubscriptionUpdateRequest) (*model.Subscription, error)
	Delete(id string) error
	Refresh(id string) error
}

// subscriptionService 订阅服务实现
type subscriptionService struct {
	subRepo repository.SubscriptionRepository
	nodeRepo repository.NodeRepository
}

// NewSubscriptionService 创建订阅服务
func NewSubscriptionService(subRepo repository.SubscriptionRepository, nodeRepo repository.NodeRepository) SubscriptionService {
	return &subscriptionService{
		subRepo:  subRepo,
		nodeRepo: nodeRepo,
	}
}

// Create 创建订阅
func (s *subscriptionService) Create(req *model.SubscriptionCreateRequest) (*model.Subscription, error) {
	// 设置默认更新间隔
	if req.UpdateInterval == 0 {
		req.UpdateInterval = 24 // 默认24小时
	}
	
	subscription := &model.Subscription{
		ID:            utils.GenerateID(),
		Name:          req.Name,
		URL:           req.URL,
		AutoUpdate:    req.AutoUpdate,
		UpdateInterval: req.UpdateInterval,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	// 保存订阅
	if err := s.subRepo.Create(subscription); err != nil {
		logger.Error("创建订阅失败: %v", err)
		return nil, errors.NewError(errors.DataSaveError, err.Error())
	}
	
	// 立即刷新订阅节点
	if err := s.Refresh(subscription.ID); err != nil {
		logger.Warn("刷新订阅节点失败: %v", err)
	}
	
	logger.Info("创建订阅成功: %s", subscription.Name)
	return subscription, nil
}

// GetByID 根据ID获取订阅
func (s *subscriptionService) GetByID(id string) (*model.Subscription, error) {
	subscription, err := s.subRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	
	if subscription == nil {
		return nil, errors.NewError(errors.SubscriptionNotFound, "")
	}
	
	return subscription, nil
}

// GetAll 获取所有订阅
func (s *subscriptionService) GetAll() ([]*model.Subscription, error) {
	return s.subRepo.GetAll()
}

// Update 更新订阅
func (s *subscriptionService) Update(id string, req *model.SubscriptionUpdateRequest) (*model.Subscription, error) {
	subscription, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	
	// 更新字段
	if req.Name != "" {
		subscription.Name = req.Name
	}
	if req.URL != "" {
		subscription.URL = req.URL
	}
	subscription.AutoUpdate = req.AutoUpdate
	if req.UpdateInterval > 0 {
		subscription.UpdateInterval = req.UpdateInterval
	}
	subscription.UpdatedAt = time.Now()
	
	// 保存
	if err := s.subRepo.Update(subscription); err != nil {
		logger.Error("更新订阅失败: %v", err)
		return nil, errors.NewError(errors.DataSaveError, err.Error())
	}
	
	logger.Info("更新订阅成功: %s", subscription.Name)
	return subscription, nil
}

// Delete 删除订阅
func (s *subscriptionService) Delete(id string) error {
	// 检查订阅是否存在
	_, err := s.GetByID(id)
	if err != nil {
		return err
	}
	
	// 删除订阅的所有节点
	if err := s.nodeRepo.DeleteBySubscriptionID(id); err != nil {
		logger.Error("删除订阅节点失败: %v", err)
		return errors.NewError(errors.DataSaveError, err.Error())
	}
	
	// 删除订阅
	if err := s.subRepo.Delete(id); err != nil {
		logger.Error("删除订阅失败: %v", err)
		return errors.NewError(errors.DataSaveError, err.Error())
	}
	
	logger.Info("删除订阅成功: %s", id)
	return nil
}

// Refresh 刷新订阅节点
func (s *subscriptionService) Refresh(id string) error {
	subscription, err := s.GetByID(id)
	if err != nil {
		return err
	}
	
	// 获取订阅内容
	content, err := s.fetchSubscription(subscription.URL)
	if err != nil {
		logger.Error("获取订阅内容失败: %v", err)
		return errors.NewError(errors.SubscriptionFetchFailed, err.Error())
	}
	
	// 解析节点
	nodes, err := s.parseNodes(content, subscription.ID)
	if err != nil {
		logger.Error("解析订阅节点失败: %v", err)
		return errors.NewError(errors.SubscriptionParseFailed, err.Error())
	}
	
	// 删除旧节点
	if err := s.nodeRepo.DeleteBySubscriptionID(id); err != nil {
		logger.Error("删除旧节点失败: %v", err)
		return errors.NewError(errors.DataSaveError, err.Error())
	}
	
	// 保存新节点
	for _, node := range nodes {
		if err := s.nodeRepo.Create(node); err != nil {
			logger.Warn("保存节点失败: %s - %v", node.Name, err)
		}
	}
	
	// 更新订阅信息
	subscription.NodeCount = len(nodes)
	subscription.LastUpdate = time.Now()
	subscription.UpdatedAt = time.Now()
	if err := s.subRepo.Update(subscription); err != nil {
		logger.Error("更新订阅信息失败: %v", err)
	}
	
	logger.Info("刷新订阅成功: %s, 节点数: %d", subscription.Name, len(nodes))
	return nil
}

// fetchSubscription 获取订阅内容
func (s *subscriptionService) fetchSubscription(url string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	// 尝试Base64解码
	decoded, err := utils.Base64Decode(string(body))
	if err == nil {
		return string(decoded), nil
	}
	
	// 如果解码失败，直接返回原始内容
	return string(body), nil
}

// parseNodes 解析节点
func (s *subscriptionService) parseNodes(content string, subscriptionID string) ([]*model.Node, error) {
	var nodes []*model.Node
	
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		// 解析不同协议的节点
		node, err := s.parseNodeLine(line, subscriptionID)
		if err != nil {
			logger.Warn("解析节点失败: %s - %v", line, err)
			continue
		}
		
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	
	return nodes, scanner.Err()
}

// parseNodeLine 解析单行节点配置
func (s *subscriptionService) parseNodeLine(line string, subscriptionID string) (*model.Node, error) {
	// 支持的协议格式：
	// vmess://base64
	// vless://uuid@server:port?params#name
	// trojan://password@server:port?params#name
	// ss://base64@server:port#name
	
	if strings.HasPrefix(line, "vmess://") {
		return s.parseVMess(line, subscriptionID)
	} else if strings.HasPrefix(line, "vless://") {
		return s.parseVLESS(line, subscriptionID)
	} else if strings.HasPrefix(line, "trojan://") {
		return s.parseTrojan(line, subscriptionID)
	} else if strings.HasPrefix(line, "ss://") {
		return s.parseShadowsocks(line, subscriptionID)
	}
	
	return nil, fmt.Errorf("不支持的协议格式")
}

// parseVMess 解析VMess节点
func (s *subscriptionService) parseVMess(line string, subscriptionID string) (*model.Node, error) {
	// vmess://base64
	encoded := strings.TrimPrefix(line, "vmess://")
	decoded, err := utils.Base64Decode(encoded)
	if err != nil {
		return nil, err
	}
	
	var config struct {
		V    string `json:"v"`
		Ps   string `json:"ps"`
		Add  string `json:"add"`
		Port int    `json:"port"`
		ID   string `json:"id"`
		Net  string `json:"net"`
		Type string `json:"type"`
		Host string `json:"host"`
		Path string `json:"path"`
		TLS  string `json:"tls"`
		Sni  string `json:"sni"`
	}
	
	if err := json.Unmarshal(decoded, &config); err != nil {
		return nil, err
	}
	
	node := &model.Node{
		ID:             utils.GenerateID(),
		SubscriptionID: subscriptionID,
		Name:           config.Ps,
		Type:           model.NodeTypeVMess,
		Server:         config.Add,
		Port:           config.Port,
		UUID:           config.ID,
		Network:        config.Net,
		Security:       config.TLS,
		Host:           config.Host,
		Path:           config.Path,
		SNI:            config.Sni,
		Enabled:        true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	return node, nil
}

// parseVLESS 解析VLESS节点
func (s *subscriptionService) parseVLESS(line string, subscriptionID string) (*model.Node, error) {
	// vless://uuid@server:port?params#name
	u, err := utils.ParseURL(line)
	if err != nil {
		return nil, err
	}
	
	params, err := utils.ParseQueryString(u.RawQuery)
	if err != nil {
		return nil, err
	}
	
	node := &model.Node{
		ID:             utils.GenerateID(),
		SubscriptionID: subscriptionID,
		Name:           u.Fragment,
		Type:           model.NodeTypeVLESS,
		Server:         u.Hostname(),
		Port:           443,
		UUID:           u.User.Username(),
		Network:        params["type"],
		Security:       params["security"],
		Host:           params["host"],
		Path:           params["path"],
		SNI:            params["sni"],
		Flow:           params["flow"],
		Enabled:        true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// 解析端口
	if port := u.Port(); port != "" {
		fmt.Sscanf(port, "%d", &node.Port)
	}
	
	// REALITY配置
	if params["security"] == "reality" {
		node.RealityPublicKey = params["pbk"]
		node.RealityShortID = params["sid"]
		node.RealityFingerprint = params["fp"]
	}
	
	return node, nil
}

// parseTrojan 解析Trojan节点
func (s *subscriptionService) parseTrojan(line string, subscriptionID string) (*model.Node, error) {
	// trojan://password@server:port?params#name
	u, err := utils.ParseURL(line)
	if err != nil {
		return nil, err
	}
	
	params, err := utils.ParseQueryString(u.RawQuery)
	if err != nil {
		return nil, err
	}
	
	node := &model.Node{
		ID:             utils.GenerateID(),
		SubscriptionID: subscriptionID,
		Name:           u.Fragment,
		Type:           model.NodeTypeTrojan,
		Server:         u.Hostname(),
		Port:           443,
		Password:       u.User.Username(),
		Network:        params["type"],
		Security:       "tls",
		Host:           params["host"],
		Path:           params["path"],
		SNI:            params["sni"],
		Enabled:        true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// 解析端口
	if port := u.Port(); port != "" {
		fmt.Sscanf(port, "%d", &node.Port)
	}
	
	return node, nil
}

// parseShadowsocks 解析Shadowsocks节点
func (s *subscriptionService) parseShadowsocks(line string, subscriptionID string) (*model.Node, error) {
	// ss://base64@server:port#name
	// 或 ss://base64#name (base64中包含server:port)
	
	u, err := utils.ParseURL(line)
	if err != nil {
		return nil, err
	}
	
	// 解析用户信息（method:password）
	userInfo := u.User.Username()
	parts := strings.SplitN(userInfo, ":", 2)
	if len(parts) != 2 {
		// 尝试Base64解码
		decoded, err := utils.Base64Decode(userInfo)
		if err != nil {
			return nil, fmt.Errorf("无效的Shadowsocks格式")
		}
		parts = strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("无效的Shadowsocks格式")
		}
	}
	
	node := &model.Node{
		ID:             utils.GenerateID(),
		SubscriptionID: subscriptionID,
		Name:           u.Fragment,
		Type:           model.NodeTypeShadowsocks,
		Server:         u.Hostname(),
		Port:           8388,
		Method:         parts[0],
		Password:       parts[1],
		Enabled:        true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// 解析端口
	if port := u.Port(); port != "" {
		fmt.Sscanf(port, "%d", &node.Port)
	}
	
	return node, nil
}
