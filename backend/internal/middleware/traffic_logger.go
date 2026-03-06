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
	"proxy_server/pkg/logger"

	"github.com/gin-gonic/gin"
)

// TrafficLogger 流量日志记录器
type TrafficLogger struct {
	config       *model.LogConfig
	logBuffer    []model.RequestLog
	bufferMutex  sync.Mutex
	writeChan    chan model.RequestLog
	stopChan     chan bool
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
		case log := <-tl.writeChan:
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

// flushBuffer 刷新缓冲区到文件
func (tl *TrafficLogger) flushBuffer() {
	if len(tl.logBuffer) == 0 {
		return
	}

	// 追加模式写入
	for _, log := range tl.logBuffer {
		data, _ := json.Marshal(log)
		logger.Info("TRAFFIC: %s", string(data))
	}

	tl.logBuffer = tl.logBuffer[:0]
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
		sanitizedValue := tl.sanitize(value)
		// 对敏感key也进行脱敏
		sanitizedKey := tl.sanitize(key)
		if sanitizedKey == "****" {
			sanitized[key] = "****"
		} else {
			sanitized[key] = sanitizedValue
		}
	}
	return sanitized
}

// RecordRequest 记录请求
func (tl *TrafficLogger) RecordRequest(log model.RequestLog) {
	if !tl.config.Enabled {
		return
	}

	// 脱敏处理
	log.Headers = tl.sanitizeHeaders(log.Headers)
	if log.Body != "" {
		log.Body = tl.sanitize(log.Body)
	}
	if log.UserID != "" {
		log.UserID = tl.sanitize(log.UserID)
	}

	select {
	case tl.writeChan <- log:
	default:
		// 通道满，丢弃日志
	}
}

// TrafficLoggerMiddleware 创建流量日志中间件
func TrafficLoggerMiddleware() gin.HandlerFunc {
	tl := GetTrafficLogger()
	tl.Start()

	return func(c *gin.Context) {
		// 跳过健康检查和不需要记录的路径
		if c.Request.URL.Path == "/health" || 
		   strings.HasPrefix(c.Request.URL.Path, "/logs") {
			c.Next()
			return
		}

		startTime := time.Now()
		clientIP := GetClientIP(c)

		// 读取请求体
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			requestBody = string(bodyBytes)
			// 重新填充Body，以便后续处理可以读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 捕获响应
		blw := &bodyLogWriter{ResponseWriter: c.Writer, body: bytes.NewBufferString("")}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 计算响应时间
		responseTime := time.Since(startTime).Milliseconds()

		// 构建日志
		log := model.RequestLog{
			ID:           fmt.Sprintf("%d", time.Now().UnixNano()),
			Timestamp:    startTime,
			ClientIP:     clientIP,
			Method:       c.Request.Method,
			URL:          c.Request.URL.String(),
			Path:         c.Request.URL.Path,
			QueryString:  c.Request.URL.RawQuery,
			Headers:      tl.sanitizeHeaders(flattenHeaders(c.Request.Header)),
			Body:         requestBody,
			BodySize:     int64(len(requestBody)),
			UserAgent:    c.Request.UserAgent(),
			Protocol:     c.Request.Proto,
			StatusCode:   blw.StatusCode,
			ResponseTime: responseTime,
			ResponseSize: int64(blw.body.Len()),
		}

		// 记录错误
		if len(c.Errors) > 0 {
			log.Error = c.Errors.String()
		}

		// 记录日志
		tl.RecordRequest(log)
	}
}

// bodyLogWriter 捕获响应体的Writer
type bodyLogWriter struct {
	gin.ResponseWriter
	body      *bytes.Buffer
	StatusCode int
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func (w *bodyLogWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// flattenHeaders 将Header转换为map
func flattenHeaders(h map[string][]string) map[string]string {
	result := make(map[string]string)
	for key, values := range h {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

// GetClientIP 获取客户端IP
func GetClientIP(c *gin.Context) string {
	// 优先从X-Forwarded-For获取
	forwarded := c.GetHeader("X-Forwarded-For")
	if forwarded != "" {
		// 取第一个IP
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 从X-Real-IP获取
	realIP := c.GetHeader("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// 从RemoteAddr获取
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}
