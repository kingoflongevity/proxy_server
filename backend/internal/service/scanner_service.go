package service

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"proxy_server/internal/model"
	"proxy_server/pkg/logger"
)

// ScannerService 局域网扫描服务
type ScannerService struct {
	sshManager *SSHManager
}

// NewScannerService 创建扫描服务
func NewScannerService(sshManager *SSHManager) *ScannerService {
	return &ScannerService{
		sshManager: sshManager,
	}
}

// ScanResult 扫描结果
type ScanResult struct {
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	OSType    string `json:"osType"`
	Hostname  string `json:"hostname"`
	Available bool   `json:"available"`
}

// ScanCIDR 扫描CIDR网段
func (s *ScannerService) ScanCIDR(ctx context.Context, cidr string, workers int, progressChan chan<- float64) ([]ScanResult, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("无效的CIDR格式: %w", err)
	}

	var ips []string
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incIP(ip) {
		ips = append(ips, ip.String())
	}

	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}

	if workers <= 0 {
		workers = 50
	}

	results := make([]ScanResult, 0)
	resultChan := make(chan ScanResult, workers*2)
	taskChan := make(chan string, workers*2)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range taskChan {
				select {
				case <-ctx.Done():
					return
				default:
					if result := s.scanHost(ip); result.Available {
						resultChan <- result
					}
				}
			}
		}()
	}

	go func() {
		for _, ip := range ips {
			select {
			case <-ctx.Done():
				close(taskChan)
				return
			default:
				taskChan <- ip
			}
		}
		close(taskChan)
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	completed := 0
	total := len(ips)

	for result := range resultChan {
		mu.Lock()
		results = append(results, result)
		mu.Unlock()

		completed++
		if progressChan != nil {
			select {
			case progressChan <- float64(completed) / float64(total) * 100:
			default:
			}
		}
	}

	if progressChan != nil {
		close(progressChan)
	}

	return results, nil
}

// scanHost 扫描单个主机
func (s *ScannerService) scanHost(ip string) ScanResult {
	result := ScanResult{
		IP:        ip,
		Port:      22,
		Available: false,
	}

	timeout := time.Second * 2
	addr := fmt.Sprintf("%s:22", ip)

	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return result
	}
	conn.Close()

	result.Available = true

	banner := s.grabBanner(ip, 22, timeout)
	if strings.Contains(strings.ToLower(banner), "ubuntu") {
		result.OSType = "ubuntu"
	} else if strings.Contains(strings.ToLower(banner), "debian") {
		result.OSType = "debian"
	} else if strings.Contains(strings.ToLower(banner), "centos") {
		result.OSType = "centos"
	} else if strings.Contains(strings.ToLower(banner), "alpine") {
		result.OSType = "alpine"
	} else {
		result.OSType = "linux"
	}

	return result
}

// grabBanner 获取服务Banner
func (s *ScannerService) grabBanner(ip string, port int, timeout time.Duration) string {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		return ""
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(timeout))

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return ""
	}

	return string(buf[:n])
}

// QuickScan 快速扫描常用端口
func (s *ScannerService) QuickScan(ip string, ports []int) map[int]bool {
	if len(ports) == 0 {
		ports = []int{22, 80, 443, 8080, 10808, 10809, 10810}
	}

	results := make(map[int]bool)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			addr := fmt.Sprintf("%s:%d", ip, p)
			conn, err := net.DialTimeout("tcp", addr, time.Second*2)
			if err == nil {
				conn.Close()
				mu.Lock()
				results[p] = true
				mu.Unlock()
			} else {
				mu.Lock()
				results[p] = false
				mu.Unlock()
			}
		}(port)
	}

	wg.Wait()
	return results
}

// TestSSH 测试SSH连接
func (s *ScannerService) TestSSH(ip string, port int, username, password string) (int, error) {
	config := SSHConfig{
		Host:     ip,
		Port:     port,
		Username: username,
		Password: password,
		Timeout:  10 * time.Second,
	}

	return s.sshManager.TestConnection(config)
}

// incIP IP地址递增
func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// ParseCIDR 解析CIDR获取IP列表
func ParseCIDR(cidr string) ([]string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incIP(ip) {
		ips = append(ips, ip.String())
	}

	if len(ips) > 2 {
		return ips[1 : len(ips)-1], nil
	}

	return ips, nil
}

// GetLocalNetwork 获取本地网络信息
func GetLocalNetwork() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var networks []string
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
				ones, _ := ipNet.Mask.Size()
				cidr := fmt.Sprintf("%s/%d", ipNet.IP.String(), ones)
				networks = append(networks, cidr)
			}
		}
	}

	return networks, nil
}

// DetectOS 通过SSH检测操作系统
func (s *ScannerService) DetectOS(ip string, port int, username, password string) (string, error) {
	config := SSHConfig{
		Host:     ip,
		Port:     port,
		Username: username,
		Password: password,
		Timeout:  10 * time.Second,
	}

	serverID := fmt.Sprintf("detect-%s", ip)
	if err := s.sshManager.Connect(serverID, config); err != nil {
		return "", err
	}
	defer s.sshManager.Disconnect(serverID)

	output, err := s.sshManager.ExecuteCommand(serverID, "cat /etc/os-release | grep '^ID=' | cut -d'=' -f2 | tr -d '\"'")
	if err != nil {
		return "", err
	}

	osType := strings.TrimSpace(strings.ToLower(output))
	return osType, nil
}

// GetServerResources 获取服务器资源信息
func (s *ScannerService) GetServerResources(ip string, port int, username, password string) (map[string]interface{}, error) {
	config := SSHConfig{
		Host:     ip,
		Port:     port,
		Username: username,
		Password: password,
		Timeout:  10 * time.Second,
	}

	serverID := fmt.Sprintf("resource-%s", ip)
	if err := s.sshManager.Connect(serverID, config); err != nil {
		return nil, err
	}
	defer s.sshManager.Disconnect(serverID)

	resources := make(map[string]interface{})

	cpuOutput, err := s.sshManager.ExecuteCommand(serverID, "top -bn1 | grep 'Cpu(s)' | awk '{print $2}' | cut -d'%' -f1")
	if err == nil {
		cpu, _ := strconv.ParseFloat(strings.TrimSpace(cpuOutput), 64)
		resources["cpu"] = cpu
	}

	memOutput, err := s.sshManager.ExecuteCommand(serverID, "free -m | awk '/Mem:/ {print $3}'")
	if err == nil {
		mem, _ := strconv.ParseUint(strings.TrimSpace(memOutput), 10, 64)
		resources["memory_used"] = mem
	}

	memTotal, err := s.sshManager.ExecuteCommand(serverID, "free -m | awk '/Mem:/ {print $2}'")
	if err == nil {
		total, _ := strconv.ParseUint(strings.TrimSpace(memTotal), 10, 64)
		resources["memory_total"] = total
	}

	arch, err := s.sshManager.ExecuteCommand(serverID, "uname -m")
	if err == nil {
		resources["arch"] = strings.TrimSpace(arch)
	}

	hostname, err := s.sshManager.ExecuteCommand(serverID, "hostname")
	if err == nil {
		resources["hostname"] = strings.TrimSpace(hostname)
	}

	return resources, nil
}
