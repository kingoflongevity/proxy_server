package service

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"proxy_server/internal/model"
	"proxy_server/pkg/logger"
)

// BackupService 备份服务
type BackupService struct {
	backupDir string
}

// NewBackupService 创建备份服务
func NewBackupService(backupDir string) *BackupService {
	if backupDir == "" {
		backupDir = "./backups"
	}
	os.MkdirAll(backupDir, 0755)
	return &BackupService{backupDir: backupDir}
}

// CreateBackup 创建备份
func (b *BackupService) CreateBackup(backupType string, servers []model.ClusterServer, groups []model.ServerGroup) (*model.BackupRecord, error) {
	timestamp := time.Now().Format("20060102-150405")
	name := fmt.Sprintf("backup-%s-%s.tar.gz", backupType, timestamp)
	filePath := filepath.Join(b.backupDir, name)

	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建备份文件失败: %w", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	backupData := map[string]interface{}{
		"type":      backupType,
		"timestamp": time.Now().Unix(),
		"servers":   servers,
		"groups":    groups,
	}

	data, err := json.MarshalIndent(backupData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化数据失败: %w", err)
	}

	header := &tar.Header{
		Name:    "backup.json",
		Size:    int64(len(data)),
		Mode:    0644,
		ModTime: time.Now(),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return nil, fmt.Errorf("写入tar头失败: %w", err)
	}

	if _, err := tarWriter.Write(data); err != nil {
		return nil, fmt.Errorf("写入数据失败: %w", err)
	}

	stat, _ := file.Stat()
	md5Hash := b.calculateMD5(filePath)

	record := &model.BackupRecord{
		ID:        fmt.Sprintf("backup-%d", time.Now().UnixNano()),
		Type:      backupType,
		Name:      name,
		Size:      stat.Size(),
		MD5:       md5Hash,
		CreatedAt: time.Now(),
	}

	logger.Info("备份创建成功: %s, 大小: %d bytes", name, stat.Size())
	return record, nil
}

// RestoreBackup 恢复备份
func (b *BackupService) RestoreBackup(backupID string) ([]model.ClusterServer, []model.ServerGroup, error) {
	files, err := filepath.Glob(filepath.Join(b.backupDir, "backup-*-"+backupID+"*.tar.gz"))
	if err != nil || len(files) == 0 {
		files, err = filepath.Glob(filepath.Join(b.backupDir, backupID+"*.tar.gz"))
		if err != nil || len(files) == 0 {
			return nil, nil, fmt.Errorf("备份文件不存在: %s", backupID)
		}
	}

	file, err := os.Open(files[0])
	if err != nil {
		return nil, nil, fmt.Errorf("打开备份文件失败: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, nil, fmt.Errorf("解压失败: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var servers []model.ClusterServer
	var groups []model.ServerGroup

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("读取tar失败: %w", err)
		}

		if header.Name == "backup.json" {
			data := make([]byte, header.Size)
			if _, err := io.ReadFull(tarReader, data); err != nil {
				return nil, nil, fmt.Errorf("读取备份数据失败: %w", err)
			}

			var backupData struct {
				Servers []model.ClusterServer `json:"servers"`
				Groups  []model.ServerGroup   `json:"groups"`
			}

			if err := json.Unmarshal(data, &backupData); err != nil {
				return nil, nil, fmt.Errorf("解析备份数据失败: %w", err)
			}

			servers = backupData.Servers
			groups = backupData.Groups
		}
	}

	logger.Info("备份恢复成功: %s", backupID)
	return servers, groups, nil
}

// ListBackups 列出备份
func (b *BackupService) ListBackups() ([]model.BackupRecord, error) {
	files, err := filepath.Glob(filepath.Join(b.backupDir, "backup-*.tar.gz"))
	if err != nil {
		return nil, err
	}

	records := make([]model.BackupRecord, 0, len(files))
	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			continue
		}

		name := filepath.Base(file)
		parts := strings.Split(name, "-")
		backupType := "unknown"
		if len(parts) >= 2 {
			backupType = parts[1]
		}

		records = append(records, model.BackupRecord{
			ID:        strings.TrimSuffix(name, ".tar.gz"),
			Type:      backupType,
			Name:      name,
			Size:      stat.Size(),
			MD5:       b.calculateMD5(file),
			CreatedAt: stat.ModTime(),
		})
	}

	return records, nil
}

// DeleteBackup 删除备份
func (b *BackupService) DeleteBackup(backupID string) error {
	files, err := filepath.Glob(filepath.Join(b.backupDir, backupID+"*.tar.gz"))
	if err != nil || len(files) == 0 {
		return fmt.Errorf("备份文件不存在")
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}
	}

	logger.Info("备份已删除: %s", backupID)
	return nil
}

// BackupServerConfig 备份单个服务器配置
func (b *BackupService) BackupServerConfig(server *model.ClusterServer, sshManager *SSHManager) (*model.BackupRecord, error) {
	if !sshManager.CheckConnection(server.ID) {
		return nil, fmt.Errorf("服务器未连接")
	}

	configData, err := sshManager.ExecuteCommand(server.ID, "cat /opt/proxy-server/config.json")
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	name := fmt.Sprintf("server-%s-%s.json", server.ID, timestamp)
	filePath := filepath.Join(b.backupDir, "servers", name)
	os.MkdirAll(filepath.Dir(filePath), 0755)

	if err := os.WriteFile(filePath, []byte(configData), 0644); err != nil {
		return nil, fmt.Errorf("保存配置失败: %w", err)
	}

	stat, _ := os.Stat(filePath)

	record := &model.BackupRecord{
		ID:        fmt.Sprintf("server-%d", time.Now().UnixNano()),
		Type:      "proxy",
		ServerID:  server.ID,
		Name:      name,
		Size:      stat.Size(),
		MD5:       b.calculateMD5(filePath),
		CreatedAt: time.Now(),
	}

	return record, nil
}

// RestoreServerConfig 恢复服务器配置
func (b *BackupService) RestoreServerConfig(server *model.ClusterServer, backupID string, sshManager *SSHManager) error {
	files, err := filepath.Glob(filepath.Join(b.backupDir, "servers", backupID+"*.json"))
	if err != nil || len(files) == 0 {
		return fmt.Errorf("备份文件不存在")
	}

	data, err := os.ReadFile(files[0])
	if err != nil {
		return fmt.Errorf("读取备份失败: %w", err)
	}

	if !sshManager.CheckConnection(server.ID) {
		return fmt.Errorf("服务器未连接")
	}

	if err := sshManager.UploadData(server.ID, data, "/opt/proxy-server/config.json"); err != nil {
		return fmt.Errorf("上传配置失败: %w", err)
	}

	_, err = sshManager.ExecuteCommand(server.ID, "systemctl restart proxy-server")
	if err != nil {
		return fmt.Errorf("重启服务失败: %w", err)
	}

	logger.Info("服务器 %s 配置已恢复", server.ID)
	return nil
}

// ScheduleBackup 定时备份
func (b *BackupService) ScheduleBackup(interval time.Duration, getServerFunc func() ([]model.ClusterServer, error), getGroupFunc func() ([]model.ServerGroup, error)) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			servers, err := getServerFunc()
			if err != nil {
				logger.Warn("获取服务器列表失败: %v", err)
				continue
			}

			groups, err := getGroupFunc()
			if err != nil {
				logger.Warn("获取分组列表失败: %v", err)
				continue
			}

			_, err = b.CreateBackup("scheduled", servers, groups)
			if err != nil {
				logger.Warn("定时备份失败: %v", err)
			}
		}
	}()
}

// CleanupOldBackups 清理旧备份
func (b *BackupService) CleanupOldBackups(maxAge time.Duration) error {
	files, err := filepath.Glob(filepath.Join(b.backupDir, "backup-*.tar.gz"))
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-maxAge)
	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			continue
		}

		if stat.ModTime().Before(cutoff) {
			os.Remove(file)
			logger.Info("清理旧备份: %s", filepath.Base(file))
		}
	}

	return nil
}

// calculateMD5 计算MD5
func (b *BackupService) calculateMD5(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return hex.EncodeToString(hash.Sum(nil))
}
