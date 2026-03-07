package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"proxy_server/internal/model"
	"proxy_server/pkg/logger"

	"golang.org/x/crypto/ssh"
)

// SSHManager SSH连接管理器
type SSHManager struct {
	mu       sync.RWMutex
	clients  map[string]*ssh.Client
	sessions map[string]*ssh.Session
}

// NewSSHManager 创建SSH管理器
func NewSSHManager() *SSHManager {
	return &SSHManager{
		clients:  make(map[string]*ssh.Client),
		sessions: make(map[string]*ssh.Session),
	}
}

// SSHConfig SSH配置
type SSHConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	KeyPath  string
	Timeout  time.Duration
}

// Connect 连接服务器
func (m *SSHManager) Connect(serverID string, config SSHConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, exists := m.clients[serverID]; exists {
		if _, _, err := client.SendRequest("keepalive", true, nil); err == nil {
			return nil
		}
		delete(m.clients, serverID)
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	var authMethods []ssh.AuthMethod

	if config.KeyPath != "" {
		key, err := os.ReadFile(config.KeyPath)
		if err != nil {
			return fmt.Errorf("读取密钥文件失败: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("解析密钥失败: %w", err)
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("未提供认证方式")
	}

	clientConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         config.Timeout,
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return fmt.Errorf("SSH连接失败: %w", err)
	}

	m.clients[serverID] = client
	logger.Info("SSH连接成功: %s (%s)", serverID, addr)

	return nil
}

// Disconnect 断开连接
func (m *SSHManager) Disconnect(serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[serverID]; exists {
		session.Close()
		delete(m.sessions, serverID)
	}

	if client, exists := m.clients[serverID]; exists {
		err := client.Close()
		delete(m.clients, serverID)
		logger.Info("SSH连接已断开: %s", serverID)
		return err
	}

	return nil
}

// ExecuteCommand 执行命令
func (m *SSHManager) ExecuteCommand(serverID, command string) (string, error) {
	m.mu.RLock()
	client, exists := m.clients[serverID]
	m.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("服务器未连接: %s", serverID)
	}

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)
	if err != nil {
		return stderr.String(), fmt.Errorf("命令执行失败: %w", err)
	}

	return stdout.String(), nil
}

// ExecuteCommandWithOutput 实时输出执行命令
func (m *SSHManager) ExecuteCommandWithOutput(serverID, command string, outputChan chan<- string) error {
	m.mu.RLock()
	client, exists := m.clients[serverID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("服务器未连接: %s", serverID)
	}

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("获取stdout失败: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("获取stderr失败: %w", err)
	}

	if err := session.Start(command); err != nil {
		return fmt.Errorf("启动命令失败: %w", err)
	}

	go func() {
		scanner := io.MultiReader(stdout, stderr)
		buf := make([]byte, 1024)
		for {
			n, err := scanner.Read(buf)
			if n > 0 {
				outputChan <- string(buf[:n])
			}
			if err != nil {
				break
			}
		}
		close(outputChan)
	}()

	return session.Wait()
}

// UploadFile 上传文件
func (m *SSHManager) UploadFile(serverID, localPath, remotePath string) error {
	m.mu.RLock()
	client, exists := m.clients[serverID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("服务器未连接: %s", serverID)
	}

	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %w", err)
	}
	defer localFile.Close()

	stat, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()

		fmt.Fprintf(w, "C%04o %d %s\n", 0644, stat.Size(), filepath.Base(remotePath))
		io.Copy(w, localFile)
		fmt.Fprint(w, "\x00")
	}()

	if err := session.Run(fmt.Sprintf("scp -t %s", filepath.Dir(remotePath))); err != nil {
		return fmt.Errorf("SCP上传失败: %w", err)
	}

	return nil
}

// UploadData 上传数据
func (m *SSHManager) UploadData(serverID string, data []byte, remotePath string) error {
	m.mu.RLock()
	client, exists := m.clients[serverID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("服务器未连接: %s", serverID)
	}

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()

		fmt.Fprintf(w, "C%04o %d %s\n", 0644, len(data), filepath.Base(remotePath))
		w.Write(data)
		fmt.Fprint(w, "\x00")
	}()

	if err := session.Run(fmt.Sprintf("scp -t %s", filepath.Dir(remotePath))); err != nil {
		return fmt.Errorf("SCP上传失败: %w", err)
	}

	return nil
}

// DownloadFile 下载文件
func (m *SSHManager) DownloadFile(serverID, remotePath, localPath string) error {
	m.mu.RLock()
	client, exists := m.clients[serverID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("服务器未连接: %s", serverID)
	}

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	var stdout bytes.Buffer
	session.Stdout = &stdout

	if err := session.Run(fmt.Sprintf("cat %s", remotePath)); err != nil {
		return fmt.Errorf("读取远程文件失败: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}

	if err := os.WriteFile(localPath, stdout.Bytes(), 0644); err != nil {
		return fmt.Errorf("写入本地文件失败: %w", err)
	}

	return nil
}

// GetSystemInfo 获取系统信息
func (m *SSHManager) GetSystemInfo(serverID string) (*model.ClusterServer, error) {
	info := &model.ClusterServer{}

	osType, err := m.ExecuteCommand(serverID, "cat /etc/os-release | grep '^ID=' | cut -d'=' -f2 | tr -d '\"'")
	if err != nil {
		return nil, err
	}
	info.OSType = strings.TrimSpace(strings.ToLower(osType))

	arch, err := m.ExecuteCommand(serverID, "uname -m")
	if err != nil {
		return nil, err
	}
	info.Arch = strings.TrimSpace(arch)

	return info, nil
}

// CheckConnection 检查连接状态
func (m *SSHManager) CheckConnection(serverID string) bool {
	m.mu.RLock()
	client, exists := m.clients[serverID]
	m.mu.RUnlock()

	if !exists {
		return false
	}

	_, _, err := client.SendRequest("keepalive", true, nil)
	return err == nil
}

// GetConnectedServers 获取已连接的服务器列表
func (m *SSHManager) GetConnectedServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	servers := make([]string, 0, len(m.clients))
	for id := range m.clients {
		servers = append(servers, id)
	}
	return servers
}

// TestConnection 测试SSH连接
func (m *SSHManager) TestConnection(config SSHConfig) (int, error) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	var authMethods []ssh.AuthMethod

	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	clientConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         config.Timeout,
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	dialer := &net.Dialer{Timeout: config.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return 0, fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, clientConfig)
	if err != nil {
		return 0, fmt.Errorf("SSH握手失败: %w", err)
	}
	defer sshConn.Close()

	client := ssh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	latency := int(time.Since(start).Milliseconds())
	return latency, nil
}
