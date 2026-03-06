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
	"runtime"
	"sync"
	"time"

	"proxy_server/internal/model"
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
// 返回：
//   - error: 错误信息
func (pm *ProcessManager) Start(node *model.Node, localPort int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.running {
		return fmt.Errorf("Xray进程已在运行")
	}

	generator := NewConfigGenerator(localPort)
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

	go pm.readOutput(stdout, "stdout")
	go pm.readOutput(stderr, "stderr")

	go pm.wait()

	logger.Info("Xray进程已启动，节点: %s, 本地端口: %d", node.Name, localPort)
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
func (pm *ProcessManager) Restart(node *model.Node, localPort int) error {
	if err := pm.Stop(); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	return pm.Start(node, localPort)
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

// GetLogs 获取日志通道
func (pm *ProcessManager) GetLogs() <-chan string {
	return pm.logChan
}

// readOutput 读取进程输出
func (pm *ProcessManager) readOutput(reader io.Reader, source string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		select {
		case pm.logChan <- fmt.Sprintf("[%s] %s", source, line):
		default:
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
// 返回：
//   - latency: 延迟（毫秒）
//   - error: 错误信息
func (pm *ProcessManager) TestConnection(node *model.Node, timeout int) (int, error) {
	pm.mu.Lock()
	wasRunning := pm.running
	var oldNode *model.Node
	if wasRunning {
		oldNode = pm.currentNode
	}
	pm.mu.Unlock()

	testPort := 20808
	if err := pm.Start(node, testPort); err != nil {
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
		pm.Start(oldNode, 10808)
	}

	return latency, nil
}
