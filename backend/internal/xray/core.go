package xray

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"proxy_server/pkg/logger"
)

// CoreInfo 内核信息
type CoreInfo struct {
	Version     string `json:"version"`
	InstallPath string `json:"installPath"`
	DownloadURL string `json:"downloadUrl"`
	Installed   bool   `json:"installed"`
}

// ReleaseInfo GitHub Release 信息
type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

// CoreManager 内核管理器
type CoreManager struct {
	corePath     string
	installDir   string
	githubAPIURL string
}

// NewCoreManager 创建内核管理器
func NewCoreManager() *CoreManager {
	installDir := filepath.Join(os.Getenv("APPDATA"), "proxy-server", "core")
	if runtime.GOOS != "windows" {
		installDir = filepath.Join(os.Getenv("HOME"), ".config", "proxy-server", "core")
	}

	corePath := filepath.Join(installDir, "xray")
	if runtime.GOOS == "windows" {
		corePath += ".exe"
	}

	return &CoreManager{
		corePath:     corePath,
		installDir:   installDir,
		githubAPIURL: "https://api.github.com/repos/XTLS/Xray-core/releases/latest",
	}
}

// GetInstalledVersion 获取已安装的内核版本
func (cm *CoreManager) GetInstalledVersion() string {
	if _, err := os.Stat(cm.corePath); os.IsNotExist(err) {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, cm.corePath, "version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	re := regexp.MustCompile(`Xray\s+(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		return matches[1]
	}

	return "unknown"
}

// GetLatestVersion 获取最新版本信息
func (cm *CoreManager) GetLatestVersion() (*ReleaseInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(cm.githubAPIURL)
	if err != nil {
		return nil, fmt.Errorf("获取最新版本失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败: %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("解析版本信息失败: %w", err)
	}

	return &release, nil
}

// GetCoreInfo 获取内核信息
func (cm *CoreManager) GetCoreInfo() (*CoreInfo, error) {
	info := &CoreInfo{
		InstallPath: cm.corePath,
	}

	info.Version = cm.GetInstalledVersion()
	info.Installed = info.Version != ""

	latest, err := cm.GetLatestVersion()
	if err == nil {
		info.DownloadURL = cm.getDownloadURL(latest)
	}

	return info, nil
}

// getDownloadURL 获取适合当前系统的下载链接
func (cm *CoreManager) getDownloadURL(release *ReleaseInfo) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var pattern string
	switch {
	case goos == "windows" && goarch == "amd64":
		pattern = "Xray-windows-64"
	case goos == "windows" && goarch == "arm64":
		pattern = "Xray-windows-arm64"
	case goos == "darwin" && goarch == "amd64":
		pattern = "Xray-macos-64"
	case goos == "darwin" && goarch == "arm64":
		pattern = "Xray-macos-arm64"
	case goos == "linux" && goarch == "amd64":
		pattern = "Xray-linux-64"
	case goos == "linux" && goarch == "arm64":
		pattern = "Xray-linux-arm64"
	case goos == "linux" && goarch == "arm":
		pattern = "Xray-linux-32"
	default:
		pattern = "Xray-" + goos + "-" + goarch
	}

	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, pattern) {
			return asset.URL
		}
	}

	return ""
}

// DownloadCore 下载内核
func (cm *CoreManager) DownloadCore(progressCallback func(progress int)) error {
	release, err := cm.GetLatestVersion()
	if err != nil {
		return err
	}

	downloadURL := cm.getDownloadURL(release)
	if downloadURL == "" {
		return fmt.Errorf("未找到适合当前系统的内核版本")
	}

	if err := os.MkdirAll(cm.installDir, 0755); err != nil {
		return fmt.Errorf("创建安装目录失败: %w", err)
	}

	zipPath := filepath.Join(cm.installDir, "xray.zip")

	if err := cm.downloadFile(downloadURL, zipPath, progressCallback); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer os.Remove(zipPath)

	if err := cm.extractCore(zipPath); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	logger.Info("Xray内核更新完成，版本: %s", release.TagName)
	return nil
}

// downloadFile 下载文件
func (cm *CoreManager) downloadFile(url, filePath string, progressCallback func(int)) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: %d", resp.StatusCode)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	total := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 32*1024)

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, err := file.Write(buf[:n]); err != nil {
				return err
			}
			downloaded += int64(n)
			if total > 0 && progressCallback != nil {
				progress := int(float64(downloaded) / float64(total) * 100)
				progressCallback(progress)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// extractCore 解压内核文件
func (cm *CoreManager) extractCore(zipPath string) error {
	_ = os.Remove(cm.corePath)

	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf("Expand-Archive -Path '%s' -DestinationPath '%s' -Force", zipPath, cm.installDir))
		if err := cmd.Run(); err != nil {
			return err
		}
	} else {
		cmd := exec.Command("unzip", "-o", zipPath, "-d", cm.installDir)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	extractedCore := filepath.Join(cm.installDir, "xray")
	if runtime.GOOS == "windows" {
		extractedCore += ".exe"
	}

	if _, err := os.Stat(extractedCore); err == nil {
		return os.Chmod(extractedCore, 0755)
	}

	return nil
}

// UploadCore 上传内核文件
func (cm *CoreManager) UploadCore(fileData []byte) error {
	if err := os.MkdirAll(cm.installDir, 0755); err != nil {
		return fmt.Errorf("创建安装目录失败: %w", err)
	}

	_ = os.Remove(cm.corePath)

	if err := os.WriteFile(cm.corePath, fileData, 0755); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	logger.Info("Xray内核上传完成")
	return nil
}

// GetCorePath 获取内核路径
func (cm *CoreManager) GetCorePath() string {
	return cm.corePath
}
