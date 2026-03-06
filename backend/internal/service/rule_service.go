package service

import (
	"time"

	"proxy_server/internal/model"
	"proxy_server/internal/repository"
	"proxy_server/pkg/errors"
	"proxy_server/pkg/logger"
	"proxy_server/pkg/utils"
)

// RuleService 规则服务接口
type RuleService interface {
	Create(req *model.RuleCreateRequest) (*model.Rule, error)
	GetByID(id string) (*model.Rule, error)
	GetAll() ([]*model.Rule, error)
	Update(id string, req *model.RuleUpdateRequest) (*model.Rule, error)
	Delete(id string) error
	UpdatePriority(ruleIDs []string) error
	GetRuleSets() ([]map[string]interface{}, error)
	UpdateRuleSet(id string) error
}

// ruleService 规则服务实现
type ruleService struct {
	ruleRepo repository.RuleRepository
}

// NewRuleService 创建规则服务
func NewRuleService(ruleRepo repository.RuleRepository) RuleService {
	return &ruleService{
		ruleRepo: ruleRepo,
	}
}

// Create 创建规则
func (s *ruleService) Create(req *model.RuleCreateRequest) (*model.Rule, error) {
	// 验证规则类型
	if !isValidRuleType(req.Type) {
		return nil, errors.NewError(errors.RuleInvalid, "无效的规则类型")
	}
	
	// 验证规则策略
	if !isValidRulePolicy(req.Policy) {
		return nil, errors.NewError(errors.RuleInvalid, "无效的规则策略")
	}
	
	rule := &model.Rule{
		ID:          utils.GenerateID(),
		Type:        req.Type,
		Value:       req.Value,
		Policy:      req.Policy,
		Description: req.Description,
		Enabled:     req.Enabled,
		Priority:    req.Priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// 保存规则
	if err := s.ruleRepo.Create(rule); err != nil {
		logger.Error("创建规则失败: %v", err)
		return nil, errors.NewError(errors.DataSaveError, err.Error())
	}
	
	logger.Info("创建规则成功: %s - %s", rule.Type, rule.Value)
	return rule, nil
}

// GetByID 根据ID获取规则
func (s *ruleService) GetByID(id string) (*model.Rule, error) {
	rule, err := s.ruleRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	
	if rule == nil {
		return nil, errors.NewError(errors.RuleNotFound, "")
	}
	
	return rule, nil
}

// GetAll 获取所有规则
func (s *ruleService) GetAll() ([]*model.Rule, error) {
	return s.ruleRepo.GetAll()
}

// Update 更新规则
func (s *ruleService) Update(id string, req *model.RuleUpdateRequest) (*model.Rule, error) {
	rule, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	
	// 更新字段
	if req.Type != "" {
		if !isValidRuleType(req.Type) {
			return nil, errors.NewError(errors.RuleInvalid, "无效的规则类型")
		}
		rule.Type = req.Type
	}
	if req.Value != "" {
		rule.Value = req.Value
	}
	if req.Policy != "" {
		if !isValidRulePolicy(req.Policy) {
			return nil, errors.NewError(errors.RuleInvalid, "无效的规则策略")
		}
		rule.Policy = req.Policy
	}
	if req.Description != "" {
		rule.Description = req.Description
	}
	rule.Enabled = req.Enabled
	rule.Priority = req.Priority
	rule.UpdatedAt = time.Now()
	
	// 保存
	if err := s.ruleRepo.Update(rule); err != nil {
		logger.Error("更新规则失败: %v", err)
		return nil, errors.NewError(errors.DataSaveError, err.Error())
	}
	
	logger.Info("更新规则成功: %s - %s", rule.Type, rule.Value)
	return rule, nil
}

// Delete 删除规则
func (s *ruleService) Delete(id string) error {
	// 检查规则是否存在
	_, err := s.GetByID(id)
	if err != nil {
		return err
	}
	
	// 删除规则
	if err := s.ruleRepo.Delete(id); err != nil {
		logger.Error("删除规则失败: %v", err)
		return errors.NewError(errors.DataSaveError, err.Error())
	}
	
	logger.Info("删除规则成功: %s", id)
	return nil
}

// UpdatePriority 更新规则优先级
func (s *ruleService) UpdatePriority(ruleIDs []string) error {
	// 实现优先级更新逻辑
	logger.Info("更新规则优先级: %v", ruleIDs)
	return nil
}

// GetRuleSets 获取规则集列表
func (s *ruleService) GetRuleSets() ([]map[string]interface{}, error) {
	// 返回模拟的规则集列表
	return []map[string]interface{}{
		{
			"id":          "1",
			"name":        "国内规则",
			"type":        "china",
			"ruleCount":   100,
			"lastUpdate":  "2024-01-01",
		},
		{
			"id":          "2",
			"name":        "国外规则",
			"type":        "foreign",
			"ruleCount":   200,
			"lastUpdate":  "2024-01-01",
		},
	}, nil
}

// UpdateRuleSet 更新规则集
func (s *ruleService) UpdateRuleSet(id string) error {
	logger.Info("更新规则集: %s", id)
	return nil
}

// isValidRuleType 验证规则类型
func isValidRuleType(ruleType model.RuleType) bool {
	switch ruleType {
	case model.RuleTypeDomain, model.RuleTypeDomainSuffix, model.RuleTypeDomainKeyword,
		model.RuleTypeIP, model.RuleTypeSrcIP, model.RuleTypeGeoIP, model.RuleTypeGeoSite,
		model.RuleTypeProcess, model.RuleTypeFinal:
		return true
	default:
		return false
	}
}

// isValidRulePolicy 验证规则策略
func isValidRulePolicy(policy model.RulePolicy) bool {
	switch policy {
	case model.PolicyProxy, model.PolicyDirect, model.PolicyReject, model.PolicyBlock:
		return true
	default:
		return false
	}
}
