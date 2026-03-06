package service

import (
	"runtime"
	"time"

	"proxy_server/internal/model"
	"proxy_server/internal/repository"
	"proxy_server/pkg/logger"
)

// SystemService 系统服务接口
type SystemService interface {
	GetStatus() (*model.SystemStatus, error)
	GetTraffic() (*model.TrafficStats, error)
	GetLogs(query *model.LogQuery) ([]*model.LogEntry, error)
	AddLog(level, message, node string)
}

// systemService 系统服务实现
type systemService struct {
	systemRepo repository.SystemRepository
	startTime  time.Time
	version    string
}

// NewSystemService 创建系统服务
func NewSystemService(systemRepo repository.SystemRepository) SystemService {
	return &systemService{
		systemRepo: systemRepo,
		startTime:  time.Now(),
		version:    "1.0.0",
	}
}

// GetStatus 获取系统状态
func (s *systemService) GetStatus() (*model.SystemStatus, error) {
	status, err := s.systemRepo.GetStatus()
	if err != nil {
		return nil, err
	}
	
	// 更新运行时信息
	status.Uptime = int64(time.Since(s.startTime).Seconds())
	status.GoroutineCount = runtime.NumGoroutine()
	status.Version = s.version
	
	return status, nil
}

// GetTraffic 获取流量统计
func (s *systemService) GetTraffic() (*model.TrafficStats, error) {
	traffic, err := s.systemRepo.GetTraffic()
	if err != nil {
		return nil, err
	}
	
	// TODO: 实际的流量统计逻辑
	// 这里需要集成代理核心获取实时流量数据
	
	traffic.Timestamp = time.Now()
	
	return traffic, nil
}

// GetLogs 获取日志
func (s *systemService) GetLogs(query *model.LogQuery) ([]*model.LogEntry, error) {
	// 设置默认值
	if query.Limit <= 0 {
		query.Limit = 100
	}
	if query.Limit > 1000 {
		query.Limit = 1000
	}
	
	return s.systemRepo.GetLogs(query)
}

// AddLog 添加日志
func (s *systemService) AddLog(level, message, node string) {
	logEntry := &model.LogEntry{
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
		Node:      node,
	}
	
	if err := s.systemRepo.SaveLog(logEntry); err != nil {
		logger.Error("保存日志失败: %v", err)
	}
}
