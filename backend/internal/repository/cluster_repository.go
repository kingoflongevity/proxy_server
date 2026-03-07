package repository

import (
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	"proxy_server/internal/model"
)

// ClusterServerRepository 集群服务器仓库接口
type ClusterServerRepository interface {
	// 服务器管理
	Create(server *model.ClusterServer) error
	GetByID(id string) (*model.ClusterServer, error)
	GetByIP(ip string) (*model.ClusterServer, error)
	GetAll() ([]*model.ClusterServer, error)
	GetByGroupID(groupID string) ([]*model.ClusterServer, error)
	Update(server *model.ClusterServer) error
	Delete(id string) error
	
	// 分组管理
	CreateGroup(group *model.ServerGroup) error
	GetGroupByID(id string) (*model.ServerGroup, error)
	GetAllGroups() ([]*model.ServerGroup, error)
	GetAutoScaleGroups() ([]*model.ServerGroup, error)
	UpdateGroup(group *model.ServerGroup) error
	DeleteGroup(id string) error
	
	// 查询
	Query(query *model.ServerListQuery) ([]*model.ClusterServer, int64, error)
}

// clusterServerRepository 集群服务器仓库实现
type clusterServerRepository struct {
	db    *sql.DB
	mu    sync.RWMutex
	cache map[string]*model.ClusterServer
}

// NewClusterServerRepository 创建集群服务器仓库
func NewClusterServerRepository(db *sql.DB) ClusterServerRepository {
	repo := &clusterServerRepository{
		db:    db,
		cache: make(map[string]*model.ClusterServer),
	}
	
	// 初始化数据库表
	repo.initTables()
	
	return repo
}

// initTables 初始化数据库表
func (r *clusterServerRepository) initTables() {
	// 创建服务器表
	r.db.Exec(`
		CREATE TABLE IF NOT EXISTS cluster_servers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			ip TEXT NOT NULL UNIQUE,
			port INTEGER DEFAULT 22,
			username TEXT NOT NULL,
			password TEXT,
			private_key TEXT,
			os_type TEXT,
			os_version TEXT,
			arch TEXT,
			status TEXT DEFAULT 'offline',
			last_heartbeat DATETIME,
			proxy_enabled INTEGER DEFAULT 0,
			proxy_port INTEGER,
			proxy_type TEXT,
			proxy_config TEXT,
			cpu INTEGER,
			memory INTEGER,
			disk INTEGER,
			cpu_usage REAL,
			memory_usage REAL,
			bandwidth_up INTEGER,
			bandwidth_down INTEGER,
			connections INTEGER,
			tags TEXT,
			group_id TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deployed_at DATETIME
		)
	`)

	// 创建分组表
	r.db.Exec(`
		CREATE TABLE IF NOT EXISTS server_groups (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			auto_scale INTEGER DEFAULT 0,
			min_servers INTEGER DEFAULT 1,
			max_servers INTEGER DEFAULT 10,
			scale_policy TEXT DEFAULT 'cpu',
			scale_threshold REAL DEFAULT 80.0,
			created_at DATETIME,
			updated_at DATETIME
		)
	`)
}

// Create 创建服务器
func (r *clusterServerRepository) Create(server *model.ClusterServer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 序列化标签
	tagsJSON, _ := json.Marshal(server.Tags)

	_, err := r.db.Exec(`
		INSERT INTO cluster_servers (
			id, name, ip, port, username, password, private_key,
			os_type, os_version, arch, status, last_heartbeat,
			proxy_enabled, proxy_port, proxy_type, proxy_config,
			cpu, memory, disk, cpu_usage, memory_usage,
			bandwidth_up, bandwidth_down, connections,
			tags, group_id, created_at, updated_at, deployed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		server.ID, server.Name, server.IP, server.Port, server.Username, server.Password, server.PrivateKey,
		server.OSType, server.OSVersion, server.Arch, server.Status, server.LastHeartbeat,
		server.ProxyEnabled, server.ProxyPort, server.ProxyType, server.ProxyConfig,
		server.CPU, server.Memory, server.Disk, server.CPUUsage, server.MemoryUsage,
		server.BandwidthUp, server.BandwidthDown, server.Connections,
		string(tagsJSON), server.GroupID, server.CreatedAt, server.UpdatedAt, server.DeployedAt,
	)

	if err != nil {
		return err
	}

	r.cache[server.ID] = server
	return nil
}

// GetByID 根据ID获取服务器
func (r *clusterServerRepository) GetByID(id string) (*model.ClusterServer, error) {
	r.mu.RLock()
	if server, ok := r.cache[id]; ok {
		r.mu.RUnlock()
		return server, nil
	}
	r.mu.RUnlock()

	server := &model.ClusterServer{}
	var tagsJSON string
	var lastHeartbeat, deployedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, name, ip, port, username, password, private_key,
			os_type, os_version, arch, status, last_heartbeat,
			proxy_enabled, proxy_port, proxy_type, proxy_config,
			cpu, memory, disk, cpu_usage, memory_usage,
			bandwidth_up, bandwidth_down, connections,
			tags, group_id, created_at, updated_at, deployed_at
		FROM cluster_servers WHERE id = ?
	`, id).Scan(
		&server.ID, &server.Name, &server.IP, &server.Port, &server.Username, &server.Password, &server.PrivateKey,
		&server.OSType, &server.OSVersion, &server.Arch, &server.Status, &lastHeartbeat,
		&server.ProxyEnabled, &server.ProxyPort, &server.ProxyType, &server.ProxyConfig,
		&server.CPU, &server.Memory, &server.Disk, &server.CPUUsage, &server.MemoryUsage,
		&server.BandwidthUp, &server.BandwidthDown, &server.Connections,
		&tagsJSON, &server.GroupID, &server.CreatedAt, &server.UpdatedAt, &deployedAt,
	)

	if err != nil {
		return nil, err
	}

	// 解析标签
	if tagsJSON != "" {
		json.Unmarshal([]byte(tagsJSON), &server.Tags)
	}

	// 处理可空时间
	if lastHeartbeat.Valid {
		server.LastHeartbeat = lastHeartbeat.Time
	}
	if deployedAt.Valid {
		server.DeployedAt = deployedAt.Time
	}

	return server, nil
}

// GetByIP 根据IP获取服务器
func (r *clusterServerRepository) GetByIP(ip string) (*model.ClusterServer, error) {
	var id string
	err := r.db.QueryRow(`SELECT id FROM cluster_servers WHERE ip = ?`, ip).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(id)
}

// GetAll 获取所有服务器
func (r *clusterServerRepository) GetAll() ([]*model.ClusterServer, error) {
	rows, err := r.db.Query(`SELECT id FROM cluster_servers ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*model.ClusterServer
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		server, err := r.GetByID(id)
		if err != nil {
			continue
		}
		servers = append(servers, server)
	}

	return servers, nil
}

// GetByGroupID 根据分组ID获取服务器
func (r *clusterServerRepository) GetByGroupID(groupID string) ([]*model.ClusterServer, error) {
	rows, err := r.db.Query(`SELECT id FROM cluster_servers WHERE group_id = ?`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*model.ClusterServer
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		server, err := r.GetByID(id)
		if err != nil {
			continue
		}
		servers = append(servers, server)
	}

	return servers, nil
}

// Update 更新服务器
func (r *clusterServerRepository) Update(server *model.ClusterServer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tagsJSON, _ := json.Marshal(server.Tags)

	_, err := r.db.Exec(`
		UPDATE cluster_servers SET
			name = ?, ip = ?, port = ?, username = ?, password = ?, private_key = ?,
			os_type = ?, os_version = ?, arch = ?, status = ?, last_heartbeat = ?,
			proxy_enabled = ?, proxy_port = ?, proxy_type = ?, proxy_config = ?,
			cpu = ?, memory = ?, disk = ?, cpu_usage = ?, memory_usage = ?,
			bandwidth_up = ?, bandwidth_down = ?, connections = ?,
			tags = ?, group_id = ?, updated_at = ?, deployed_at = ?
		WHERE id = ?
	`,
		server.Name, server.IP, server.Port, server.Username, server.Password, server.PrivateKey,
		server.OSType, server.OSVersion, server.Arch, server.Status, server.LastHeartbeat,
		server.ProxyEnabled, server.ProxyPort, server.ProxyType, server.ProxyConfig,
		server.CPU, server.Memory, server.Disk, server.CPUUsage, server.MemoryUsage,
		server.BandwidthUp, server.BandwidthDown, server.Connections,
		string(tagsJSON), server.GroupID, server.UpdatedAt, server.DeployedAt,
		server.ID,
	)

	if err != nil {
		return err
	}

	r.cache[server.ID] = server
	return nil
}

// Delete 删除服务器
func (r *clusterServerRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec(`DELETE FROM cluster_servers WHERE id = ?`, id)
	if err != nil {
		return err
	}

	delete(r.cache, id)
	return nil
}

// Query 查询服务器
func (r *clusterServerRepository) Query(query *model.ServerListQuery) ([]*model.ClusterServer, int64, error) {
	sql := `SELECT id FROM cluster_servers WHERE 1=1`
	args := []interface{}{}

	if query.Status != "" {
		sql += ` AND status = ?`
		args = append(args, query.Status)
	}
	if query.GroupID != "" {
		sql += ` AND group_id = ?`
		args = append(args, query.GroupID)
	}
	if query.Keyword != "" {
		sql += ` AND (name LIKE ? OR ip LIKE ?)`
		keyword := "%" + query.Keyword + "%"
		args = append(args, keyword, keyword)
	}

	// 获取总数
	var total int64
	countSQL := `SELECT COUNT(*) FROM (` + sql + `)`
	r.db.QueryRow(countSQL, args...).Scan(&total)

	// 分页
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	offset := (query.Page - 1) * query.PageSize
	sql += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, query.PageSize, offset)

	rows, err := r.db.Query(sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var servers []*model.ClusterServer
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		server, err := r.GetByID(id)
		if err != nil {
			continue
		}
		servers = append(servers, server)
	}

	return servers, total, nil
}

// ====== 分组管理 ======

// CreateGroup 创建分组
func (r *clusterServerRepository) CreateGroup(group *model.ServerGroup) error {
	_, err := r.db.Exec(`
		INSERT INTO server_groups (id, name, description, auto_scale, min_servers, max_servers, scale_policy, scale_threshold, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, group.ID, group.Name, group.Description, group.AutoScale, group.MinServers, group.MaxServers, group.ScalePolicy, group.ScaleThreshold, group.CreatedAt, group.UpdatedAt)
	return err
}

// GetGroupByID 根据ID获取分组
func (r *clusterServerRepository) GetGroupByID(id string) (*model.ServerGroup, error) {
	group := &model.ServerGroup{}
	err := r.db.QueryRow(`
		SELECT id, name, description, auto_scale, min_servers, max_servers, scale_policy, scale_threshold, created_at, updated_at
		FROM server_groups WHERE id = ?
	`, id).Scan(
		&group.ID, &group.Name, &group.Description, &group.AutoScale, &group.MinServers, &group.MaxServers, &group.ScalePolicy, &group.ScaleThreshold, &group.CreatedAt, &group.UpdatedAt,
	)
	return group, err
}

// GetAllGroups 获取所有分组
func (r *clusterServerRepository) GetAllGroups() ([]*model.ServerGroup, error) {
	rows, err := r.db.Query(`SELECT id FROM server_groups ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*model.ServerGroup
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		group, err := r.GetGroupByID(id)
		if err != nil {
			continue
		}
		groups = append(groups, group)
	}

	return groups, nil
}

// GetAutoScaleGroups 获取启用自动伸缩的分组
func (r *clusterServerRepository) GetAutoScaleGroups() ([]*model.ServerGroup, error) {
	rows, err := r.db.Query(`SELECT id FROM server_groups WHERE auto_scale = 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*model.ServerGroup
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		group, err := r.GetGroupByID(id)
		if err != nil {
			continue
		}
		groups = append(groups, group)
	}

	return groups, nil
}

// UpdateGroup 更新分组
func (r *clusterServerRepository) UpdateGroup(group *model.ServerGroup) error {
	group.UpdatedAt = time.Now()
	_, err := r.db.Exec(`
		UPDATE server_groups SET name = ?, description = ?, auto_scale = ?, min_servers = ?, max_servers = ?, scale_policy = ?, scale_threshold = ?, updated_at = ?
		WHERE id = ?
	`, group.Name, group.Description, group.AutoScale, group.MinServers, group.MaxServers, group.ScalePolicy, group.ScaleThreshold, group.UpdatedAt, group.ID)
	return err
}

// DeleteGroup 删除分组
func (r *clusterServerRepository) DeleteGroup(id string) error {
	_, err := r.db.Exec(`DELETE FROM server_groups WHERE id = ?`, id)
	return err
}
