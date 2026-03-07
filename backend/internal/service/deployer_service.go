package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"proxy_server/internal/model"
	"proxy_server/pkg/logger"
)

// DeployerService 部署服务
type DeployerService struct {
	sshManager *SSHManager
	tasks      map[string]*model.DeployTask
	mu         sync.RWMutex
}

// NewDeployerService 创建部署服务
func NewDeployerService(sshManager *SSHManager) *DeployerService {
	return &DeployerService{
		sshManager: sshManager,
		tasks:      make(map[string]*model.DeployTask),
	}
}

// DeployConfig 部署配置
type DeployConfig struct {
	ServerID   string `json:"serverId"`
	ProxyType  string `json:"proxyType"`  // vmess, vless, trojan, ss
	ProxyPort  int    `json:"proxyPort"`  // 代理端口
	APIToken   string `json:"apiToken"`   // API令牌
	AutoStart  bool   `json:"autoStart"`  // 自动启动
}

// Deploy 部署代理服务
func (d *DeployerService) Deploy(ctx context.Context, config DeployConfig, progressChan chan<- model.DeployStep) (*model.DeployTask, error) {
	task := &model.DeployTask{
		ID:        config.ServerID,
		ServerID:  config.ServerID,
		Status:    "pending",
		Progress:  0,
		CreatedAt: time.Now(),
		Steps: []model.DeployStep{
			{Name: "检查系统环境", Status: "pending"},
			{Name: "安装依赖", Status: "pending"},
			{Name: "下载Xray核心", Status: "pending"},
			{Name: "配置代理", Status: "pending"},
			{Name: "创建systemd服务", Status: "pending"},
			{Name: "启动服务", Status: "pending"},
			{Name: "验证部署", Status: "pending"},
			{Name: "完成", Status: "pending"},
		},
	}

	d.mu.Lock()
	d.tasks[task.ID] = task
	d.mu.Unlock()

	go d.runDeploy(ctx, task, config, progressChan)

	return task, nil
}

// runDeploy 执行部署
func (d *DeployerService) runDeploy(ctx context.Context, task *model.DeployTask, config DeployConfig, progressChan chan<- model.DeployStep) {
	task.Status = "running"

	steps := []func(context.Context, DeployConfig) (string, error){
		d.stepCheckEnvironment,
		d.stepInstallDependencies,
		d.stepDownloadCore,
		d.stepConfigureProxy,
		d.stepCreateService,
		d.stepStartService,
		d.stepVerifyDeployment,
	}

	for i, step := range steps {
		select {
		case <-ctx.Done():
			task.Status = "failed"
			task.Steps[i].Status = "failed"
			task.Steps[i].Message = "部署已取消"
			return
		default:
		}

		task.CurrentStep = i
		task.Steps[i].Status = "running"
		start := time.Now()

		msg, err := step(ctx, config)
		duration := int(time.Since(start).Seconds())

		if err != nil {
			task.Status = "failed"
			task.Steps[i].Status = "failed"
			task.Steps[i].Message = err.Error()
			task.Steps[i].Duration = duration

			if progressChan != nil {
				progressChan <- task.Steps[i]
			}
			return
		}

		task.Steps[i].Status = "success"
		task.Steps[i].Message = msg
		task.Steps[i].Duration = duration
		task.Progress = float64(i+1) / float64(len(steps)+1) * 100

		if progressChan != nil {
			progressChan <- task.Steps[i]
		}
	}

	task.Steps[len(task.Steps)-1].Status = "success"
	task.Steps[len(task.Steps)-1].Message = "部署完成"
	task.Status = "success"
	task.Progress = 100
	task.UpdatedAt = time.Now()

	if progressChan != nil {
		progressChan <- task.Steps[len(task.Steps)-1]
		close(progressChan)
	}

	logger.Info("服务器 %s 部署完成", config.ServerID)
}

// stepCheckEnvironment 检查系统环境
func (d *DeployerService) stepCheckEnvironment(ctx context.Context, config DeployConfig) (string, error) {
	output, err := d.sshManager.ExecuteCommand(config.ServerID, "uname -m")
	if err != nil {
		return "", fmt.Errorf("获取系统架构失败: %w", err)
	}

	arch := strings.TrimSpace(output)
	if arch != "x86_64" && arch != "aarch64" && arch != "arm64" {
		return "", fmt.Errorf("不支持的系统架构: %s", arch)
	}

	output, err = d.sshManager.ExecuteCommand(config.ServerID, "cat /etc/os-release | grep '^ID=' | cut -d'=' -f2 | tr -d '\"'")
	if err != nil {
		return "", fmt.Errorf("获取系统类型失败: %w", err)
	}

	osType := strings.TrimSpace(strings.ToLower(output))
	return fmt.Sprintf("系统: %s, 架构: %s", osType, arch), nil
}

// stepInstallDependencies 安装依赖
func (d *DeployerService) stepInstallDependencies(ctx context.Context, config DeployConfig) (string, error) {
	commands := map[string]string{
		"ubuntu": "apt-get update && apt-get install -y curl wget unzip",
		"debian": "apt-get update && apt-get install -y curl wget unzip",
		"centos": "yum install -y curl wget unzip",
		"alpine": "apk add --no-cache curl wget unzip",
	}

	output, err := d.sshManager.ExecuteCommand(config.ServerID, "cat /etc/os-release | grep '^ID=' | cut -d'=' -f2 | tr -d '\"'")
	if err != nil {
		return "", err
	}

	osType := strings.TrimSpace(strings.ToLower(output))
	cmd, ok := commands[osType]
	if !ok {
		cmd = commands["ubuntu"]
	}

	_, err = d.sshManager.ExecuteCommand(config.ServerID, cmd)
	if err != nil {
		return "", fmt.Errorf("安装依赖失败: %w", err)
	}

	return "依赖安装完成", nil
}

// stepDownloadCore 下载Xray核心
func (d *DeployerService) stepDownloadCore(ctx context.Context, config DeployConfig) (string, error) {
	arch, _ := d.sshManager.ExecuteCommand(config.ServerID, "uname -m")
	arch = strings.TrimSpace(arch)

	var downloadArch string
	switch arch {
	case "x86_64":
		downloadArch = "64"
	case "aarch64", "arm64":
		downloadArch = "arm64-v8a"
	default:
		downloadArch = "64"
	}

	downloadURL := fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/latest/download/Xray-linux-%s.zip", downloadArch)

	commands := []string{
		"mkdir -p /opt/proxy-server",
		fmt.Sprintf("cd /opt/proxy-server && curl -L -o xray.zip %s", downloadURL),
		"cd /opt/proxy-server && unzip -o xray.zip && rm xray.zip",
		"chmod +x /opt/proxy-server/xray",
	}

	for _, cmd := range commands {
		if _, err := d.sshManager.ExecuteCommand(config.ServerID, cmd); err != nil {
			return "", fmt.Errorf("下载核心失败: %w", err)
		}
	}

	return "Xray核心下载完成", nil
}

// stepConfigureProxy 配置代理
func (d *DeployerService) stepConfigureProxy(ctx context.Context, config DeployConfig) (string, error) {
	proxyConfig := d.generateProxyConfig(config)
	configJSON, _ := json.MarshalIndent(proxyConfig, "", "  ")

	if err := d.sshManager.UploadData(config.ServerID, configJSON, "/opt/proxy-server/config.json"); err != nil {
		return "", fmt.Errorf("上传配置失败: %w", err)
	}

	return "代理配置完成", nil
}

// generateProxyConfig 生成代理配置
func (d *DeployerService) generateProxyConfig(config DeployConfig) map[string]interface{} {
	return map[string]interface{}{
		"log": map[string]interface{}{
			"loglevel": "warning",
		},
		"inbounds": []map[string]interface{}{
			{
				"port":     config.ProxyPort,
				"protocol": "socks",
				"settings": map[string]interface{}{
					"udp": true,
				},
			},
			{
				"port":     config.ProxyPort + 1,
				"protocol": "http",
			},
		},
		"outbounds": []map[string]interface{}{
			{
				"protocol": "freedom",
				"tag":      "direct",
			},
		},
	}
}

// stepCreateService 创建systemd服务
func (d *DeployerService) stepCreateService(ctx context.Context, config DeployConfig) (string, error) {
	serviceContent := `[Unit]
Description=Proxy Server
After=network.target

[Service]
Type=simple
ExecStart=/opt/proxy-server/xray run -c /opt/proxy-server/config.json
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`

	if err := d.sshManager.UploadData(config.ServerID, []byte(serviceContent), "/etc/systemd/system/proxy-server.service"); err != nil {
		return "", fmt.Errorf("创建服务文件失败: %w", err)
	}

	commands := []string{
		"systemctl daemon-reload",
		"systemctl enable proxy-server",
	}

	for _, cmd := range commands {
		if _, err := d.sshManager.ExecuteCommand(config.ServerID, cmd); err != nil {
			return "", fmt.Errorf("配置服务失败: %w", err)
		}
	}

	return "服务创建完成", nil
}

// stepStartService 启动服务
func (d *DeployerService) stepStartService(ctx context.Context, config DeployConfig) (string, error) {
	_, err := d.sshManager.ExecuteCommand(config.ServerID, "systemctl restart proxy-server")
	if err != nil {
		return "", fmt.Errorf("启动服务失败: %w", err)
	}

	time.Sleep(2 * time.Second)

	output, err := d.sshManager.ExecuteCommand(config.ServerID, "systemctl is-active proxy-server")
	if err != nil || strings.TrimSpace(output) != "active" {
		return "", fmt.Errorf("服务启动失败")
	}

	return "服务启动成功", nil
}

// stepVerifyDeployment 验证部署
func (d *DeployerService) stepVerifyDeployment(ctx context.Context, config DeployConfig) (string, error) {
	output, err := d.sshManager.ExecuteCommand(config.ServerID, fmt.Sprintf("ss -tlnp | grep %d", config.ProxyPort))
	if err != nil || output == "" {
		return "", fmt.Errorf("端口 %d 未监听", config.ProxyPort)
	}

	return "部署验证通过", nil
}

// GetTask 获取部署任务
func (d *DeployerService) GetTask(taskID string) (*model.DeployTask, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	task, exists := d.tasks[taskID]
	return task, exists
}

// CancelTask 取消部署任务
func (d *DeployerService) CancelTask(taskID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if task, exists := d.tasks[taskID]; exists {
		if task.Status == "running" {
			task.Status = "cancelled"
			return nil
		}
	}

	return fmt.Errorf("任务不存在或已完成")
}

// Undeploy 卸载代理服务
func (d *DeployerService) Undeploy(serverID string) error {
	commands := []string{
		"systemctl stop proxy-server || true",
		"systemctl disable proxy-server || true",
		"rm -f /etc/systemd/system/proxy-server.service",
		"systemctl daemon-reload",
		"rm -rf /opt/proxy-server",
	}

	for _, cmd := range commands {
		if _, err := d.sshManager.ExecuteCommand(serverID, cmd); err != nil {
			logger.Warn("卸载命令执行失败: %v", err)
		}
	}

	logger.Info("服务器 %s 代理服务已卸载", serverID)
	return nil
}

// RestartService 重启远程服务
func (d *DeployerService) RestartService(serverID string) error {
	_, err := d.sshManager.ExecuteCommand(serverID, "systemctl restart proxy-server")
	return err
}

// GetServiceStatus 获取服务状态
func (d *DeployerService) GetServiceStatus(serverID string) (string, error) {
	output, err := d.sshManager.ExecuteCommand(serverID, "systemctl is-active proxy-server")
	if err != nil {
		return "unknown", err
	}
	return strings.TrimSpace(output), nil
}

// StartProxy 启动代理服务
func (d *DeployerService) StartProxy(serverID string) error {
	_, err := d.sshManager.ExecuteCommand(serverID, "systemctl start proxy-server")
	if err != nil {
		logger.Error("启动代理服务失败: %s - %v", serverID, err)
		return err
	}
	logger.Info("代理服务已启动: %s", serverID)
	return nil
}

// StopProxy 停止代理服务
func (d *DeployerService) StopProxy(serverID string) error {
	_, err := d.sshManager.ExecuteCommand(serverID, "systemctl stop proxy-server")
	if err != nil {
		logger.Error("停止代理服务失败: %s - %v", serverID, err)
		return err
	}
	logger.Info("代理服务已停止: %s", serverID)
	return nil
}
