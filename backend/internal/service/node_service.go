package service

import (
	"fmt"
	"net"
	"sort"
	"time"

	"proxy_server/internal/model"
	"proxy_server/internal/repository"
	"proxy_server/pkg/errors"
	"proxy_server/pkg/logger"
)

// NodeService 节点服务接口
type NodeService interface {
	GetByID(id string) (*model.Node, error)
	GetList(query *model.NodeListQuery) ([]*model.Node, int64, error)
	Update(id string, req *model.NodeUpdateRequest) (*model.Node, error)
	Test(id string) (int, error)
	Connect(nodeID string) error
	Disconnect() error
	GetCurrentNode() *model.Node
}

// nodeService 节点服务实现
type nodeService struct {
	nodeRepo   repository.NodeRepository
	systemRepo repository.SystemRepository
	currentNode *model.Node
}

// NewNodeService 创建节点服务
func NewNodeService(nodeRepo repository.NodeRepository, systemRepo repository.SystemRepository) NodeService {
	return &nodeService{
		nodeRepo:   nodeRepo,
		systemRepo: systemRepo,
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
	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	
	// 获取所有节点
	nodes, err := s.nodeRepo.GetAll()
	if err != nil {
		return nil, 0, err
	}
	
	// 筛选
	var filtered []*model.Node
	for _, node := range nodes {
		// 按订阅ID筛选
		if query.SubscriptionID != "" && node.SubscriptionID != query.SubscriptionID {
			continue
		}
		
		// 按类型筛选
		if query.Type != "" && string(node.Type) != query.Type {
			continue
		}
		
		// 按启用状态筛选
		if query.Enabled != nil && node.Enabled != *query.Enabled {
			continue
		}
		
		filtered = append(filtered, node)
	}
	
	// 排序
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
	
	// 分页
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
	
	// 更新字段
	if req.Name != "" {
		node.Name = req.Name
	}
	node.Enabled = req.Enabled
	node.UpdatedAt = time.Now()
	
	// 保存
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
	
	// 测试TCP连接延迟
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
	
	// 更新节点延迟
	node.Latency = latency
	node.LastTest = time.Now()
	node.Score = s.calculateScore(node)
	s.nodeRepo.Update(node)
	
	logger.Info("节点测试成功: %s - %dms", node.Name, latency)
	return latency, nil
}

// calculateScore 计算节点评分
func (s *nodeService) calculateScore(node *model.Node) int {
	// 简单的评分算法：延迟越低分数越高
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
func (s *nodeService) Connect(nodeID string) error {
	node, err := s.GetByID(nodeID)
	if err != nil {
		return err
	}
	
	// 检查是否已连接
	if s.currentNode != nil && s.currentNode.ID == nodeID {
		return errors.NewError(errors.NodeAlreadyConnected, "")
	}
	
	// TODO: 实际的代理连接逻辑
	// 这里需要集成Xray-core或其他代理核心
	
	// 更新连接状态
	s.currentNode = node
	node.Connected = true
	s.nodeRepo.Update(node)
	
	// 更新系统状态
	status, _ := s.systemRepo.GetStatus()
	status.Connected = true
	status.CurrentNode = node
	s.systemRepo.SaveStatus(status)
	
	logger.Info("连接节点成功: %s", node.Name)
	return nil
}

// Disconnect 断开连接
func (s *nodeService) Disconnect() error {
	if s.currentNode == nil {
		return nil
	}
	
	// TODO: 实际的代理断开逻辑
	
	// 更新节点状态
	s.currentNode.Connected = false
	s.nodeRepo.Update(s.currentNode)
	
	// 更新系统状态
	status, _ := s.systemRepo.GetStatus()
	status.Connected = false
	status.CurrentNode = nil
	s.systemRepo.SaveStatus(status)
	
	logger.Info("断开节点连接: %s", s.currentNode.Name)
	s.currentNode = nil
	
	return nil
}

// GetCurrentNode 获取当前连接的节点
func (s *nodeService) GetCurrentNode() *model.Node {
	return s.currentNode
}
