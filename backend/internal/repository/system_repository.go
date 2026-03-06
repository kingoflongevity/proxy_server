package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"proxy_server/internal/model"
	"proxy_server/pkg/logger"
)

// SystemRepository 系统仓库接口
type SystemRepository interface {
	SaveStatus(status *model.SystemStatus) error
	GetStatus() (*model.SystemStatus, error)
	SaveTraffic(traffic *model.TrafficStats) error
	GetTraffic() (*model.TrafficStats, error)
	SaveLog(log *model.LogEntry) error
	GetLogs(query *model.SystemLogQuery) ([]*model.LogEntry, error)
	GetSettings() (*model.SystemSettings, error)
	SaveSettings(settings *model.SystemSettings) error
}

// systemRepository 系统仓库实现
type systemRepository struct {
	dataDir  string
	statusMu sync.RWMutex
	trafficMu sync.RWMutex
	logMu    sync.RWMutex
}

// NewSystemRepository 创建系统仓库
func NewSystemRepository(dataDir string) SystemRepository {
	return &systemRepository{
		dataDir: dataDir,
	}
}

// SaveStatus 保存系统状态
func (r *systemRepository) SaveStatus(status *model.SystemStatus) error {
	r.statusMu.Lock()
	defer r.statusMu.Unlock()
	
	dataFile := filepath.Join(r.dataDir, "system_status.json")
	
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(dataFile, data, 0644)
}

// GetStatus 获取系统状态
func (r *systemRepository) GetStatus() (*model.SystemStatus, error) {
	r.statusMu.RLock()
	defer r.statusMu.RUnlock()
	
	dataFile := filepath.Join(r.dataDir, "system_status.json")
	
	data, err := os.ReadFile(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &model.SystemStatus{
				Connected:  false,
				StartTime:  time.Now(),
				Version:    "1.0.0",
			}, nil
		}
		return nil, err
	}
	
	var status model.SystemStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}
	
	return &status, nil
}

// SaveTraffic 保存流量统计
func (r *systemRepository) SaveTraffic(traffic *model.TrafficStats) error {
	r.trafficMu.Lock()
	defer r.trafficMu.Unlock()
	
	dataFile := filepath.Join(r.dataDir, "traffic_stats.json")
	
	data, err := json.MarshalIndent(traffic, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(dataFile, data, 0644)
}

// GetTraffic 获取流量统计
func (r *systemRepository) GetTraffic() (*model.TrafficStats, error) {
	r.trafficMu.RLock()
	defer r.trafficMu.RUnlock()
	
	dataFile := filepath.Join(r.dataDir, "traffic_stats.json")
	
	data, err := os.ReadFile(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &model.TrafficStats{
				Timestamp: time.Now(),
			}, nil
		}
		return nil, err
	}
	
	var traffic model.TrafficStats
	if err := json.Unmarshal(data, &traffic); err != nil {
		return nil, err
	}
	
	return &traffic, nil
}

// SaveLog 保存日志
func (r *systemRepository) SaveLog(logEntry *model.LogEntry) error {
	r.logMu.Lock()
	defer r.logMu.Unlock()
	
	dataFile := filepath.Join(r.dataDir, "logs.json")
	
	// 读取现有日志
	var logs []*model.LogEntry
	data, err := os.ReadFile(dataFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		logs = []*model.LogEntry{}
	} else {
		if err := json.Unmarshal(data, &logs); err != nil {
			logger.Error("解析日志文件失败: %v", err)
			logs = []*model.LogEntry{}
		}
	}
	
	// 添加新日志
	logs = append(logs, logEntry)
	
	// 限制日志数量（保留最新的1000条）
	if len(logs) > 1000 {
		logs = logs[len(logs)-1000:]
	}
	
	// 保存日志
	data, err = json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(dataFile, data, 0644)
}

// GetLogs 获取日志
func (r *systemRepository) GetLogs(query *model.SystemLogQuery) ([]*model.LogEntry, error) {
	r.logMu.RLock()
	defer r.logMu.RUnlock()
	
	dataFile := filepath.Join(r.dataDir, "logs.json")
	
	data, err := os.ReadFile(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*model.LogEntry{}, nil
		}
		return nil, err
	}
	
	var logs []*model.LogEntry
	if err := json.Unmarshal(data, &logs); err != nil {
		return nil, err
	}
	
	// 过滤日志
	var result []*model.LogEntry
	for _, log := range logs {
		// 按级别过滤
		if query.Level != "" && log.Level != query.Level {
			continue
		}
		
		// 按节点过滤
		if query.Node != "" && log.Node != query.Node {
			continue
		}
		
		// 按关键字过滤
		if query.Keyword != "" && !contains(log.Message, query.Keyword) {
			continue
		}
		
		result = append(result, log)
	}
	
	// 分页
	if query.Offset > 0 && query.Offset < len(result) {
		result = result[query.Offset:]
	}
	
	if query.Limit > 0 && query.Limit < len(result) {
		result = result[:query.Limit]
	}
	
	return result, nil
}

// contains 检查字符串是否包含子串（不区分大小写）
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(str) > 0 && containsHelper(str, substr))
}

func containsHelper(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (r *systemRepository) GetSettings() (*model.SystemSettings, error) {
	dataFile := filepath.Join(r.dataDir, "settings.json")
	
	data, err := os.ReadFile(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &model.SystemSettings{
				Theme:       "dark",
				Language:    "zh-CN",
				AutoStart:   false,
				SilentStart: false,
				AllowLan:    false,
				BindAddress: "127.0.0.1",
				Port:        7890,
				SocksPort:   7891,
				HttpPort:    7892,
				MixedPort:   7893,
				LogLevel:    "info",
				ProxyMode:   "rule",
			}, nil
		}
		return nil, err
	}
	
	var settings model.SystemSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}
	
	return &settings, nil
}

func (r *systemRepository) SaveSettings(settings *model.SystemSettings) error {
	dataFile := filepath.Join(r.dataDir, "settings.json")
	
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(dataFile, data, 0644)
}
