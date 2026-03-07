package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"proxy_server/internal/model"
	"proxy_server/internal/service"
	"proxy_server/pkg/logger"

	"github.com/gin-gonic/gin"
)

// TrafficLogger 流量日志记录器
type TrafficLogger struct {
	config      *model.LogConfig
	logBuffer   []model.RequestLog
	bufferMutex sync.Mutex
	writeChan   chan model.RequestLog
	stopChan    chan bool
	logService  service.LogService
}

var (
	trafficLogger *TrafficLogger
	loggerOnce    sync.Once
)

// GetTrafficLogger 获取流量日志记录器单例
func GetTrafficLogger() *TrafficLogger {
	loggerOnce.Do(func() {
		trafficLogger = NewTrafficLogger(model.DefaultLogConfig())
	})
	return trafficLogger
}

// NewTrafficLogger 创建新的流量日志记录器
func NewTrafficLogger(config *model.LogConfig) *TrafficLogger {
	tl := &TrafficLogger{
		config:    config,
		logBuffer: make([]model.RequestLog, 0, 1000),
		writeChan: make(chan model.RequestLog, 1000),
		stopChan:  make(chan bool),
	}

	// 启动后台写入goroutine
	go tl.writeWorker()

	return tl
}

// SetLogService 设置日志服务
func (tl *TrafficLogger) SetLogService(ls service.LogService) {
	tl.logService = ls
}

// Start 启动日志记录器
func (tl *TrafficLogger) Start() {
	if tl.config.Enabled {
		logger.Info("流量日志记录器已启动")
	}
}

// Stop 停止日志记录器
func (tl *TrafficLogger) Stop() {
	close(tl.writeChan)
	<-tl.stopChan
	logger.Info("流量日志记录器已停止")
}

// writeWorker 后台写入worker
func (tl *TrafficLogger) writeWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case log, ok := <-tl.writeChan:
			if !ok {
				tl.stopChan <- true
				return
			}
			tl.bufferMutex.Lock()
			tl.logBuffer = append(tl.logBuffer, log)
			if len(tl.logBuffer) >= 100 {
				tl.flushBuffer()
			}
			tl.bufferMutex.Unlock()
		case <-ticker.C:
			tl.bufferMutex.Lock()
			if len(tl.logBuffer) > 0 {
				tl.flushBuffer()
			}
			tl.bufferMutex.Unlock()
		}
	}
}

// flushBuffer 刷新缓冲区
func (tl *TrafficLogger) flushBuffer() {
	if len(tl.logBuffer) == 0 {
		return
	}

	// 保存到日志服务
	for _, log := range tl.logBuffer {
		data, _ := json.Marshal(log)
		logger.Info("TRAFFIC: %s", string(data))

		// 通过WebSocket推送流量日志到前端
		BroadcastTraffic(log)

		// 保存到持久化存储
		if tl.logService != nil {
			tl.logService.AddLog(log)
		}
	}

	tl.logBuffer = tl.logBuffer[:0]
}

// BroadcastTraffic 广播流量日志（调用handler中的广播函数）
var broadcastTrafficFunc func(model.RequestLog)

// SetBroadcastTrafficFunc 设置流量广播函数
func SetBroadcastTrafficFunc(fn func(model.RequestLog)) {
	broadcastTrafficFunc = fn
}

// BroadcastTraffic 广播流量日志
func BroadcastTraffic(log model.RequestLog) {
	if broadcastTrafficFunc != nil {
		broadcastTrafficFunc(log)
	}
}

// sanitize 脱敏处理
func (tl *TrafficLogger) sanitize(value string) string {
	if value == "" {
		return ""
	}

	lowerValue := strings.ToLower(value)
	for _, sensitive := range tl.config.SensitiveWords {
		if strings.Contains(lowerValue, sensitive) {
			// 替换敏感值为****
			return "****"
		}
	}
	return value
}

// sanitizeHeaders 脱敏请求头
func (tl *TrafficLogger) sanitizeHeaders(headers map[string]string) map[string]string {
	sanitized := make(map[string]string)
	for key, value := range headers {
		sanitized[key] = tl.sanitize(value)
	}
	return sanitized
}

// TrafficLoggerMiddleware 流量日志记录中间件
func TrafficLoggerMiddleware() gin.HandlerFunc {
	tl := GetTrafficLogger()
	tl.Start()

	return func(c *gin.Context) {
		if !tl.config.Enabled {
			c.Next()
			return
		}

		// 记录开始时间
		startTime := time.Now()

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 包装响应写入器
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 记录结束时间
		endTime := time.Now()
		duration := endTime.Sub(startTime)

		// 构建请求日志
		log := model.RequestLog{
			ID:           fmt.Sprintf("%d", startTime.UnixNano()),
			Timestamp:    startTime,
			ClientIP:     getClientIP(c),
			Method:       c.Request.Method,
			URL:          c.Request.URL.String(),
			Path:         c.Request.URL.Path,
			QueryString:  c.Request.URL.RawQuery,
			Headers:      getHeaders(c),
			BodySize:     int64(len(requestBody)),
			StatusCode:   c.Writer.Status(),
			ResponseTime: duration.Milliseconds(),
			ResponseSize: int64(blw.body.Len()),
			UserAgent:    c.Request.UserAgent(),
			Protocol:     c.Request.Proto,
		}

		// 发送到写入通道
		select {
		case tl.writeChan <- log:
		default:
			// 通道满，丢弃日志
		}
	}
}

// bodyLogWriter 包装响应写入器
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// getClientIP 获取客户端IP
func getClientIP(c *gin.Context) string {
	// 尝试从X-Forwarded-For获取
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从X-Real-IP获取
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

// getHeaders 获取请求头
func getHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}
