package xray

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"proxy_server/internal/model"
	"proxy_server/pkg/broadcaster"
	"proxy_server/pkg/logger"
)

// ProcessManager Xray进程管理器
type ProcessManager struct {
	mu          sync.RWMutex
	cmd         *exec.Cmd
	cancel      context.CancelFunc
	configPath  string
	xrayPath    string
	running     bool
	currentNode *model.Node
	logChan     chan string
	statsClient *StatsClient
	localPort   int
	proxyMode   string
	rules       []*model.Rule
	lastUpload   int64
	lastDownload int64
	lastTime     time.Time
}

// NewProcessManager 创建进程管理器
func NewProcessManager(xrayPath string) *ProcessManager {
	if xrayPath == "" {
		xrayPath = "xray"
	}

	return &ProcessManager{
		xrayPath: xrayPath,
		logChan:  make(chan string, 100),
	}
}

// Start 启动Xray进程
// 参数：
//   - node: 要连接的节点
//   - localPort: 本地SOCKS5端口
//   - proxyMode: 代理模式 (global/rule/direct)
//   - rules: 路由规则列表
//
// 返回：
//   - error: 错误信息
func (pm *ProcessManager) Start(node *model.Node, localPort int, proxyMode string, rules []*model.Rule) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.running {
		return fmt.Errorf("Xray进程已在运行")
	}

	generator := NewConfigGenerator(localPort)
	generator.SetProxyMode(proxyMode)
	if rules != nil && len(rules) > 0 {
		generator.SetRules(rules)
	}

	config, err := generator.GenerateConfig(node)
	if err != nil {
		return fmt.Errorf("生成配置失败: %w", err)
	}

	configPath := filepath.Join(os.TempDir(), fmt.Sprintf("xray_config_%d.json", time.Now().Unix()))
	configData, err := config.ToJSON()
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	pm.configPath = configPath
	pm.localPort = localPort
	pm.proxyMode = proxyMode
	pm.rules = rules

	ctx, cancel := context.WithCancel(context.Background())
	pm.cancel = cancel

	pm.cmd = exec.CommandContext(ctx, pm.xrayPath, "run", "-c", configPath)

	stdout, err := pm.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("获取stdout失败: %w", err)
	}

	stderr, err := pm.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("获取stderr失败: %w", err)
	}

	if err := pm.cmd.Start(); err != nil {
		return fmt.Errorf("启动Xray进程失败: %w", err)
	}

	pm.running = true
	pm.currentNode = node
	pm.lastTime = time.Now()

	pm.statsClient = NewStatsClient(localPort + 2)

	go pm.readOutput(stdout, "stdout")
	go pm.readOutput(stderr, "stderr")

	go pm.wait()

	logger.Info("Xray进程已启动，节点: %s, 本地端口: %d, 代理模式: %s", node.Name, localPort, proxyMode)
	return nil
}

// Stop 停止Xray进程
func (pm *ProcessManager) Stop() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.running {
		return nil
	}

	if pm.cancel != nil {
		pm.cancel()
	}

	if pm.cmd != nil && pm.cmd.Process != nil {
		if runtime.GOOS == "windows" {
			pm.cmd.Process.Kill()
		} else {
			pm.cmd.Process.Signal(os.Interrupt)
		}

		done := make(chan error, 1)
		go func() {
			done <- pm.cmd.Wait()
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			pm.cmd.Process.Kill()
		}
	}

	if pm.configPath != "" {
		os.Remove(pm.configPath)
		pm.configPath = ""
	}

	pm.running = false
	pm.currentNode = nil
	pm.cmd = nil

	logger.Info("Xray进程已停止")
	return nil
}

// Restart 重启Xray进程
func (pm *ProcessManager) Restart(node *model.Node, localPort int, proxyMode string, rules []*model.Rule) error {
	if err := pm.Stop(); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	return pm.Start(node, localPort, proxyMode, rules)
}

// IsRunning 检查进程是否运行中
func (pm *ProcessManager) IsRunning() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.running
}

// GetCurrentNode 获取当前连接的节点
func (pm *ProcessManager) GetCurrentNode() *model.Node {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.currentNode
}

// GetProxyMode 获取当前代理模式
func (pm *ProcessManager) GetProxyMode() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	if pm.proxyMode == "" {
		return "rule"
	}
	return pm.proxyMode
}

// GetRules 获取当前路由规则
func (pm *ProcessManager) GetRules() []*model.Rule {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.rules
}

// GetLogs 获取日志通道
func (pm *ProcessManager) GetLogs() <-chan string {
	return pm.logChan
}

// readOutput 读取进程输出并解析流量日志
// Xray访问日志格式: 2024/01/01 12:00:00 [Info] [socks-in] 192.168.1.100:12345 accepted tcp:google.com:443
func (pm *ProcessManager) readOutput(reader io.Reader, source string) {
	scanner := bufio.NewScanner(reader)

	// 匹配Xray访问日志的正则表达式
	accessLogPattern := regexp.MustCompile(`(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+\[(\w+)\]\s+\[([^\]]+)\]\s+(.+)`)
	// 匹配accepted/rejected行
	trafficPattern := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+):(\d+)\s+(accepted|rejected)\s+(tcp|udp):([^\s]+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// 发送到日志通道
		select {
		case pm.logChan <- fmt.Sprintf("[%s] %s", source, line):
		default:
		}

		// 解析访问日志
		if matches := accessLogPattern.FindStringSubmatch(line); matches != nil {
			logLevel := matches[2]
			inboundTag := matches[3]
			message := matches[4]

			// 检查是否是流量日志
			if trafficMatches := trafficPattern.FindStringSubmatch(message); trafficMatches != nil {
				clientIP := trafficMatches[1]
				clientPort := trafficMatches[2]
				action := trafficMatches[3]
				protocol := trafficMatches[4]
				target := trafficMatches[5]

				// 构建流量日志消息
				trafficMsg := fmt.Sprintf("[%s] %s %s -> %s (%s)",
					inboundTag,
					action,
					fmt.Sprintf("%s:%s", clientIP, clientPort),
					target,
					protocol)

				// 广播流量日志到前端
				broadcaster.BroadcastLog(strings.ToUpper(logLevel), trafficMsg, "traffic")

				logger.Info("代理流量: %s", trafficMsg)
			}
		}
	}
}

// wait 等待进程结束
func (pm *ProcessManager) wait() {
	if pm.cmd == nil {
		return
	}

	err := pm.cmd.Wait()

	pm.mu.Lock()
	pm.running = false

	if err != nil {
		logger.Error("Xray进程异常退出: %v", err)
		pm.logChan <- fmt.Sprintf("[error] 进程异常退出: %v", err)
	} else {
		logger.Info("Xray进程正常退出")
		pm.logChan <- "[info] 进程正常退出"
	}

	pm.mu.Unlock()
}

// TestConnection 测试节点连接
// 参数：
//   - node: 要测试的节点
//   - timeout: 超时时间（秒）
//
// 返回：
//   - latency: 延迟（毫秒）
//   - error: 错误信息
func (pm *ProcessManager) TestConnection(node *model.Node, timeout int) (int, error) {
	pm.mu.Lock()
	wasRunning := pm.running
	var oldNode *model.Node
	var oldProxyMode string
	var oldRules []*model.Rule
	if wasRunning {
		oldNode = pm.currentNode
		oldProxyMode = pm.proxyMode
		oldRules = pm.rules
	}
	pm.mu.Unlock()

	testPort := 20808
	if err := pm.Start(node, testPort, "global", nil); err != nil {
		return 0, fmt.Errorf("启动测试失败: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	start := time.Now()

	testURL := "https://www.google.com/generate_204"
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "socks5",
				Host:   fmt.Sprintf("127.0.0.1:%d", testPort),
			}),
		},
		Timeout: time.Duration(timeout) * time.Second,
	}

	resp, err := client.Get(testURL)
	if err != nil {
		pm.Stop()
		return 0, fmt.Errorf("连接测试失败: %w", err)
	}
	defer resp.Body.Close()

	latency := int(time.Since(start).Milliseconds())

	pm.Stop()

	if wasRunning && oldNode != nil {
		pm.Start(oldNode, 10808, oldProxyMode, oldRules)
	}

	return latency, nil
}

// TrafficInfo 流量信息
type TrafficInfo struct {
	UploadSpeed   int64 // 上传速度 (bytes/s)
	DownloadSpeed int64 // 下载速度 (bytes/s)
	UploadTotal   int64 // 总上传 (bytes)
	DownloadTotal int64 // 总下载 (bytes)
}

// GetTraffic 获取流量统计
func (pm *ProcessManager) GetTraffic() *TrafficInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if !pm.running || pm.statsClient == nil {
		return &TrafficInfo{}
	}

	// 获取SOCKS入站流量
	socksStats, err := pm.statsClient.GetInboundTraffic("socks-in")
	if err != nil {
		logger.Debug("获取SOCKS流量失败: %v", err)
		return &TrafficInfo{}
	}

	// 获取HTTP入站流量
	httpStats, err := pm.statsClient.GetInboundTraffic("http-in")
	if err != nil {
		logger.Debug("获取HTTP流量失败: %v", err)
	}

	totalUpload := socksStats.Upload + httpStats.Upload
	totalDownload := socksStats.Download + httpStats.Download

	now := time.Now()
	duration := now.Sub(pm.lastTime).Seconds()
	if duration <= 0 {
		duration = 1
	}

	// 计算速度
	uploadSpeed := int64(float64(totalUpload-pm.lastUpload) / duration)
	downloadSpeed := int64(float64(totalDownload-pm.lastDownload) / duration)

	// 更新上次记录
	pm.lastUpload = totalUpload
	pm.lastDownload = totalDownload
	pm.lastTime = now

	// 防止负数
	if uploadSpeed < 0 {
		uploadSpeed = 0
	}
	if downloadSpeed < 0 {
		downloadSpeed = 0
	}

	return &TrafficInfo{
		UploadSpeed:   uploadSpeed,
		DownloadSpeed: downloadSpeed,
		UploadTotal:   totalUpload,
		DownloadTotal: totalDownload,
	}
}
