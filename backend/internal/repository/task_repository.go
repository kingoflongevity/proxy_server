package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"proxy_server/internal/model"
)

// ScanRepository 扫描任务仓库接口
type ScanRepository interface {
	CreateScanTask(task *model.ScanTask) error
	GetScanTask(id string) (*model.ScanTask, error)
	UpdateScanTask(task *model.ScanTask) error
	GetScanHistory(limit int) ([]*model.ScanTask, error)
}

// DeployRepository 部署任务仓库接口
type DeployRepository interface {
	CreateDeployTask(task *model.DeployTask) error
	GetDeployTask(id string) (*model.DeployTask, error)
	UpdateDeployTask(task *model.DeployTask) error
	GetDeployHistory(serverID string, limit int) ([]*model.DeployTask, error)
}

// BackupRepository 备份仓库接口
type BackupRepository interface {
	CreateBackup(record *model.BackupRecord) error
	GetBackup(id string) (*model.BackupRecord, error)
	UpdateBackup(record *model.BackupRecord) error
	DeleteBackup(id string) error
	ListBackups(backupType model.BackupType, serverID string, limit int) ([]*model.BackupRecord, error)
}

// ScaleEventRepository 伸缩事件仓库接口
type ScaleEventRepository interface {
	CreateEvent(event *model.ScaleEvent) error
	GetEvent(id string) (*model.ScaleEvent, error)
	UpdateEvent(event *model.ScaleEvent) error
	GetEventsByGroupID(groupID string, limit int) ([]*model.ScaleEvent, error)
}

// ====== 扫描任务仓库实现 ======

type scanRepository struct {
	db *sql.DB
}

func NewScanRepository(db *sql.DB) ScanRepository {
	r := &scanRepository{db: db}
	r.initTable()
	return r
}

func (r *scanRepository) initTable() {
	r.db.Exec(`
		CREATE TABLE IF NOT EXISTS scan_tasks (
			id TEXT PRIMARY KEY,
			status TEXT DEFAULT 'pending',
			network_cidr TEXT NOT NULL,
			progress INTEGER DEFAULT 0,
			results TEXT,
			started_at DATETIME,
			completed_at DATETIME,
			created_at DATETIME
		)
	`)
}

func (r *scanRepository) CreateScanTask(task *model.ScanTask) error {
	resultsJSON, _ := json.Marshal(task.Results)
	_, err := r.db.Exec(`
		INSERT INTO scan_tasks (id, status, network_cidr, progress, results, started_at, completed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, task.ID, task.Status, task.NetworkCIDR, task.Progress, string(resultsJSON), task.StartedAt, task.CompletedAt, task.CreatedAt)
	return err
}

func (r *scanRepository) GetScanTask(id string) (*model.ScanTask, error) {
	task := &model.ScanTask{}
	var resultsJSON string
	var startedAt, completedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, status, network_cidr, progress, results, started_at, completed_at, created_at
		FROM scan_tasks WHERE id = ?
	`, id).Scan(&task.ID, &task.Status, &task.NetworkCIDR, &task.Progress, &resultsJSON, &startedAt, &completedAt, &task.CreatedAt)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(resultsJSON), &task.Results)
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	return task, nil
}

func (r *scanRepository) UpdateScanTask(task *model.ScanTask) error {
	resultsJSON, _ := json.Marshal(task.Results)
	_, err := r.db.Exec(`
		UPDATE scan_tasks SET status = ?, progress = ?, results = ?, started_at = ?, completed_at = ?
		WHERE id = ?
	`, task.Status, task.Progress, string(resultsJSON), task.StartedAt, task.CompletedAt, task.ID)
	return err
}

func (r *scanRepository) GetScanHistory(limit int) ([]*model.ScanTask, error) {
	rows, err := r.db.Query(`SELECT id FROM scan_tasks ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*model.ScanTask
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		task, err := r.GetScanTask(id)
		if err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// ====== 部署任务仓库实现 ======

type deployRepository struct {
	db *sql.DB
}

func NewDeployRepository(db *sql.DB) DeployRepository {
	r := &deployRepository{db: db}
	r.initTable()
	return r
}

func (r *deployRepository) initTable() {
	r.db.Exec(`
		CREATE TABLE IF NOT EXISTS deploy_tasks (
			id TEXT PRIMARY KEY,
			server_id TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			progress INTEGER DEFAULT 0,
			current_step TEXT,
			logs TEXT,
			started_at DATETIME,
			completed_at DATETIME,
			created_at DATETIME
		)
	`)
}

func (r *deployRepository) CreateDeployTask(task *model.DeployTask) error {
	logsJSON, _ := json.Marshal(task.Logs)
	_, err := r.db.Exec(`
		INSERT INTO deploy_tasks (id, server_id, status, progress, current_step, logs, started_at, completed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, task.ID, task.ServerID, task.Status, task.Progress, task.CurrentStep, string(logsJSON), task.StartedAt, task.CompletedAt, task.CreatedAt)
	return err
}

func (r *deployRepository) GetDeployTask(id string) (*model.DeployTask, error) {
	task := &model.DeployTask{}
	var logsJSON string
	var startedAt, completedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, server_id, status, progress, current_step, logs, started_at, completed_at, created_at
		FROM deploy_tasks WHERE id = ?
	`, id).Scan(&task.ID, &task.ServerID, &task.Status, &task.Progress, &task.CurrentStep, &logsJSON, &startedAt, &completedAt, &task.CreatedAt)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(logsJSON), &task.Logs)
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	return task, nil
}

func (r *deployRepository) UpdateDeployTask(task *model.DeployTask) error {
	logsJSON, _ := json.Marshal(task.Logs)
	_, err := r.db.Exec(`
		UPDATE deploy_tasks SET status = ?, progress = ?, current_step = ?, logs = ?, started_at = ?, completed_at = ?
		WHERE id = ?
	`, task.Status, task.Progress, task.CurrentStep, string(logsJSON), task.StartedAt, task.CompletedAt, task.ID)
	return err
}

func (r *deployRepository) GetDeployHistory(serverID string, limit int) ([]*model.DeployTask, error) {
	var rows *sql.Rows
	var err error

	if serverID != "" {
		rows, err = r.db.Query(`SELECT id FROM deploy_tasks WHERE server_id = ? ORDER BY created_at DESC LIMIT ?`, serverID, limit)
	} else {
		rows, err = r.db.Query(`SELECT id FROM deploy_tasks ORDER BY created_at DESC LIMIT ?`, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*model.DeployTask
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		task, err := r.GetDeployTask(id)
		if err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// ====== 备份仓库实现 ======

type backupRepository struct {
	db *sql.DB
}

func NewBackupRepository(db *sql.DB) BackupRepository {
	r := &backupRepository{db: db}
	r.initTable()
	return r
}

func (r *backupRepository) initTable() {
	r.db.Exec(`
		CREATE TABLE IF NOT EXISTS backup_records (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			server_id TEXT,
			file_path TEXT,
			file_size INTEGER,
			checksum TEXT,
			status TEXT DEFAULT 'creating',
			error_message TEXT,
			created_at DATETIME
		)
	`)
}

func (r *backupRepository) CreateBackup(record *model.BackupRecord) error {
	_, err := r.db.Exec(`
		INSERT INTO backup_records (id, name, type, server_id, file_path, file_size, checksum, status, error_message, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, record.ID, record.Name, record.Type, record.ServerID, record.FilePath, record.FileSize, record.Checksum, record.Status, record.ErrorMessage, record.CreatedAt)
	return err
}

func (r *backupRepository) GetBackup(id string) (*model.BackupRecord, error) {
	record := &model.BackupRecord{}
	err := r.db.QueryRow(`
		SELECT id, name, type, server_id, file_path, file_size, checksum, status, error_message, created_at
		FROM backup_records WHERE id = ?
	`, id).Scan(&record.ID, &record.Name, &record.Type, &record.ServerID, &record.FilePath, &record.FileSize, &record.Checksum, &record.Status, &record.ErrorMessage, &record.CreatedAt)
	return record, err
}

func (r *backupRepository) UpdateBackup(record *model.BackupRecord) error {
	_, err := r.db.Exec(`
		UPDATE backup_records SET file_path = ?, file_size = ?, checksum = ?, status = ?, error_message = ?
		WHERE id = ?
	`, record.FilePath, record.FileSize, record.Checksum, record.Status, record.ErrorMessage, record.ID)
	return err
}

func (r *backupRepository) DeleteBackup(id string) error {
	_, err := r.db.Exec(`DELETE FROM backup_records WHERE id = ?`, id)
	return err
}

func (r *backupRepository) ListBackups(backupType model.BackupType, serverID string, limit int) ([]*model.BackupRecord, error) {
	var rows *sql.Rows
	var err error

	if backupType != "" && serverID != "" {
		rows, err = r.db.Query(`SELECT id FROM backup_records WHERE type = ? AND server_id = ? ORDER BY created_at DESC LIMIT ?`, backupType, serverID, limit)
	} else if backupType != "" {
		rows, err = r.db.Query(`SELECT id FROM backup_records WHERE type = ? ORDER BY created_at DESC LIMIT ?`, backupType, limit)
	} else if serverID != "" {
		rows, err = r.db.Query(`SELECT id FROM backup_records WHERE server_id = ? ORDER BY created_at DESC LIMIT ?`, serverID, limit)
	} else {
		rows, err = r.db.Query(`SELECT id FROM backup_records ORDER BY created_at DESC LIMIT ?`, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*model.BackupRecord
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		record, err := r.GetBackup(id)
		if err != nil {
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

// ====== 伸缩事件仓库实现 ======

type scaleEventRepository struct {
	db *sql.DB
}

func NewScaleEventRepository(db *sql.DB) ScaleEventRepository {
	r := &scaleEventRepository{db: db}
	r.initTable()
	return r
}

func (r *scaleEventRepository) initTable() {
	r.db.Exec(`
		CREATE TABLE IF NOT EXISTS scale_events (
			id TEXT PRIMARY KEY,
			group_id TEXT NOT NULL,
			type TEXT NOT NULL,
			reason TEXT,
			target_count INTEGER,
			status TEXT DEFAULT 'pending',
			created_at DATETIME,
			completed_at DATETIME
		)
	`)
}

func (r *scaleEventRepository) CreateEvent(event *model.ScaleEvent) error {
	_, err := r.db.Exec(`
		INSERT INTO scale_events (id, group_id, type, reason, target_count, status, created_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, event.ID, event.GroupID, event.Type, event.Reason, event.TargetCount, event.Status, event.CreatedAt, event.CompletedAt)
	return err
}

func (r *scaleEventRepository) GetEvent(id string) (*model.ScaleEvent, error) {
	event := &model.ScaleEvent{}
	var completedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, group_id, type, reason, target_count, status, created_at, completed_at
		FROM scale_events WHERE id = ?
	`, id).Scan(&event.ID, &event.GroupID, &event.Type, &event.Reason, &event.TargetCount, &event.Status, &event.CreatedAt, &completedAt)

	if err != nil {
		return nil, err
	}

	if completedAt.Valid {
		event.CompletedAt = &completedAt.Time
	}

	return event, nil
}

func (r *scaleEventRepository) UpdateEvent(event *model.ScaleEvent) error {
	_, err := r.db.Exec(`
		UPDATE scale_events SET status = ?, completed_at = ?
		WHERE id = ?
	`, event.Status, event.CompletedAt, event.ID)
	return err
}

func (r *scaleEventRepository) GetEventsByGroupID(groupID string, limit int) ([]*model.ScaleEvent, error) {
	rows, err := r.db.Query(`SELECT id FROM scale_events WHERE group_id = ? ORDER BY created_at DESC LIMIT ?`, groupID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*model.ScaleEvent
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		event, err := r.GetEvent(id)
		if err != nil {
			continue
		}
		events = append(events, event)
	}

	return events, nil
}
