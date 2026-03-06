package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"proxy_server/internal/model"
	"proxy_server/pkg/utils"
)

// NodeRepository 节点仓库接口
type NodeRepository interface {
	Create(node *model.Node) error
	GetByID(id string) (*model.Node, error)
	GetAll() ([]*model.Node, error)
	GetBySubscriptionID(subscriptionID string) ([]*model.Node, error)
	Update(node *model.Node) error
	Delete(id string) error
	DeleteBySubscriptionID(subscriptionID string) error
}

// nodeRepository 节点仓库实现
type nodeRepository struct {
	dataFile string
	mu       sync.RWMutex
}

// NewNodeRepository 创建节点仓库
func NewNodeRepository(dataDir string) NodeRepository {
	return &nodeRepository{
		dataFile: filepath.Join(dataDir, "nodes.json"),
	}
}

// loadFromFile 从文件加载数据
func (r *nodeRepository) loadFromFile() ([]*model.Node, error) {
	data, err := os.ReadFile(r.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*model.Node{}, nil
		}
		return nil, err
	}
	
	var nodes []*model.Node
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, err
	}
	
	return nodes, nil
}

// saveToFile 保存数据到文件
func (r *nodeRepository) saveToFile(nodes []*model.Node) error {
	// 确保目录存在
	dir := filepath.Dir(r.dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(r.dataFile, data, 0644)
}

// Create 创建节点
func (r *nodeRepository) Create(node *model.Node) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	nodes, err := r.loadFromFile()
	if err != nil {
		return err
	}
	
	// 生成ID
	if node.ID == "" {
		node.ID = utils.GenerateID()
	}
	
	nodes = append(nodes, node)
	
	return r.saveToFile(nodes)
}

// GetByID 根据ID获取节点
func (r *nodeRepository) GetByID(id string) (*model.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	nodes, err := r.loadFromFile()
	if err != nil {
		return nil, err
	}
	
	for _, node := range nodes {
		if node.ID == id {
			return node, nil
		}
	}
	
	return nil, nil
}

// GetAll 获取所有节点
func (r *nodeRepository) GetAll() ([]*model.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return r.loadFromFile()
}

// GetBySubscriptionID 根据订阅ID获取节点
func (r *nodeRepository) GetBySubscriptionID(subscriptionID string) ([]*model.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	nodes, err := r.loadFromFile()
	if err != nil {
		return nil, err
	}
	
	var result []*model.Node
	for _, node := range nodes {
		if node.SubscriptionID == subscriptionID {
			result = append(result, node)
		}
	}
	
	return result, nil
}

// Update 更新节点
func (r *nodeRepository) Update(node *model.Node) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	nodes, err := r.loadFromFile()
	if err != nil {
		return err
	}
	
	for i, n := range nodes {
		if n.ID == node.ID {
			nodes[i] = node
			return r.saveToFile(nodes)
		}
	}
	
	return nil
}

// Delete 删除节点
func (r *nodeRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	nodes, err := r.loadFromFile()
	if err != nil {
		return err
	}
	
	for i, node := range nodes {
		if node.ID == id {
			nodes = append(nodes[:i], nodes[i+1:]...)
			return r.saveToFile(nodes)
		}
	}
	
	return nil
}

// DeleteBySubscriptionID 删除指定订阅的所有节点
func (r *nodeRepository) DeleteBySubscriptionID(subscriptionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	nodes, err := r.loadFromFile()
	if err != nil {
		return err
	}
	
	var result []*model.Node
	for _, node := range nodes {
		if node.SubscriptionID != subscriptionID {
			result = append(result, node)
		}
	}
	
	return r.saveToFile(result)
}
