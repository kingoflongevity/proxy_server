package handler

import (
	"proxy_server/internal/model"
	"proxy_server/internal/service"
	"proxy_server/pkg/response"

	"github.com/gin-gonic/gin"
)

// NodeHandler 节点处理器
type NodeHandler struct {
	nodeService service.NodeService
}

// NewNodeHandler 创建节点处理器
func NewNodeHandler(nodeService service.NodeService) *NodeHandler {
	return &NodeHandler{
		nodeService: nodeService,
	}
}

// GetList 获取节点列表
func (h *NodeHandler) GetList(c *gin.Context) {
	var query model.NodeListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	
	nodes, total, err := h.nodeService.GetList(&query)
	if err != nil {
		response.Error(c, 2000, err.Error())
		return
	}
	
	response.Page(c, nodes, total, query.Page, query.PageSize)
}

// GetByID 根据ID获取节点
func (h *NodeHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少节点ID")
		return
	}
	
	node, err := h.nodeService.GetByID(id)
	if err != nil {
		response.Error(c, 2000, err.Error())
		return
	}
	
	response.Success(c, node)
}

// Update 更新节点
func (h *NodeHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少节点ID")
		return
	}
	
	var req model.NodeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	
	node, err := h.nodeService.Update(id, &req)
	if err != nil {
		response.Error(c, 2000, err.Error())
		return
	}
	
	response.Success(c, node)
}

// Test 测试节点
func (h *NodeHandler) Test(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少节点ID")
		return
	}
	
	latency, err := h.nodeService.Test(id)
	if err != nil {
		response.Error(c, 2001, err.Error())
		return
	}
	
	response.Success(c, map[string]interface{}{
		"latency": latency,
		"message": "测试成功",
	})
}

// Connect 连接节点
func (h *NodeHandler) Connect(c *gin.Context) {
	var req model.NodeConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	
	if err := h.nodeService.Connect(req.NodeID); err != nil {
		response.Error(c, 2002, err.Error())
		return
	}
	
	response.SuccessWithMessage(c, "连接成功", nil)
}

// Disconnect 断开连接
func (h *NodeHandler) Disconnect(c *gin.Context) {
	if err := h.nodeService.Disconnect(); err != nil {
		response.Error(c, 2003, err.Error())
		return
	}

	response.SuccessWithMessage(c, "断开成功", nil)
}

// GetCurrent 获取当前连接的节点
func (h *NodeHandler) GetCurrent(c *gin.Context) {
	node := h.nodeService.GetCurrentNode()
	if node == nil {
		response.Success(c, nil)
		return
	}

	response.Success(c, node)
}
