package sysproxy

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"proxy_server/pkg/logger"
)

type ProxyConfig struct {
	Enabled  bool
	Server   string
	Port     int
	Bypass   string
}

type SystemProxyManager struct {
	originalProxy *ProxyConfig
}

func NewSystemProxyManager() *SystemProxyManager {
	return &SystemProxyManager{}
}

// EnableSystemProxy 启用系统代理
func (m *SystemProxyManager) EnableSystemProxy(server string, port int) error {
	if runtime.GOOS != "windows" {
		logger.Warn("系统代理设置仅支持Windows系统")
		return nil
	}

	proxyServer := fmt.Sprintf("%s:%d", server, port)
	bypass := "localhost;127.*;10.*;172.16.*;172.31.*;192.168.*;<local>"

	// 保存原始代理设置
	original, _ := m.GetCurrentProxy()
	m.originalProxy = original

	// 设置代理
	commands := [][]string{
		{"add", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings", "/v", "ProxyEnable", "/t", "REG_DWORD", "/d", "1", "/f"},
		{"add", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings", "/v", "ProxyServer", "/t", "REG_SZ", "/d", proxyServer, "/f"},
		{"add", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings", "/v", "ProxyOverride", "/t", "REG_SZ", "/d", bypass, "/f"},
	}

	for _, args := range commands {
		cmd := exec.Command("reg", args...)
		if err := cmd.Run(); err != nil {
			logger.Error("设置注册表失败: %v, 命令: %v", err, args)
			return fmt.Errorf("设置系统代理失败: %w", err)
		}
	}

	// 刷新系统设置
	if err := m.refreshSystemProxy(); err != nil {
		logger.Warn("刷新系统代理设置失败: %v", err)
	}

	logger.Info("系统代理已启用: %s", proxyServer)
	return nil
}

// DisableSystemProxy 禁用系统代理
func (m *SystemProxyManager) DisableSystemProxy() error {
	if runtime.GOOS != "windows" {
		logger.Warn("系统代理设置仅支持Windows系统")
		return nil
	}

	cmd := exec.Command("reg", "add", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings", "/v", "ProxyEnable", "/t", "REG_DWORD", "/d", "0", "/f")
	if err := cmd.Run(); err != nil {
		logger.Error("禁用系统代理失败: %v", err)
		return fmt.Errorf("禁用系统代理失败: %w", err)
	}

	// 刷新系统设置
	if err := m.refreshSystemProxy(); err != nil {
		logger.Warn("刷新系统代理设置失败: %v", err)
	}

	logger.Info("系统代理已禁用")
	return nil
}

// GetCurrentProxy 获取当前系统代理设置
func (m *SystemProxyManager) GetCurrentProxy() (*ProxyConfig, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("系统代理设置仅支持Windows系统")
	}

	config := &ProxyConfig{}

	// 获取ProxyEnable
	cmd := exec.Command("reg", "query", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings", "/v", "ProxyEnable")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "ProxyEnable") {
				if strings.Contains(line, "REG_DWORD") && strings.Contains(line, "0x1") {
					config.Enabled = true
				}
				break
			}
		}
	}

	// 获取ProxyServer
	cmd = exec.Command("reg", "query", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings", "/v", "ProxyServer")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "ProxyServer") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					serverStr := parts[len(parts)-1]
					parts = strings.Split(serverStr, ":")
					if len(parts) == 2 {
						config.Server = parts[0]
						fmt.Sscanf(parts[1], "%d", &config.Port)
					}
				}
				break
			}
		}
	}

	// 获取ProxyOverride
	cmd = exec.Command("reg", "query", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings", "/v", "ProxyOverride")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "ProxyOverride") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					config.Bypass = parts[len(parts)-1]
				}
				break
			}
		}
	}

	return config, nil
}

// refreshSystemProxy 刷新系统代理设置
// 这会通知系统代理设置已更改
func (m *SystemProxyManager) refreshSystemProxy() error {
	// 使用 PowerShell 刷新系统设置
	psScript := `
Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;
public class Internet {
    [DllImport("wininet.dll")]
    public static extern bool InternetSetOption(IntPtr hInternet, int dwOption, IntPtr lpBuffer, int dwBufferLength);
}
"@
[Internet]::InternetSetOption([IntPtr]::Zero, 39, [IntPtr]::Zero, 0)
[Internet]::InternetSetOption([IntPtr]::Zero, 37, [IntPtr]::Zero, 0)
`
	cmd := exec.Command("powershell", "-Command", psScript)
	return cmd.Run()
}

// RestoreOriginalProxy 恢复原始代理设置
func (m *SystemProxyManager) RestoreOriginalProxy() error {
	if m.originalProxy == nil {
		return m.DisableSystemProxy()
	}

	if m.originalProxy.Enabled {
		return m.EnableSystemProxy(m.originalProxy.Server, m.originalProxy.Port)
	}
	return m.DisableSystemProxy()
}
