package service

import (
	"fmt"
	"runtime"
	"time"

	"proxy_server/internal/model"
	"proxy_server/internal/repository"
	"proxy_server/pkg/logger"
	"proxy_server/pkg/sysproxy"
)

// currentProxyMode 当前代理模式（全局变量，由node_service设置）
var currentProxyMode = "rule"

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
	GetProxyMode() string
	SetProxyMode(mode string) error
	EnableSystemProxy() error
	DisableSystemProxy() error
	GetSystemProxyStatus() (*sysproxy.ProxyConfig, error)
}

// 获取当前代理模式
func GetProxyMode() string {
	return currentProxyMode
}

// systemService 系统服务实现
type systemService struct {
	systemRepo       repository.SystemRepository
	nodeService      NodeService
	startTime        time.Time
	version          string
	proxyManager     *sysproxy.SystemProxyManager
}

// NewSystemService 创建系统服务
func NewSystemService(systemRepo repository.SystemRepository, nodeService NodeService) SystemService {
	currentProxyMode = "rule"
	return &systemService{
		systemRepo:   systemRepo,
		nodeService:  nodeService,
		startTime:    time.Now(),
		version:      "1.0.0",
		proxyManager: sysproxy.NewSystemProxyManager(),
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
		currentProxyMode = req.ProxyMode
		logger.Info("代理模式已切换为: %s", req.ProxyMode)
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
	if req.MixedPort > 0 {
		settings.MixedPort = req.MixedPort
	}
	if req.LogLevel != "" {
		settings.LogLevel = req.LogLevel
	}
	// 高级设置
	if req.DNSServers != nil {
		settings.DNSServers = req.DNSServers
	}
	if req.EnableMux {
		settings.EnableMux = req.EnableMux
	}
	if req.EnableIpv6 {
		settings.EnableIpv6 = req.EnableIpv6
	}
	if req.DomainStrategy != "" {
		settings.DomainStrategy = req.DomainStrategy
	}
	if req.TunMode {
		settings.TunMode = req.TunMode
	}

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

	// 获取流量统计
	traffic := s.nodeService.GetTraffic()

	return &model.ConnectionStatus{
		Connected:       status.Connected,
		CurrentMode:     status.Mode,
		UploadSpeed:     uint64(traffic.UploadSpeed),
		DownloadSpeed:   uint64(traffic.DownloadSpeed),
		UploadTotal:     uint64(traffic.UploadTotal),
		DownloadTotal:   uint64(traffic.DownloadTotal),
		ConnectionCount: 0,
	}, nil
}

// GetSystemInfo 获取系统信息
func (s *systemService) GetSystemInfo() (*model.SystemInfo, error) {
	return &model.SystemInfo{
		Version:      s.version,
		GoVersion:    runtime.Version(),
		Os:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		NumCPU:       runtime.NumCPU(),
		GoroutineNum: runtime.NumGoroutine(),
		Uptime:       int64(time.Since(s.startTime).Seconds()),
	}, nil
}

// GetProxyMode 获取代理模式
func (s *systemService) GetProxyMode() string {
	return currentProxyMode
}

// SetProxyMode 设置代理模式
func (s *systemService) SetProxyMode(mode string) error {
	if mode != "global" && mode != "rule" && mode != "direct" {
		return fmt.Errorf("无效的代理模式: %s", mode)
	}

	oldMode := currentProxyMode
	currentProxyMode = mode

	settings, err := s.systemRepo.GetSettings()
	if err != nil {
		return err
	}
	settings.ProxyMode = mode
	if err := s.systemRepo.SaveSettings(settings); err != nil {
		currentProxyMode = oldMode
		return err
	}

	logger.Info("代理模式已切换为: %s", mode)
	return nil
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

// EnableSystemProxy 启用系统代理
func (s *systemService) EnableSystemProxy() error {
	settings, err := s.systemRepo.GetSettings()
	if err != nil {
		return err
	}

	httpPort := settings.HttpPort
	if httpPort == 0 {
		httpPort = 10809
	}

	bindAddress := settings.BindAddress
	if bindAddress == "" || bindAddress == "0.0.0.0" {
		bindAddress = "127.0.0.1"
	}

	if err := s.proxyManager.EnableSystemProxy(bindAddress, httpPort); err != nil {
		logger.Error("启用系统代理失败: %v", err)
		return err
	}

	logger.Info("系统代理已启用: %s:%d", bindAddress, httpPort)
	return nil
}

// DisableSystemProxy 禁用系统代理
func (s *systemService) DisableSystemProxy() error {
	if err := s.proxyManager.DisableSystemProxy(); err != nil {
		logger.Error("禁用系统代理失败: %v", err)
		return err
	}

	logger.Info("系统代理已禁用")
	return nil
}

// GetSystemProxyStatus 获取系统代理状态
func (s *systemService) GetSystemProxyStatus() (*sysproxy.ProxyConfig, error) {
	return s.proxyManager.GetCurrentProxy()
}
