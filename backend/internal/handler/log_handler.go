package handler

import (
	"time"

	"proxy_server/internal/model"
	"proxy_server/internal/service"
	"proxy_server/pkg/response"

	"github.com/gin-gonic/gin"
)

// LogHandler 日志处理器
type LogHandler struct {
	logService service.LogService
}

// NewLogHandler 创建日志处理器
func NewLogHandler(logService service.LogService) *LogHandler {
	return &LogHandler{
		logService: logService,
	}
}

// QueryLogs 查询日志
func (h *LogHandler) QueryLogs(c *gin.Context) {
	var query model.LogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 解析时间参数
	if startTime := c.Query("start_time"); startTime != "" {
		t, err := time.Parse(time.RFC3339, startTime)
		if err == nil {
			query.StartTime = &t
		}
	}
	if endTime := c.Query("end_time"); endTime != "" {
		t, err := time.Parse(time.RFC3339, endTime)
		if err == nil {
			query.EndTime = &t
		}
	}

	logs, total, err := h.logService.QueryLogs(&query)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}

	response.Success(c, gin.H{
		"logs":  logs,
		"total": total,
	})
}

// GetLogStats 获取日志统计
func (h *LogHandler) GetLogStats(c *gin.Context) {
	var query struct {
		StartTime *time.Time `form:"start_time"`
		EndTime   *time.Time `form:"end_time"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	stats, err := h.logService.GetLogStats(query.StartTime, query.EndTime)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}

	response.Success(c, stats)
}

// GetTrafficLogs 查询流量日志
func (h *LogHandler) GetTrafficLogs(c *gin.Context) {
	var query model.TrafficQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if query.Limit <= 0 {
		query.Limit = 100
	}
	if query.Limit > 1000 {
		query.Limit = 1000
	}

	logs, total, err := h.logService.GetTrafficLogs(&query)
	if err != nil {
		response.Error(c, 5002, err.Error())
		return
	}

	response.Success(c, gin.H{
		"logs":  logs,
		"total": total,
	})
}

// GetTrafficStats 获取流量统计
func (h *LogHandler) GetTrafficStats(c *gin.Context) {
	var query struct {
		StartTime *time.Time `form:"start_time"`
		EndTime   *time.Time `form:"end_time"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	stats, err := h.logService.GetTrafficStats(query.StartTime, query.EndTime)
	if err != nil {
		response.Error(c, 5002, err.Error())
		return
	}

	response.Success(c, stats)
}

// ClearLogs 清理日志
func (h *LogHandler) ClearLogs(c *gin.Context) {
	var query struct {
		Before string `form:"before" json:"before"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var before *time.Time
	if query.Before != "" {
		t, err := time.Parse(time.RFC3339, query.Before)
		if err == nil {
			before = &t
		}
	}

	if err := h.logService.ClearLogs(before); err != nil {
		response.Error(c, 5003, err.Error())
		return
	}

	response.SuccessWithMessage(c, "日志清理成功", nil)
}
