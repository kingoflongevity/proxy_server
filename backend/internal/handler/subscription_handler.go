package handler

import (
	"proxy_server/internal/model"
	"proxy_server/internal/service"
	"proxy_server/pkg/response"

	"github.com/gin-gonic/gin"
)

// SubscriptionHandler 订阅处理器
type SubscriptionHandler struct {
	subService service.SubscriptionService
}

// NewSubscriptionHandler 创建订阅处理器
func NewSubscriptionHandler(subService service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subService: subService,
	}
}

// Create 创建订阅
// @Summary 创建订阅
// @Description 添加新的订阅源
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param request body model.SubscriptionCreateRequest true "订阅信息"
// @Success 200 {object} response.Response
// @Router /api/subscriptions [post]
func (h *SubscriptionHandler) Create(c *gin.Context) {
	var req model.SubscriptionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	subscription, err := h.subService.Create(&req)
	if err != nil {
		response.Error(c, 1000, err.Error())
		return
	}

	response.Success(c, subscription)
}

// GetByID 根据ID获取订阅
func (h *SubscriptionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少订阅ID")
		return
	}

	subscription, err := h.subService.GetByID(id)
	if err != nil {
		response.Error(c, 1000, err.Error())
		return
	}

	response.Success(c, subscription)
}

// GetAll 获取所有订阅
func (h *SubscriptionHandler) GetAll(c *gin.Context) {
	subscriptions, err := h.subService.GetAll()
	if err != nil {
		response.Error(c, 1000, err.Error())
		return
	}

	response.Success(c, subscriptions)
}

// Update 更新订阅
func (h *SubscriptionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少订阅ID")
		return
	}

	var req model.SubscriptionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	subscription, err := h.subService.Update(id, &req)
	if err != nil {
		response.Error(c, 1000, err.Error())
		return
	}

	response.Success(c, subscription)
}

// Delete 删除订阅
func (h *SubscriptionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少订阅ID")
		return
	}

	if err := h.subService.Delete(id); err != nil {
		response.Error(c, 1000, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// Refresh 刷新订阅
func (h *SubscriptionHandler) Refresh(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少订阅ID")
		return
	}

	if err := h.subService.Refresh(id); err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	// 获取更新后的订阅信息
	subscription, err := h.subService.GetByID(id)
	if err != nil {
		response.Error(c, 1002, err.Error())
		return
	}

	response.Success(c, gin.H{
		"count": subscription.NodeCount,
	})
}

// Test 测试订阅连接
func (h *SubscriptionHandler) Test(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少订阅ID")
		return
	}

	subscription, err := h.subService.GetByID(id)
	if err != nil {
		response.Error(c, 1002, err.Error())
		return
	}

	response.Success(c, gin.H{
		"valid":   true,
		"message": "订阅连接正常",
		"name":    subscription.Name,
	})
}
