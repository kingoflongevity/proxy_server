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
	GetLogs(query *model.SystemLogQuery) ([]*model.LogEntry, error)
	AddLog(level, message, node string)
	GetSettings() (*model.SystemSettings, error)
	UpdateSettings(req *model.UpdateSettingsRequest) (*model.SystemSettings, error)
	GetConnectionStatus() (*model.ConnectionStatus, error)
	GetSystemInfo() (*model.SystemInfo, error)
	RestartService() error
	ExportConfig() (string, error)
	ImportConfig(config string) error
	ClearCache() error
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
	
	traffic.Timestamp = time.Now()
	
	return traffic, nil
}

// GetLogs 获取日志
func (s *systemService) GetLogs(query *model.SystemLogQuery) ([]*model.LogEntry, error) {
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

// GetSettings 获取系统设置
func (s *systemService) GetSettings() (*model.SystemSettings, error) {
	return s.systemRepo.GetSettings()
}

// UpdateSettings 更新系统设置
func (s *systemService) UpdateSettings(req *model.UpdateSettingsRequest) (*model.SystemSettings, error) {
	settings, err := s.systemRepo.GetSettings()
	if err != nil {
		return nil, err
	}
	
	if req.Theme != "" {
		settings.Theme = req.Theme
	}
	if req.Language != "" {
		settings.Language = req.Language
	}
	if req.ProxyMode != "" {
		settings.ProxyMode = req.ProxyMode
	}
	if req.BindAddress != "" {
		settings.BindAddress = req.BindAddress
	}
	if req.Port > 0 {
		settings.Port = req.Port
	}
	if req.SocksPort > 0 {
		settings.SocksPort = req.SocksPort
	}
	if req.HttpPort > 0 {
		settings.HttpPort = req.HttpPort
	}
	if req.LogLevel != "" {
		settings.LogLevel = req.LogLevel
	}
	settings.AutoStart = req.AutoStart
	settings.AllowLan = req.AllowLan
	
	if err := s.systemRepo.SaveSettings(settings); err != nil {
		return nil, err
	}
	
	return settings, nil
}

// GetConnectionStatus 获取连接状态
func (s *systemService) GetConnectionStatus() (*model.ConnectionStatus, error) {
	status, err := s.systemRepo.GetStatus()
	if err != nil {
		return nil, err
	}
	
	return &model.ConnectionStatus{
		Connected:      status.Connected,
		CurrentMode:    status.Mode,
		UploadSpeed:    0,
		DownloadSpeed: 0,
		UploadTotal:    0,
		DownloadTotal: 0,
		ConnectionCount: 0,
	}, nil
}

// GetSystemInfo 获取系统信息
func (s *systemService) GetSystemInfo() (*model.SystemInfo, error) {
	return &model.SystemInfo{
		Version:       s.version,
		GoVersion:    runtime.Version(),
		Os:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		NumCPU:       runtime.NumCPU(),
		GoroutineNum: runtime.NumGoroutine(),
		Uptime:       int64(time.Since(s.startTime).Seconds()),
	}, nil
}

// RestartService 重启服务
func (s *systemService) RestartService() error {
	logger.Info("重启服务请求")
	return nil
}

// ExportConfig 导出配置
func (s *systemService) ExportConfig() (string, error) {
	settings, err := s.systemRepo.GetSettings()
	if err != nil {
		return "", err
	}
	
	return settings.ExportConfig(), nil
}

// ImportConfig 导入配置
func (s *systemService) ImportConfig(config string) error {
	settings, err := s.systemRepo.GetSettings()
	if err != nil {
		return err
	}
	
	if err := settings.ImportConfig(config); err != nil {
		return err
	}
	
	return s.systemRepo.SaveSettings(settings)
}

// ClearCache 清除缓存
func (s *systemService) ClearCache() error {
	logger.Info("清除缓存请求")
	return nil
}
