package logger

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"proxy_server/internal/model"
)

// LogRotator 日志轮转器
type LogRotator struct {
	config       *model.LogConfig
	logFile      *os.File
	logEncoder   *json.Encoder
	rotateMutex  sync.Mutex
	currentDate  string
}

// NewLogRotator 创建日志轮转器
func NewLogRotator(config *model.LogConfig) *LogRotator {
	return &LogRotator{
		config:      config,
		currentDate: time.Now().Format("2006-01-02"),
	}
}

// Start 启动日志轮转器
func (lr *LogRotator) Start() error {
	if err := lr.ensureDirectory(); err != nil {
		return err
	}

	if err := lr.openLogFile(); err != nil {
		return err
	}

	// 启动定期检查
	go lr.checkRotation()

	return nil
}

// ensureDirectory 确保目录存在
func (lr *LogRotator) ensureDirectory() error {
	return os.MkdirAll(lr.config.Directory, 0755)
}

// openLogFile 打开日志文件
func (lr *LogRotator) openLogFile() error {
	filename := filepath.Join(lr.config.Directory, fmt.Sprintf("%s.json", lr.currentDate))
	
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	lr.logFile = file
	lr.logEncoder = json.NewEncoder(file)
	lr.currentDate = time.Now().Format("2006-01-02")

	return nil
}

// Write 写入日志
func (lr *LogRotator) Write(log model.RequestLog) error {
	lr.rotateMutex.Lock()
	defer lr.rotateMutex.Unlock()

	// 检查是否需要轮转
	if lr.needRotation() {
		if err := lr.rotate(); err != nil {
			return err
		}
	}

	return lr.logEncoder.Encode(log)
}

// needRotation 检查是否需要轮转
func (lr *LogRotator) needRotation() bool {
	// 按日期轮转
	currentDate := time.Now().Format("2006-01-02")
	if currentDate != lr.currentDate {
		return true
	}

	// 按大小轮转
	if lr.config.MaxFileSize > 0 {
		stat, err := lr.logFile.Stat()
		if err == nil && stat.Size() >= lr.config.MaxFileSize {
			return true
		}
	}

	return false
}

// rotate 执行轮转
func (lr *LogRotator) rotate() error {
	// 关闭当前文件
	if lr.logFile != nil {
		lr.logFile.Close()
	}

	// 压缩旧文件
	if lr.config.Compress {
		go lr.compressOldFiles()
	}

	// 删除过期文件
	if lr.config.MaxBackups > 0 {
		lr.cleanOldFiles()
	}

	// 打开新文件
	return lr.openLogFile()
}

// compressOldFiles 压缩旧文件
func (lr *LogRotator) compressOldFiles() {
	files, err := filepath.Glob(filepath.Join(lr.config.Directory, "*.json"))
	if err != nil {
		return
	}

	for _, file := range files {
		gzFile := file + ".gz"
		if _, err := os.Stat(gzFile); err == nil {
			continue // 已压缩
		}

		// 压缩文件
		lr.compressFile(file, gzFile)
	}
}

// compressFile 压缩单个文件
func (lr *LogRotator) compressFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	writer := gzip.NewWriter(dstFile)
	defer writer.Close()

	_, err = io.Copy(writer, srcFile)
	if err != nil {
		return err
	}

	// 删除原文件
	os.Remove(src)

	return nil
}

// cleanOldFiles 清理旧文件
func (lr *LogRotator) cleanOldFiles() {
	files, err := filepath.Glob(filepath.Join(lr.config.Directory, "*.json"))
	if err != nil {
		return
	}

	if len(files) <= lr.config.MaxBackups {
		return
	}

	// 按修改时间排序
	sort.Slice(files, func(i, j int) bool {
		iInfo, _ := os.Stat(files[i])
		jInfo, _ := os.Stat(files[j])
		return iInfo.ModTime().Before(jInfo.ModTime())
	})

	// 删除最老的文件
	for i := 0; i < len(files)-lr.config.MaxBackups; i++ {
		os.Remove(files[i])
	}
}

// checkRotation 定期检查轮转
func (lr *LogRotator) checkRotation() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		lr.rotateMutex.Lock()
		if lr.needRotation() {
			lr.rotate()
		}
		lr.rotateMutex.Unlock()
	}
}

// Stop 停止日志轮转器
func (lr *LogRotator) Stop() error {
	if lr.logFile != nil {
		return lr.logFile.Close()
	}
	return nil
}

// QueryLogs 查询日志
func QueryLogs(ctx context.Context, config *model.LogConfig, query *model.LogQuery) ([]model.RequestLog, error) {
	var logs []model.RequestLog

	// 收集所有日志文件
	pattern := filepath.Join(config.Directory, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return logs, err
	}

	// 也检查压缩文件
	gzPattern := filepath.Join(config.Directory, "*.json.gz")
	gzFiles, _ := filepath.Glob(gzPattern)
	files = append(files, gzFiles...)

	// 按时间过滤
	startTime := query.StartTime
	endTime := query.EndTime

	for _, file := range files {
		fileLogs, err := readLogFile(file, startTime, endTime)
		if err != nil {
			continue
		}
		logs = append(logs, fileLogs...)
	}

	// 应用过滤条件
	logs = filterLogs(logs, query)

	// 分页
	start := query.Offset
	end := start + query.Limit
	if start >= len(logs) {
		return []model.RequestLog{}, nil
	}
	if end > len(logs) {
		end = len(logs)
	}

	return logs[start:end], nil
}

// readLogFile 读取日志文件
func readLogFile(filename string, startTime, endTime *time.Time) ([]model.RequestLog, error) {
	var logs []model.RequestLog

	file, err := os.Open(filename)
	if err != nil {
		return logs, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		var log model.RequestLog
		if err := decoder.Decode(&log); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		// 时间过滤
		if startTime != nil && log.Timestamp.Before(*startTime) {
			continue
		}
		if endTime != nil && log.Timestamp.After(*endTime) {
			continue
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// filterLogs 过滤日志
func filterLogs(logs []model.RequestLog, query *model.LogQuery) []model.RequestLog {
	var filtered []model.RequestLog

	for _, log := range logs {
		// 按客户端IP过滤
		if query.ClientIP != "" && !strings.Contains(log.ClientIP, query.ClientIP) {
			continue
		}

		// 按方法过滤
		if query.Method != "" && log.Method != query.Method {
			continue
		}

		// 按状态码过滤
		if query.StatusCode > 0 && log.StatusCode != query.StatusCode {
			continue
		}

		// 按URL过滤
		if query.URL != "" && !strings.Contains(log.URL, query.URL) {
			continue
		}

		// 按关键词过滤
		if query.Keyword != "" {
			keyword := strings.ToLower(query.Keyword)
			if !strings.Contains(strings.ToLower(log.URL), keyword) &&
			   !strings.Contains(strings.ToLower(log.Error), keyword) {
				continue
			}
		}

		// 按用户ID过滤
		if query.UserID != "" && log.UserID != query.UserID {
			continue
		}

		filtered = append(filtered, log)
	}

	// 按时间倒序
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	return filtered
}

// GetLogStats 获取日志统计
func GetLogStats(ctx context.Context, config *model.LogConfig, query *model.LogQuery) (*model.LogStats, error) {
	logs, err := QueryLogs(ctx, config, &model.LogQuery{
		StartTime: query.StartTime,
		EndTime:   query.EndTime,
		Limit:     10000,
	})
	if err != nil {
		return nil, err
	}

	stats := &model.LogStats{
		StatusCounts: make(map[string]int64),
	}

	var totalResponseTime int64

	for _, log := range logs {
		stats.TotalRequests++
		stats.TotalTraffic += log.ResponseSize
		stats.UploadBytes += log.BodySize
		stats.DownloadBytes += log.ResponseSize
		totalResponseTime += log.ResponseTime

		// 状态码统计
		statusStr := fmt.Sprintf("%d", log.StatusCode)
		stats.StatusCounts[statusStr]++
	}

	if stats.TotalRequests > 0 {
		stats.AvgResponseTime = totalResponseTime / stats.TotalRequests
	}

	return stats, nil
}
