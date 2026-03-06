package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"proxy_server/internal/model"
	"proxy_server/pkg/utils"
)

// SubscriptionRepository 订阅仓库接口
type SubscriptionRepository interface {
	Create(subscription *model.Subscription) error
	GetByID(id string) (*model.Subscription, error)
	GetAll() ([]*model.Subscription, error)
	Update(subscription *model.Subscription) error
	Delete(id string) error
}

// subscriptionRepository 订阅仓库实现
type subscriptionRepository struct {
	dataFile string
	mu       sync.RWMutex
}

// NewSubscriptionRepository 创建订阅仓库
func NewSubscriptionRepository(dataDir string) SubscriptionRepository {
	return &subscriptionRepository{
		dataFile: filepath.Join(dataDir, "subscriptions.json"),
	}
}

// loadFromFile 从文件加载数据
func (r *subscriptionRepository) loadFromFile() ([]*model.Subscription, error) {
	data, err := os.ReadFile(r.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*model.Subscription{}, nil
		}
		return nil, err
	}
	
	var subscriptions []*model.Subscription
	if err := json.Unmarshal(data, &subscriptions); err != nil {
		return nil, err
	}
	
	return subscriptions, nil
}

// saveToFile 保存数据到文件
func (r *subscriptionRepository) saveToFile(subscriptions []*model.Subscription) error {
	// 确保目录存在
	dir := filepath.Dir(r.dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(subscriptions, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(r.dataFile, data, 0644)
}

// Create 创建订阅
func (r *subscriptionRepository) Create(subscription *model.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	subscriptions, err := r.loadFromFile()
	if err != nil {
		return err
	}
	
	// 生成ID
	if subscription.ID == "" {
		subscription.ID = utils.GenerateID()
	}
	
	subscriptions = append(subscriptions, subscription)
	
	return r.saveToFile(subscriptions)
}

// GetByID 根据ID获取订阅
func (r *subscriptionRepository) GetByID(id string) (*model.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	subscriptions, err := r.loadFromFile()
	if err != nil {
		return nil, err
	}
	
	for _, sub := range subscriptions {
		if sub.ID == id {
			return sub, nil
		}
	}
	
	return nil, nil
}

// GetAll 获取所有订阅
func (r *subscriptionRepository) GetAll() ([]*model.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return r.loadFromFile()
}

// Update 更新订阅
func (r *subscriptionRepository) Update(subscription *model.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	subscriptions, err := r.loadFromFile()
	if err != nil {
		return err
	}
	
	for i, sub := range subscriptions {
		if sub.ID == subscription.ID {
			subscriptions[i] = subscription
			return r.saveToFile(subscriptions)
		}
	}
	
	return nil
}

// Delete 删除订阅
func (r *subscriptionRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	subscriptions, err := r.loadFromFile()
	if err != nil {
		return err
	}
	
	for i, sub := range subscriptions {
		if sub.ID == id {
			subscriptions = append(subscriptions[:i], subscriptions[i+1:]...)
			return r.saveToFile(subscriptions)
		}
	}
	
	return nil
}
