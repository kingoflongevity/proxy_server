package service

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"proxy_server/internal/model"
	"proxy_server/internal/repository"
	"proxy_server/pkg/logger"
)

// ScalerService 自动伸缩服务接口
type ScalerService interface {
	// 伸缩策略管理
	UpdateScalePolicy(req *model.ScalePolicyUpdateRequest) error
	GetScalePolicy(groupID string) (*model.ServerGroup, error)
	
	// 手动伸缩
	ScaleUp(groupID string, count int) error
	ScaleDown(groupID string, count int) error
	
	// 自动伸缩
	StartAutoScale(ctx context.Context) error
	StopAutoScale()
	
	// 伸缩事件
	GetScaleEvents(groupID string, limit int) ([]*model.ScaleEvent, error)
	
	// 监控数据
	GetGroupMetrics(groupID string) (*GroupMetrics, error)
}

// GroupMetrics 分组监控指标
type GroupMetrics struct {
	GroupID          string    `json:"group_id"`
	ServerCount      int       `json:"server_count"`
	ActiveCount      int       `json:"active_count"`
	AvgCPUUsage      float64   `json:"avg_cpu_usage"`
	AvgMemoryUsage   float64   `json:"avg_memory_usage"`
	TotalConnections int       `json:"total_connections"`
	TotalBandwidthUp   int64   `json:"total_bandwidth_up"`
	TotalBandwidthDown int64   `json:"total_bandwidth_down"`
	Timestamp        time.Time `json:"timestamp"`
}

// scalerService 自动伸缩服务实现
type scalerService struct {
	groupRepo    repository.ClusterServerRepository
	serverRepo   repository.ClusterServerRepository
	eventRepo    repository.ScaleEventRepository
	deployer     DeployerService
	scanner      ScannerService
	
	mu           sync.RWMutex
	running      bool
	stopChan     chan struct{}
	checkInterval time.Duration
}

// NewScalerService 创建自动伸缩服务
func NewScalerService(
	groupRepo repository.ClusterServerRepository,
	serverRepo repository.ClusterServerRepository,
	eventRepo repository.ScaleEventRepository,
	deployer DeployerService,
	scanner ScannerService,
) ScalerService {
	return &scalerService{
		groupRepo:    groupRepo,
		serverRepo:   serverRepo,
		eventRepo:    eventRepo,
		deployer:     deployer,
		scanner:      scanner,
		checkInterval: 30 * time.Second,
		stopChan:     make(chan struct{}),
	}
}

// UpdateScalePolicy 更新伸缩策略
func (s *scalerService) UpdateScalePolicy(req *model.ScalePolicyUpdateRequest) error {
	group, err := s.groupRepo.GetGroupByID(req.GroupID)
	if err != nil {
		return err
	}

	// 更新策略
	group.AutoScale = req.AutoScale
	group.MinServers = req.MinServers
	group.MaxServers = req.MaxServers
	group.ScalePolicy = req.ScalePolicy
	group.ScaleThreshold = req.ScaleThreshold
	group.UpdatedAt = time.Now()

	return s.groupRepo.UpdateGroup(group)
}

// GetScalePolicy 获取伸缩策略
func (s *scalerService) GetScalePolicy(groupID string) (*model.ServerGroup, error) {
	return s.groupRepo.GetGroupByID(groupID)
}

// ScaleUp 扩容
// 增加指定数量的服务器到分组
func (s *scalerService) ScaleUp(groupID string, count int) error {
	group, err := s.groupRepo.GetGroupByID(groupID)
	if err != nil {
		return err
	}

	// 获取分组内当前服务器
	servers, err := s.serverRepo.GetByGroupID(groupID)
	if err != nil {
		return err
	}

	currentCount := len(servers)
	targetCount := currentCount + count

	// 检查是否超过最大限制
	if targetCount > group.MaxServers {
		return fmt.Errorf("超过最大服务器数量限制 (%d)", group.MaxServers)
	}

	// 创建伸缩事件
	event := &model.ScaleEvent{
		ID:          generateID(),
		GroupID:     groupID,
		Type:        "scale_up",
		Reason:      fmt.Sprintf("手动扩容: %d -> %d", currentCount, targetCount),
		TargetCount: targetCount,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	s.eventRepo.CreateEvent(event)

	// 获取可用服务器（已部署但未分配分组）
	availableServers, err := s.getAvailableServers(count)
	if err != nil {
		event.Status = "failed"
		s.eventRepo.UpdateEvent(event)
		return err
	}

	// 分配服务器到分组
	for _, server := range availableServers {
		server.GroupID = groupID
		s.serverRepo.Update(server)
		
		// 启动代理服务
		s.deployer.StartProxy(server.ID)
	}

	// 更新事件状态
	event.Status = "completed"
	completedAt := time.Now()
	event.CompletedAt = &completedAt
	s.eventRepo.UpdateEvent(event)

	logger.Info("扩容完成: 分组 %s, %d -> %d 台服务器", groupID, currentCount, targetCount)
	return nil
}

// ScaleDown 缩容
// 从分组移除指定数量的服务器
func (s *scalerService) ScaleDown(groupID string, count int) error {
	group, err := s.groupRepo.GetGroupByID(groupID)
	if err != nil {
		return err
	}

	// 获取分组内当前服务器
	servers, err := s.serverRepo.GetByGroupID(groupID)
	if err != nil {
		return err
	}

	currentCount := len(servers)
	targetCount := currentCount - count

	// 检查是否低于最小限制
	if targetCount < group.MinServers {
		return fmt.Errorf("低于最小服务器数量限制 (%d)", group.MinServers)
	}

	// 创建伸缩事件
	event := &model.ScaleEvent{
		ID:          generateID(),
		GroupID:     groupID,
		Type:        "scale_down",
		Reason:      fmt.Sprintf("手动缩容: %d -> %d", currentCount, targetCount),
		TargetCount: targetCount,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	s.eventRepo.CreateEvent(event)

	// 选择要移除的服务器（优先移除负载最低的）
	serversToRemove := s.selectServersToRemove(servers, count)

	for _, server := range serversToRemove {
		// 停止代理服务
		s.deployer.StopProxy(server.ID)
		
		// 移出分组
		server.GroupID = ""
		s.serverRepo.Update(server)
	}

	// 更新事件状态
	event.Status = "completed"
	completedAt := time.Now()
	event.CompletedAt = &completedAt
	s.eventRepo.UpdateEvent(event)

	logger.Info("缩容完成: 分组 %s, %d -> %d 台服务器", groupID, currentCount, targetCount)
	return nil
}

// StartAutoScale 启动自动伸缩
func (s *scalerService) StartAutoScale(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("自动伸缩已在运行")
	}

	s.running = true
	s.stopChan = make(chan struct{})

	go s.autoScaleLoop(ctx)

	logger.Info("自动伸缩服务已启动")
	return nil
}

// StopAutoScale 停止自动伸缩
func (s *scalerService) StopAutoScale() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	close(s.stopChan)
	s.running = false

	logger.Info("自动伸缩服务已停止")
}

// autoScaleLoop 自动伸缩循环
func (s *scalerService) autoScaleLoop(ctx context.Context) {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAndScale()
		}
	}
}

// checkAndScale 检查并执行伸缩
func (s *scalerService) checkAndScale() {
	// 获取所有启用自动伸缩的分组
	groups, err := s.groupRepo.GetAutoScaleGroups()
	if err != nil {
		logger.Warn("获取自动伸缩分组失败: %v", err)
		return
	}

	for _, group := range groups {
		metrics, err := s.GetGroupMetrics(group.ID)
		if err != nil {
			logger.Warn("获取分组指标失败: %s - %v", group.ID, err)
			continue
		}

		// 判断是否需要伸缩
		shouldScale, direction := s.evaluateScaleNeed(group, metrics)
		if !shouldScale {
			continue
		}

		// 执行伸缩
		if direction == "up" {
			err = s.ScaleUp(group.ID, 1)
		} else {
			err = s.ScaleDown(group.ID, 1)
		}

		if err != nil {
			logger.Warn("自动伸缩失败: %s - %v", group.ID, err)
		}
	}
}

// evaluateScaleNeed 评估是否需要伸缩
func (s *scalerService) evaluateScaleNeed(group *model.ServerGroup, metrics *GroupMetrics) (bool, string) {
	var currentValue float64

	// 根据策略选择指标
	switch group.ScalePolicy {
	case "cpu":
		currentValue = metrics.AvgCPUUsage
	case "memory":
		currentValue = metrics.AvgMemoryUsage
	case "connections":
		// 连接数需要归一化
		if metrics.ServerCount > 0 {
			currentValue = float64(metrics.TotalConnections) / float64(metrics.ServerCount)
		}
	default:
		currentValue = metrics.AvgCPUUsage
	}

	// 判断是否需要扩容
	if currentValue > group.ScaleThreshold && metrics.ServerCount < group.MaxServers {
		return true, "up"
	}

	// 判断是否需要缩容（低于阈值的一半）
	if currentValue < group.ScaleThreshold/2 && metrics.ServerCount > group.MinServers {
		return true, "down"
	}

	return false, ""
}

// GetGroupMetrics 获取分组监控指标
func (s *scalerService) GetGroupMetrics(groupID string) (*GroupMetrics, error) {
	servers, err := s.serverRepo.GetByGroupID(groupID)
	if err != nil {
		return nil, err
	}

	metrics := &GroupMetrics{
		GroupID:    groupID,
		Timestamp:  time.Now(),
	}

	if len(servers) == 0 {
		return metrics, nil
	}

	var totalCPU, totalMemory float64
	var activeCount int

	for _, server := range servers {
		if server.Status == model.ServerStatusOnline {
			activeCount++
		}

		totalCPU += server.CPUUsage
		totalMemory += server.MemoryUsage
		metrics.TotalConnections += server.Connections
		metrics.TotalBandwidthUp += int64(server.BandwidthUp)
		metrics.TotalBandwidthDown += int64(server.BandwidthDown)
	}

	metrics.ServerCount = len(servers)
	metrics.ActiveCount = activeCount
	metrics.AvgCPUUsage = totalCPU / float64(len(servers))
	metrics.AvgMemoryUsage = totalMemory / float64(len(servers))

	// 四舍五入
	metrics.AvgCPUUsage = math.Round(metrics.AvgCPUUsage*100) / 100
	metrics.AvgMemoryUsage = math.Round(metrics.AvgMemoryUsage*100) / 100

	return metrics, nil
}

// GetScaleEvents 获取伸缩事件
func (s *scalerService) GetScaleEvents(groupID string, limit int) ([]*model.ScaleEvent, error) {
	return s.eventRepo.GetEventsByGroupID(groupID, limit)
}

// ====== 辅助方法 ======

// getAvailableServers 获取可用服务器
func (s *scalerService) getAvailableServers(count int) ([]*model.ClusterServer, error) {
	// 获取未分配分组的服务器
	allServers, err := s.serverRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var available []*model.ClusterServer
	for _, server := range allServers {
		if server.GroupID == "" && server.Status == model.ServerStatusOnline {
			available = append(available, server)
			if len(available) >= count {
				break
			}
		}
	}

	if len(available) < count {
		return nil, fmt.Errorf("可用服务器不足，需要 %d 台，可用 %d 台", count, len(available))
	}

	return available, nil
}

// selectServersToRemove 选择要移除的服务器
// 优先移除负载最低的服务器
func (s *scalerService) selectServersToRemove(servers []*model.ClusterServer, count int) []*model.ClusterServer {
	if len(servers) <= count {
		return servers
	}

	// 按负载排序（CPU使用率 + 内存使用率）
	type serverLoad struct {
		server *model.ClusterServer
		load   float64
	}

	var loads []serverLoad
	for _, s := range servers {
		load := s.CPUUsage + s.MemoryUsage
		loads = append(loads, serverLoad{server: s, load: load})
	}

	// 排序（负载低的排前面）
	for i := 0; i < len(loads); i++ {
		for j := i + 1; j < len(loads); j++ {
			if loads[j].load < loads[i].load {
				loads[i], loads[j] = loads[j], loads[i]
			}
		}
	}

	// 选择前count个
	result := make([]*model.ClusterServer, count)
	for i := 0; i < count; i++ {
		result[i] = loads[i].server
	}

	return result
}

// generateID 生成ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
