package service

import (
	"context"
	"time"

	"proxy_server/internal/logger"
	"proxy_server/internal/model"
	plogger "proxy_server/pkg/logger"
)

// LogService 日志服务接口
type LogService interface {
	QueryLogs(query *model.LogQuery) ([]*model.RequestLog, int64, error)
	GetLogStats(startTime, endTime *time.Time) (*model.LogStats, error)
	GetTrafficLogs(query *model.TrafficQuery) ([]*model.TrafficLog, int64, error)
	GetTrafficStats(startTime, endTime *time.Time) (*model.LogStats, error)
	ClearLogs(before *time.Time) error
	AddLog(log model.RequestLog)
}

// logService 日志服务实现
type logService struct {
	logConfig  *model.LogConfig
	logRotator *logger.LogRotator
}

// NewLogService 创建日志服务
func NewLogService() LogService {
	return &logService{
		logConfig: model.DefaultLogConfig(),
	}
}

// NewLogServiceWithDataDir 创建日志服务（带数据目录）
func NewLogServiceWithDataDir(dataDir string) LogService {
	config := model.DefaultLogConfig()
	config.Directory = dataDir + "/logs"

	rotator := logger.NewLogRotator(config)
	if err := rotator.Start(); err != nil {
		plogger.Error("启动日志轮转器失败: %v", err)
	}

	return &logService{
		logConfig:  config,
		logRotator: rotator,
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
	// 将流量日志转换为TrafficLog格式
	logQuery := &model.LogQuery{
		Limit: query.Limit,
	}

	if !query.StartTime.IsZero() {
		logQuery.StartTime = &query.StartTime
	}
	if !query.EndTime.IsZero() {
		logQuery.EndTime = &query.EndTime
	}

	logs, err := logger.QueryLogs(context.Background(), s.logConfig, logQuery)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*model.TrafficLog, len(logs))
	for i, log := range logs {
		result[i] = &model.TrafficLog{
			ID:         log.ID,
			Timestamp:  log.Timestamp,
			ClientIP:   log.ClientIP,
			Method:     log.Method,
			Path:       log.Path,
			StatusCode: log.StatusCode,
		}
	}

	return result, int64(len(result)), nil
}

// GetTrafficStats 获取流量统计
func (s *logService) GetTrafficStats(startTime, endTime *time.Time) (*model.LogStats, error) {
	// 使用请求日志统计作为流量统计
	return s.GetLogStats(startTime, endTime)
}

// ClearLogs 清理日志
func (s *logService) ClearLogs(before *time.Time) error {
	// 暂时返回nil，后续实现
	return nil
}

// AddLog 添加日志
func (s *logService) AddLog(log model.RequestLog) {
	if s.logRotator != nil {
		s.logRotator.Write(log)
	}
	plogger.Info("添加流量日志: %s %s", log.Method, log.Path)
}
