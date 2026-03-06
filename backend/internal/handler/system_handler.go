package handler

import (
	"proxy_server/internal/model"
	"proxy_server/internal/service"
	"proxy_server/pkg/response"

	"github.com/gin-gonic/gin"
)

// SystemHandler 系统处理器
type SystemHandler struct {
	systemService service.SystemService
}

// NewSystemHandler 创建系统处理器
func NewSystemHandler(systemService service.SystemService) *SystemHandler {
	return &SystemHandler{
		systemService: systemService,
	}
}

// GetStatus 获取系统状态
func (h *SystemHandler) GetStatus(c *gin.Context) {
	status, err := h.systemService.GetStatus()
	if err != nil {
		response.Error(c, 4000, err.Error())
		return
	}
	
	response.Success(c, status)
}

// GetTraffic 获取流量统计
func (h *SystemHandler) GetTraffic(c *gin.Context) {
	traffic, err := h.systemService.GetTraffic()
	if err != nil {
		response.Error(c, 4000, err.Error())
		return
	}
	
	response.Success(c, traffic)
}

// GetLogs 获取日志
func (h *SystemHandler) GetLogs(c *gin.Context) {
	var query model.LogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	
	logs, err := h.systemService.GetLogs(&query)
	if err != nil {
		response.Error(c, 4000, err.Error())
		return
	}
	
	response.Success(c, logs)
}
