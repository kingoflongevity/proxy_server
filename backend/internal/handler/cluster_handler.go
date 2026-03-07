package handler

import (
	"context"
	"fmt"
	"proxy_server/internal/model"
	"proxy_server/internal/service"
	"proxy_server/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

// ClusterHandler 集群处理器
type ClusterHandler struct {
	sshManager *service.SSHManager
	scanner    *service.ScannerService
	deployer   *service.DeployerService
	backup     *service.BackupService
	servers    map[string]*model.ClusterServer
	groups     map[string]*model.ServerGroup
	tasks      map[string]*model.ScanTask
	deployTasks map[string]*model.DeployTask
}

// NewClusterHandler 创建集群处理器
func NewClusterHandler(serverRepo interface{}, groupRepo interface{}) *ClusterHandler {
	sshManager := service.NewSSHManager()
	return &ClusterHandler{
		sshManager:  sshManager,
		scanner:     service.NewScannerService(sshManager),
		deployer:    service.NewDeployerService(sshManager),
		backup:      service.NewBackupService(""),
		servers:     make(map[string]*model.ClusterServer),
		groups:      make(map[string]*model.ServerGroup),
		tasks:       make(map[string]*model.ScanTask),
		deployTasks: make(map[string]*model.DeployTask),
	}
}

// ListServers 列出所有服务器
func (h *ClusterHandler) ListServers(c *gin.Context) {
	servers := make([]model.ClusterServer, 0)
	for _, s := range h.servers {
		servers = append(servers, *s)
	}
	response.Success(c, servers)
}

// GetServer 获取单个服务器
func (h *ClusterHandler) GetServer(c *gin.Context) {
	id := c.Param("id")
	server, exists := h.servers[id]
	if !exists {
		response.Error(c, 404, "服务器不存在")
		return
	}
	response.Success(c, server)
}

// CreateServer 创建服务器
func (h *ClusterHandler) CreateServer(c *gin.Context) {
	var server model.ClusterServer
	if err := c.ShouldBindJSON(&server); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	server.ID = fmt.Sprintf("srv-%d", time.Now().UnixNano())
	server.CreatedAt = time.Now()
	server.UpdatedAt = time.Now()
	h.servers[server.ID] = &server

	response.Success(c, server)
}

// UpdateServer 更新服务器
func (h *ClusterHandler) UpdateServer(c *gin.Context) {
	id := c.Param("id")
	var server model.ClusterServer
	if err := c.ShouldBindJSON(&server); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	existing, exists := h.servers[id]
	if !exists {
		response.Error(c, 404, "服务器不存在")
		return
	}

	server.ID = id
	server.CreatedAt = existing.CreatedAt
	server.UpdatedAt = time.Now()
	h.servers[id] = &server

	response.Success(c, server)
}

// DeleteServer 删除服务器
func (h *ClusterHandler) DeleteServer(c *gin.Context) {
	id := c.Param("id")
	h.sshManager.Disconnect(id)
	delete(h.servers, id)
	response.Success(c, nil)
}

// TestConnection 测试连接
func (h *ClusterHandler) TestConnection(c *gin.Context) {
	id := c.Param("id")
	server, exists := h.servers[id]
	if !exists {
		response.Error(c, 404, "服务器不存在")
		return
	}

	var req struct {
		IP       string `json:"ip"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err == nil {
		server.IP = req.IP
		server.Port = req.Port
		server.Username = req.Username
		server.Password = req.Password
	}

	latency, err := h.scanner.TestSSH(server.IP, server.Port, server.Username, server.Password)
	if err != nil {
		response.Error(c, 400, "连接测试失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{"latency": latency, "status": "success"})
}

// StartProxy 启动代理
func (h *ClusterHandler) StartProxy(c *gin.Context) {
	id := c.Param("id")
	server, exists := h.servers[id]
	if !exists {
		response.Error(c, 404, "服务器不存在")
		return
	}

	config := service.SSHConfig{
		Host:     server.IP,
		Port:     server.Port,
		Username: server.Username,
		Password: server.Password,
		Timeout:  10 * time.Second,
	}

	if err := h.sshManager.Connect(id, config); err != nil {
		response.Error(c, 400, "连接失败: "+err.Error())
		return
	}

	server.Status = "active"
	server.UpdatedAt = time.Now()

	response.Success(c, gin.H{"status": "started"})
}

// StopProxy 停止代理
func (h *ClusterHandler) StopProxy(c *gin.Context) {
	id := c.Param("id")
	h.sshManager.Disconnect(id)

	if server, exists := h.servers[id]; exists {
		server.Status = "idle"
		server.UpdatedAt = time.Now()
	}

	response.Success(c, gin.H{"status": "stopped"})
}

// RestartProxy 重启代理
func (h *ClusterHandler) RestartProxy(c *gin.Context) {
	id := c.Param("id")
	h.sshManager.Disconnect(id)

	server, exists := h.servers[id]
	if !exists {
		response.Error(c, 404, "服务器不存在")
		return
	}

	config := service.SSHConfig{
		Host:     server.IP,
		Port:     server.Port,
		Username: server.Username,
		Password: server.Password,
		Timeout:  10 * time.Second,
	}

	if err := h.sshManager.Connect(id, config); err != nil {
		response.Error(c, 400, "重启失败: "+err.Error())
		return
	}

	server.Status = "active"
	server.UpdatedAt = time.Now()

	response.Success(c, gin.H{"status": "restarted"})
}

// GetProxyStatus 获取代理状态
func (h *ClusterHandler) GetProxyStatus(c *gin.Context) {
	id := c.Param("id")
	server, exists := h.servers[id]
	if !exists {
		response.Error(c, 404, "服务器不存在")
		return
	}

	connected := h.sshManager.CheckConnection(id)
	status := "idle"
	if connected {
		status = "active"
	}

	response.Success(c, gin.H{
		"status":      status,
		"connected":   connected,
		"cpu":         server.CPU,
		"memory":      server.Memory,
		"connections": server.Connections,
	})
}

// StartScan 开始扫描
func (h *ClusterHandler) StartScan(c *gin.Context) {
	var req struct {
		CIDR    string `json:"cidr"`
		Workers int    `json:"workers"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if req.Workers == 0 {
		req.Workers = 50
	}

	task := &model.ScanTask{
		ID:        fmt.Sprintf("scan-%d", time.Now().UnixNano()),
		CIDR:      req.CIDR,
		Status:    "running",
		Progress:  0,
		CreatedAt: time.Now(),
	}
	h.tasks[task.ID] = task

	go func() {
		progressChan := make(chan float64, 10)
		results, _ := h.scanner.ScanCIDR(context.Background(), req.CIDR, req.Workers, progressChan)
		task.Found = len(results)
		task.Status = "completed"
		task.Progress = 100
		task.UpdatedAt = time.Now()
	}()

	response.Success(c, task)
}

// GetScanTask 获取扫描任务
func (h *ClusterHandler) GetScanTask(c *gin.Context) {
	id := c.Param("id")
	task, exists := h.tasks[id]
	if !exists {
		response.Error(c, 404, "任务不存在")
		return
	}
	response.Success(c, task)
}

// CancelScan 取消扫描
func (h *ClusterHandler) CancelScan(c *gin.Context) {
	id := c.Param("id")
	if task, exists := h.tasks[id]; exists {
		task.Status = "cancelled"
	}
	response.Success(c, nil)
}

// QuickScan 快速扫描
func (h *ClusterHandler) QuickScan(c *gin.Context) {
	networks, err := service.GetLocalNetwork()
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.Success(c, networks)
}

// DeployProxy 部署代理
func (h *ClusterHandler) DeployProxy(c *gin.Context) {
	var req struct {
		ServerID  string `json:"serverId"`
		ProxyPort int    `json:"proxyPort"`
		ProxyType string `json:"proxyType"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	config := service.DeployConfig{
		ServerID:  req.ServerID,
		ProxyPort: req.ProxyPort,
		ProxyType: req.ProxyType,
	}

	progressChan := make(chan model.DeployStep, 20)
	task, err := h.deployer.Deploy(context.Background(), config, progressChan)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	h.deployTasks[task.ID] = task
	response.Success(c, task)
}

// GetDeployTask 获取部署任务
func (h *ClusterHandler) GetDeployTask(c *gin.Context) {
	id := c.Param("id")
	task, exists := h.deployer.GetTask(id)
	if !exists {
		if t, ok := h.deployTasks[id]; ok {
			response.Success(c, t)
			return
		}
		response.Error(c, 404, "任务不存在")
		return
	}
	response.Success(c, task)
}

// BatchDeploy 批量部署
func (h *ClusterHandler) BatchDeploy(c *gin.Context) {
	var req struct {
		ServerIDs []string `json:"serverIds"`
		ProxyPort int      `json:"proxyPort"`
		ProxyType string   `json:"proxyType"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	tasks := make([]*model.DeployTask, 0)
	for _, serverID := range req.ServerIDs {
		config := service.DeployConfig{
			ServerID:  serverID,
			ProxyPort: req.ProxyPort,
			ProxyType: req.ProxyType,
		}

		task, _ := h.deployer.Deploy(context.Background(), config, nil)
		tasks = append(tasks, task)
	}

	response.Success(c, tasks)
}

// CreateBackup 创建备份
func (h *ClusterHandler) CreateBackup(c *gin.Context) {
	var req struct {
		Type string `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	servers := make([]model.ClusterServer, 0)
	for _, s := range h.servers {
		servers = append(servers, *s)
	}

	groups := make([]model.ServerGroup, 0)
	for _, g := range h.groups {
		groups = append(groups, *g)
	}

	record, err := h.backup.CreateBackup(req.Type, servers, groups)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, record)
}

// ListBackups 列出备份
func (h *ClusterHandler) ListBackups(c *gin.Context) {
	records, err := h.backup.ListBackups()
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.Success(c, records)
}

// RestoreBackup 恢复备份
func (h *ClusterHandler) RestoreBackup(c *gin.Context) {
	var req struct {
		BackupID string `json:"backupId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	servers, groups, err := h.backup.RestoreBackup(req.BackupID)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	for _, server := range servers {
		h.servers[server.ID] = &server
	}

	for _, group := range groups {
		h.groups[group.ID] = &group
	}

	response.Success(c, gin.H{
		"servers": len(servers),
		"groups":  len(groups),
	})
}

// DeleteBackup 删除备份
func (h *ClusterHandler) DeleteBackup(c *gin.Context) {
	id := c.Param("id")
	if err := h.backup.DeleteBackup(id); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

// UpdateScalePolicy 更新伸缩策略
func (h *ClusterHandler) UpdateScalePolicy(c *gin.Context) {
	var req struct {
		MinServers  int     `json:"minServers"`
		MaxServers  int     `json:"maxServers"`
		CPUThreshold float64 `json:"cpuThreshold"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	response.Success(c, gin.H{"status": "updated"})
}

// ScaleUp 扩容
func (h *ClusterHandler) ScaleUp(c *gin.Context) {
	response.Success(c, gin.H{"status": "scaling_up"})
}

// ScaleDown 缩容
func (h *ClusterHandler) ScaleDown(c *gin.Context) {
	response.Success(c, gin.H{"status": "scaling_down"})
}

// GetScaleEvents 获取伸缩事件
func (h *ClusterHandler) GetScaleEvents(c *gin.Context) {
	events := []model.ScaleEvent{}
	response.Success(c, events)
}

// ListGroups 列出分组
func (h *ClusterHandler) ListGroups(c *gin.Context) {
	groups := make([]model.ServerGroup, 0)
	for _, g := range h.groups {
		groups = append(groups, *g)
	}
	response.Success(c, groups)
}

// CreateGroup 创建分组
func (h *ClusterHandler) CreateGroup(c *gin.Context) {
	var group model.ServerGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	group.ID = fmt.Sprintf("grp-%d", time.Now().UnixNano())
	group.CreatedAt = time.Now()
	h.groups[group.ID] = &group

	response.Success(c, group)
}

// GetGroupMetrics 获取分组指标
func (h *ClusterHandler) GetGroupMetrics(c *gin.Context) {
	id := c.Param("id")
	_ = id
	response.Success(c, gin.H{
		"cpu":         0,
		"memory":      0,
		"connections": 0,
	})
}

// GetTopology 获取拓扑图
func (h *ClusterHandler) GetTopology(c *gin.Context) {
	servers := make([]model.ClusterServer, 0)
	for _, s := range h.servers {
		servers = append(servers, *s)
	}

	groups := make([]model.ServerGroup, 0)
	for _, g := range h.groups {
		groups = append(groups, *g)
	}

	connections := make([]model.Connection, 0)
	for _, server := range servers {
		if server.GroupID != "" {
			connections = append(connections, model.Connection{
				From: server.GroupID,
				To:   server.ID,
				Type: "group",
			})
		}
	}

	response.Success(c, gin.H{
		"servers":     servers,
		"groups":      groups,
		"connections": connections,
	})
}

// GetNetworkSegment 获取当前网段
func (h *ClusterHandler) GetNetworkSegment(c *gin.Context) {
	networks, err := service.GetLocalNetwork()
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	// 返回第一个网段作为默认扫描范围
	var cidr string
	if len(networks) > 0 {
		cidr = networks[0]
	} else {
		cidr = "192.168.1.0/24"
	}

	response.Success(c, gin.H{
		"cidr":     cidr,
		"networks": networks,
	})
}
