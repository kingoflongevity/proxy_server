package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"proxy_server/internal/model"
	"proxy_server/pkg/utils"
)

// RuleRepository 规则仓库接口
type RuleRepository interface {
	Create(rule *model.Rule) error
	GetByID(id string) (*model.Rule, error)
	GetAll() ([]*model.Rule, error)
	Update(rule *model.Rule) error
	Delete(id string) error
}

// ruleRepository 规则仓库实现
type ruleRepository struct {
	dataFile string
	mu       sync.RWMutex
}

// NewRuleRepository 创建规则仓库
func NewRuleRepository(dataDir string) RuleRepository {
	return &ruleRepository{
		dataFile: filepath.Join(dataDir, "rules.json"),
	}
}

// loadFromFile 从文件加载数据
func (r *ruleRepository) loadFromFile() ([]*model.Rule, error) {
	data, err := os.ReadFile(r.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*model.Rule{}, nil
		}
		return nil, err
	}
	
	var rules []*model.Rule
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, err
	}
	
	return rules, nil
}

// saveToFile 保存数据到文件
func (r *ruleRepository) saveToFile(rules []*model.Rule) error {
	// 确保目录存在
	dir := filepath.Dir(r.dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(r.dataFile, data, 0644)
}

// Create 创建规则
func (r *ruleRepository) Create(rule *model.Rule) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	rules, err := r.loadFromFile()
	if err != nil {
		return err
	}
	
	// 生成ID
	if rule.ID == "" {
		rule.ID = utils.GenerateID()
	}
	
	rules = append(rules, rule)
	
	return r.saveToFile(rules)
}

// GetByID 根据ID获取规则
func (r *ruleRepository) GetByID(id string) (*model.Rule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	rules, err := r.loadFromFile()
	if err != nil {
		return nil, err
	}
	
	for _, rule := range rules {
		if rule.ID == id {
			return rule, nil
		}
	}
	
	return nil, nil
}

// GetAll 获取所有规则
func (r *ruleRepository) GetAll() ([]*model.Rule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return r.loadFromFile()
}

// Update 更新规则
func (r *ruleRepository) Update(rule *model.Rule) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	rules, err := r.loadFromFile()
	if err != nil {
		return err
	}
	
	for i, ru := range rules {
		if ru.ID == rule.ID {
			rules[i] = rule
			return r.saveToFile(rules)
		}
	}
	
	return nil
}

// Delete 删除规则
func (r *ruleRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	rules, err := r.loadFromFile()
	if err != nil {
		return err
	}
	
	for i, rule := range rules {
		if rule.ID == id {
			rules = append(rules[:i], rules[i+1:]...)
			return r.saveToFile(rules)
		}
	}
	
	return nil
}
