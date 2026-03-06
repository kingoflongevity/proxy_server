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
	var query model.SystemLogQuery
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

// GetSettings 获取系统设置
func (h *SystemHandler) GetSettings(c *gin.Context) {
	settings, err := h.systemService.GetSettings()
	if err != nil {
		response.Error(c, 4000, err.Error())
		return
	}

	response.Success(c, settings)
}

// UpdateSettings 更新系统设置
func (h *SystemHandler) UpdateSettings(c *gin.Context) {
	var req model.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	settings, err := h.systemService.UpdateSettings(&req)
	if err != nil {
		response.Error(c, 4000, err.Error())
		return
	}

	response.Success(c, settings)
}

// GetConnectionStatus 获取连接状态
func (h *SystemHandler) GetConnectionStatus(c *gin.Context) {
	status, err := h.systemService.GetConnectionStatus()
	if err != nil {
		response.Error(c, 4000, err.Error())
		return
	}

	response.Success(c, status)
}

// GetSystemInfo 获取系统信息
func (h *SystemHandler) GetSystemInfo(c *gin.Context) {
	info, err := h.systemService.GetSystemInfo()
	if err != nil {
		response.Error(c, 4000, err.Error())
		return
	}

	response.Success(c, info)
}

// GetProxyMode 获取当前代理模式
func (h *SystemHandler) GetProxyMode(c *gin.Context) {
	mode := service.GetProxyMode()
	response.Success(c, map[string]string{
		"proxyMode": mode,
	})
}

// RestartService 重启服务
func (h *SystemHandler) RestartService(c *gin.Context) {
	if err := h.systemService.RestartService(); err != nil {
		response.Error(c, 4000, err.Error())
		return
	}

	response.Success(c, nil)
}

// ExportConfig 导出配置
func (h *SystemHandler) ExportConfig(c *gin.Context) {
	config, err := h.systemService.ExportConfig()
	if err != nil {
		response.Error(c, 4000, err.Error())
		return
	}

	response.Success(c, config)
}

// ImportConfig 导入配置
func (h *SystemHandler) ImportConfig(c *gin.Context) {
	var req struct {
		Config string `json:"config" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.systemService.ImportConfig(req.Config); err != nil {
		response.Error(c, 4000, err.Error())
		return
	}

	response.Success(c, nil)
}

// ClearCache 清除缓存
func (h *SystemHandler) ClearCache(c *gin.Context) {
	if err := h.systemService.ClearCache(); err != nil {
		response.Error(c, 4000, err.Error())
		return
	}

	response.Success(c, nil)
}
