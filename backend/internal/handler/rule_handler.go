package handler

import (
	"proxy_server/internal/model"
	"proxy_server/internal/service"
	"proxy_server/pkg/response"

	"github.com/gin-gonic/gin"
)

// RuleHandler 规则处理器
type RuleHandler struct {
	ruleService service.RuleService
}

// NewRuleHandler 创建规则处理器
func NewRuleHandler(ruleService service.RuleService) *RuleHandler {
	return &RuleHandler{
		ruleService: ruleService,
	}
}

// Create 创建规则
func (h *RuleHandler) Create(c *gin.Context) {
	var req model.RuleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	
	rule, err := h.ruleService.Create(&req)
	if err != nil {
		response.Error(c, 3000, err.Error())
		return
	}
	
	response.Success(c, rule)
}

// GetAll 获取所有规则
func (h *RuleHandler) GetAll(c *gin.Context) {
	rules, err := h.ruleService.GetAll()
	if err != nil {
		response.Error(c, 3000, err.Error())
		return
	}
	
	response.Success(c, rules)
}

// GetByID 根据ID获取规则
func (h *RuleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少规则ID")
		return
	}
	
	rule, err := h.ruleService.GetByID(id)
	if err != nil {
		response.Error(c, 3000, err.Error())
		return
	}
	
	response.Success(c, rule)
}

// Update 更新规则
func (h *RuleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少规则ID")
		return
	}
	
	var req model.RuleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	
	rule, err := h.ruleService.Update(id, &req)
	if err != nil {
		response.Error(c, 3000, err.Error())
		return
	}
	
	response.Success(c, rule)
}

// Delete 删除规则
func (h *RuleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少规则ID")
		return
	}
	
	if err := h.ruleService.Delete(id); err != nil {
		response.Error(c, 3000, err.Error())
		return
	}
	
	response.SuccessWithMessage(c, "删除成功", nil)
}
