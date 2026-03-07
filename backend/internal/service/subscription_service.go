package service

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"proxy_server/internal/model"
	"proxy_server/internal/repository"
	"proxy_server/internal/xray"
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
	RefreshAndGetNodes(id string) ([]*model.Node, error)
}

// subscriptionService 订阅服务实现
type subscriptionService struct {
	subRepo  repository.SubscriptionRepository
	nodeRepo repository.NodeRepository
	parser   *xray.SubscriptionParser
}

// NewSubscriptionService 创建订阅服务
func NewSubscriptionService(subRepo repository.SubscriptionRepository, nodeRepo repository.NodeRepository) SubscriptionService {
	return &subscriptionService{
		subRepo:  subRepo,
		nodeRepo: nodeRepo,
		parser:   xray.NewSubscriptionParser(),
	}
}

// Create 创建订阅
func (s *subscriptionService) Create(req *model.SubscriptionCreateRequest) (*model.Subscription, error) {
	if req.UpdateInterval == 0 {
		req.UpdateInterval = 24
	}

	subscriptionType := req.Type
	if subscriptionType == "" {
		subscriptionType = "mixed"
	}

	parseFormat := req.ParseFormat
	if parseFormat == "" {
		parseFormat = model.ParseFormatAuto
	}

	subscription := &model.Subscription{
		ID:             utils.GenerateID(),
		Name:           req.Name,
		URL:            req.URL,
		Type:           subscriptionType,
		ParseFormat:    parseFormat,
		Status:         "active",
		AutoUpdate:     req.AutoUpdate,
		UpdateInterval: req.UpdateInterval,
		NodeCount:      0,
		LastUpdate:     time.Time{},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.subRepo.Create(subscription); err != nil {
		logger.Error("创建订阅失败: %v", err)
		return nil, errors.NewError(errors.DataSaveError, err.Error())
	}

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

	if req.Name != "" {
		subscription.Name = req.Name
	}
	if req.URL != "" {
		subscription.URL = req.URL
	}
	if req.Type != "" {
		subscription.Type = req.Type
	}
	if req.ParseFormat != "" {
		subscription.ParseFormat = req.ParseFormat
	}
	subscription.AutoUpdate = req.AutoUpdate
	if req.UpdateInterval > 0 {
		subscription.UpdateInterval = req.UpdateInterval
	}
	subscription.UpdatedAt = time.Now()

	if err := s.subRepo.Update(subscription); err != nil {
		logger.Error("更新订阅失败: %v", err)
		return nil, errors.NewError(errors.DataSaveError, err.Error())
	}

	logger.Info("更新订阅成功: %s", subscription.Name)
	return subscription, nil
}

// Delete 删除订阅
func (s *subscriptionService) Delete(id string) error {
	_, err := s.GetByID(id)
	if err != nil {
		return err
	}

	if err := s.nodeRepo.DeleteBySubscriptionID(id); err != nil {
		logger.Error("删除订阅节点失败: %v", err)
		return errors.NewError(errors.DataSaveError, err.Error())
	}

	if err := s.subRepo.Delete(id); err != nil {
		logger.Error("删除订阅失败: %v", err)
		return errors.NewError(errors.DataSaveError, err.Error())
	}

	logger.Info("删除订阅成功: %s", id)
	return nil
}

// Refresh 刷新订阅节点
// 使用Xray订阅解析器解析节点，支持v26.2.6新特性
func (s *subscriptionService) Refresh(id string) error {
	_, err := s.RefreshAndGetNodes(id)
	return err
}

// RefreshAndGetNodes 刷新订阅并返回节点列表
func (s *subscriptionService) RefreshAndGetNodes(id string) ([]*model.Node, error) {
	subscription, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	content, err := s.fetchSubscription(subscription.URL)
	if err != nil {
		logger.Error("获取订阅内容失败: %v", err)
		subscription.Status = "error"
		s.subRepo.Update(subscription)
		return nil, errors.NewError(errors.SubscriptionFetchFailed, err.Error())
	}

	nodes, err := s.parser.ParseWithFormat(content, subscription.ParseFormat)
	if err != nil {
		logger.Error("解析订阅节点失败: %v", err)
		subscription.Status = "error"
		s.subRepo.Update(subscription)
		return nil, errors.NewError(errors.SubscriptionParseFailed, err.Error())
	}

	if err := s.nodeRepo.DeleteBySubscriptionID(id); err != nil {
		logger.Error("删除旧节点失败: %v", err)
		return nil, errors.NewError(errors.DataSaveError, err.Error())
	}

	for _, node := range nodes {
		node.SubscriptionID = id
		if err := s.nodeRepo.Create(node); err != nil {
			logger.Warn("保存节点失败: %s - %v", node.Name, err)
		}
	}

	subscription.NodeCount = len(nodes)
	subscription.LastUpdate = time.Now()
	subscription.UpdatedAt = time.Now()
	subscription.Status = "active"
	if err := s.subRepo.Update(subscription); err != nil {
		logger.Error("更新订阅信息失败: %v", err)
	}

	logger.Info("刷新订阅成功: %s, 节点数: %d", subscription.Name, len(nodes))
	return nodes, nil
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

	decoded, err := utils.Base64Decode(string(body))
	if err == nil {
		return string(decoded), nil
	}

	return string(body), nil
}
