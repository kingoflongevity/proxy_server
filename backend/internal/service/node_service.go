package service

import (
	"fmt"
	"net"
	"sort"
	"time"

	"proxy_server/internal/model"
	"proxy_server/internal/repository"
	"proxy_server/internal/xray"
	"proxy_server/pkg/errors"
	"proxy_server/pkg/logger"
)

// NodeService 节点服务接口
type NodeService interface {
	GetByID(id string) (*model.Node, error)
	GetList(query *model.NodeListQuery) ([]*model.Node, int64, error)
	Update(id string, req *model.NodeUpdateRequest) (*model.Node, error)
	Test(id string) (int, error)
	TestBatch(ids []string) ([]map[string]interface{}, error)
	TestAll() ([]map[string]interface{}, error)
	Connect(nodeID string) error
	Disconnect() error
	Select(id string) error
	GetCurrentNode() *model.Node
	GetStats(id string) (map[string]interface{}, error)
}

// nodeService 节点服务实现
type nodeService struct {
	nodeRepo       repository.NodeRepository
	systemRepo     repository.SystemRepository
	processManager *xray.ProcessManager
	currentNode    *model.Node
}

// NewNodeService 创建节点服务
func NewNodeService(nodeRepo repository.NodeRepository, systemRepo repository.SystemRepository) NodeService {
	return &nodeService{
		nodeRepo:       nodeRepo,
		systemRepo:     systemRepo,
		processManager: xray.NewProcessManager(""),
	}
}

// GetByID 根据ID获取节点
func (s *nodeService) GetByID(id string) (*model.Node, error) {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, errors.NewError(errors.NodeNotFound, "")
	}

	return node, nil
}

// GetList 获取节点列表（支持分页、筛选、排序）
func (s *nodeService) GetList(query *model.NodeListQuery) ([]*model.Node, int64, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	nodes, err := s.nodeRepo.GetAll()
	if err != nil {
		return nil, 0, err
	}

	var filtered []*model.Node
	for _, node := range nodes {
		if query.SubscriptionID != "" && node.SubscriptionID != query.SubscriptionID {
			continue
		}

		if query.Type != "" && string(node.Type) != query.Type {
			continue
		}

		if query.Enabled != nil && node.Enabled != *query.Enabled {
			continue
		}

		filtered = append(filtered, node)
	}

	if query.SortBy != "" {
		sort.Slice(filtered, func(i, j int) bool {
			switch query.SortBy {
			case "latency":
				if query.SortOrder == "desc" {
					return filtered[i].Latency > filtered[j].Latency
				}
				return filtered[i].Latency < filtered[j].Latency
			case "speed":
				if query.SortOrder == "desc" {
					return filtered[i].Speed > filtered[j].Speed
				}
				return filtered[i].Speed < filtered[j].Speed
			case "score":
				if query.SortOrder == "desc" {
					return filtered[i].Score > filtered[j].Score
				}
				return filtered[i].Score < filtered[j].Score
			default:
				return filtered[i].Name < filtered[j].Name
			}
		})
	}

	total := int64(len(filtered))
	start := (query.Page - 1) * query.PageSize
	end := start + query.PageSize

	if start >= len(filtered) {
		return []*model.Node{}, total, nil
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], total, nil
}

// Update 更新节点
func (s *nodeService) Update(id string, req *model.NodeUpdateRequest) (*model.Node, error) {
	node, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		node.Name = req.Name
	}
	node.Enabled = req.Enabled
	node.UpdatedAt = time.Now()

	if err := s.nodeRepo.Update(node); err != nil {
		logger.Error("更新节点失败: %v", err)
		return nil, errors.NewError(errors.DataSaveError, err.Error())
	}

	logger.Info("更新节点成功: %s", node.Name)
	return node, nil
}

// Test 测试节点延迟
func (s *nodeService) Test(id string) (int, error) {
	node, err := s.GetByID(id)
	if err != nil {
		return 0, err
	}

	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", node.Server, node.Port), 5*time.Second)
	if err != nil {
		logger.Warn("节点测试失败: %s - %v", node.Name, err)
		node.Latency = 0
		node.LastTest = time.Now()
		s.nodeRepo.Update(node)
		return 0, errors.NewError(errors.NodeTestFailed, err.Error())
	}
	defer conn.Close()

	latency := int(time.Since(start).Milliseconds())

	node.Latency = latency
	node.LastTest = time.Now()
	node.Score = s.calculateScore(node)
	s.nodeRepo.Update(node)

	logger.Info("节点测试成功: %s - %dms", node.Name, latency)
	return latency, nil
}

// calculateScore 计算节点评分
func (s *nodeService) calculateScore(node *model.Node) int {
	score := 100
	if node.Latency > 0 {
		if node.Latency < 100 {
			score = 100
		} else if node.Latency < 200 {
			score = 90
		} else if node.Latency < 300 {
			score = 80
		} else if node.Latency < 500 {
			score = 70
		} else {
			score = 60
		}
	}
	return score
}

// Connect 连接到指定节点
// 使用Xray进程管理器启动代理
func (s *nodeService) Connect(nodeID string) error {
	node, err := s.GetByID(nodeID)
	if err != nil {
		return err
	}

	if s.processManager.IsRunning() {
		currentNode := s.processManager.GetCurrentNode()
		if currentNode != nil && currentNode.ID == nodeID {
			return errors.NewError(errors.NodeAlreadyConnected, "")
		}
		s.processManager.Stop()
	}

	if err := s.processManager.Start(node, 10808); err != nil {
		logger.Error("启动Xray进程失败: %v", err)
		return errors.NewError(errors.NodeConnectFailed, err.Error())
	}

	s.currentNode = node
	node.Connected = true
	s.nodeRepo.Update(node)

	status, _ := s.systemRepo.GetStatus()
	status.Connected = true
	status.CurrentNode = node
	s.systemRepo.SaveStatus(status)

	logger.Info("连接节点成功: %s (SOCKS5: 0.0.0.0:10808, HTTP: 0.0.0.0:10809)", node.Name)
	return nil
}

// Disconnect 断开连接
func (s *nodeService) Disconnect() error {
	if !s.processManager.IsRunning() {
		return nil
	}

	if err := s.processManager.Stop(); err != nil {
		logger.Error("停止Xray进程失败: %v", err)
		return err
	}

	if s.currentNode != nil {
		s.currentNode.Connected = false
		s.nodeRepo.Update(s.currentNode)
	}

	status, _ := s.systemRepo.GetStatus()
	status.Connected = false
	status.CurrentNode = nil
	s.systemRepo.SaveStatus(status)

	logger.Info("断开节点连接")
	s.currentNode = nil

	return nil
}

// GetCurrentNode 获取当前连接的节点
func (s *nodeService) GetCurrentNode() *model.Node {
	if s.processManager.IsRunning() {
		return s.processManager.GetCurrentNode()
	}
	return nil
}

// Select 选择节点并启动代理连接
func (s *nodeService) Select(id string) error {
	// 先断开当前连接
	if s.processManager.IsRunning() {
		s.processManager.Stop()
		if s.currentNode != nil {
			s.currentNode.Connected = false
			s.nodeRepo.Update(s.currentNode)
		}
	}

	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 启动代理连接
	if err := s.processManager.Start(node, 10808); err != nil {
		logger.Error("启动Xray进程失败: %v", err)
		return errors.NewError(errors.NodeConnectFailed, err.Error())
	}

	s.currentNode = node
	node.Connected = true
	s.nodeRepo.Update(node)

	status, _ := s.systemRepo.GetStatus()
	status.Connected = true
	status.CurrentNode = node
	s.systemRepo.SaveStatus(status)

	logger.Info("选择并连接节点成功: %s (SOCKS5: 0.0.0.0:10808, HTTP: 0.0.0.0:10809)", node.Name)
	return nil
}

// GetStats 获取节点统计信息
func (s *nodeService) GetStats(id string) (map[string]interface{}, error) {
	// 返回模拟的统计信息
	return map[string]interface{}{
		"uploadSpeed":     0,
		"downloadSpeed":   0,
		"uploadTotal":     0,
		"downloadTotal":   0,
		"connectionCount": 0,
	}, nil
}

// TestBatch 批量测试节点
func (s *nodeService) TestBatch(ids []string) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0, len(ids))
	for _, id := range ids {
		latency, err := s.Test(id)
		status := "available"
		if err != nil || latency == 0 {
			status = "unavailable"
		}
		results = append(results, map[string]interface{}{
			"nodeId":   id,
			"latency":  latency,
			"status":   status,
			"testTime": time.Now().Format(time.RFC3339),
			"error":    err,
		})
	}
	return results, nil
}

// TestAll 测试所有节点
func (s *nodeService) TestAll() ([]map[string]interface{}, error) {
	nodes, err := s.nodeRepo.GetAll()
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(nodes))
	for i, node := range nodes {
		ids[i] = node.ID
	}

	return s.TestBatch(ids)
}
