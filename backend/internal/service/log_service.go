package service

import (
	"context"
	"time"

	"proxy_server/internal/logger"
	"proxy_server/internal/model"
)

// LogService 日志服务接口
type LogService interface {
	QueryLogs(query *model.LogQuery) ([]*model.RequestLog, int64, error)
	GetLogStats(startTime, endTime *time.Time) (*model.LogStats, error)
	GetTrafficLogs(query *model.TrafficQuery) ([]*model.TrafficLog, int64, error)
	GetTrafficStats(startTime, endTime *time.Time) (*model.LogStats, error)
	ClearLogs(before *time.Time) error
}

// logService 日志服务实现
type logService struct {
	logConfig *model.LogConfig
}

// NewLogService 创建日志服务
func NewLogService() LogService {
	return &logService{
		logConfig: model.DefaultLogConfig(),
	}
}

// QueryLogs 查询请求日志
func (s *logService) QueryLogs(query *model.LogQuery) ([]*model.RequestLog, int64, error) {
	if query.Limit <= 0 {
		query.Limit = 100
	}
	if query.Limit > 1000 {
		query.Limit = 1000
	}

	logs, err := logger.QueryLogs(context.Background(), s.logConfig, query)
	if err != nil {
		return nil, 0, err
	}

	// 转换为指针切片
	result := make([]*model.RequestLog, len(logs))
	for i := range logs {
		result[i] = &logs[i]
	}

	return result, int64(len(result)), nil
}

// GetLogStats 获取日志统计
func (s *logService) GetLogStats(startTime, endTime *time.Time) (*model.LogStats, error) {
	query := &model.LogQuery{
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     10000,
	}

	return logger.GetLogStats(context.Background(), s.logConfig, query)
}

// GetTrafficLogs 查询流量日志
func (s *logService) GetTrafficLogs(query *model.TrafficQuery) ([]*model.TrafficLog, int64, error) {
	// 暂时返回空列表，实际实现需要从代理服务获取
	return []*model.TrafficLog{}, 0, nil
}

// GetTrafficStats 获取流量统计
func (s *logService) GetTrafficStats(startTime, endTime *time.Time) (*model.LogStats, error) {
	// 使用请求日志统计作为流量统计
	return s.GetLogStats(startTime, endTime)
}

// ClearLogs 清理日志
func (s *logService) ClearLogs(before *time.Time) error {
	// 实现日志清理逻辑
	return nil
}
