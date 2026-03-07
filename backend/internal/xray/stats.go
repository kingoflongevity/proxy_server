package xray

import (
	"context"
	"fmt"
	"time"

	"proxy_server/pkg/logger"

	"github.com/xtls/xray-core/app/stats/command"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// StatsClient Xray统计客户端
type StatsClient struct {
	apiPort int
	conn    *grpc.ClientConn
	client  command.StatsServiceClient
}

// NewStatsClient 创建统计客户端
func NewStatsClient(apiPort int) *StatsClient {
	return &StatsClient{
		apiPort: apiPort,
	}
}

// Connect 连接到Xray API
func (c *StatsClient) Connect() error {
	if c.conn != nil {
		return nil
	}

	addr := fmt.Sprintf("127.0.0.1:%d", c.apiPort)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("连接Xray API失败: %w", err)
	}

	c.conn = conn
	c.client = command.NewStatsServiceClient(conn)
	return nil
}

// Close 关闭连接
func (c *StatsClient) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		c.client = nil
	}
}

// TrafficStats 流量统计
type TrafficStats struct {
	Upload   int64
	Download int64
}

// GetInboundTraffic 获取入站流量统计
func (c *StatsClient) GetInboundTraffic(tag string) (*TrafficStats, error) {
	if err := c.Connect(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 获取上传流量
	uploadResp, err := c.client.GetStats(ctx, &command.GetStatsRequest{
		Name: fmt.Sprintf("inbound>>>%s>>>traffic>>>uplink", tag),
	})
	upload := int64(0)
	if err == nil && uploadResp != nil && uploadResp.Stat != nil {
		upload = uploadResp.Stat.Value
	}

	// 获取下载流量
	downloadResp, err := c.client.GetStats(ctx, &command.GetStatsRequest{
		Name: fmt.Sprintf("inbound>>>%s>>>traffic>>>downlink", tag),
	})
	download := int64(0)
	if err == nil && downloadResp != nil && downloadResp.Stat != nil {
		download = downloadResp.Stat.Value
	}

	return &TrafficStats{
		Upload:   upload,
		Download: download,
	}, nil
}

// GetOutboundTraffic 获取出站流量统计
func (c *StatsClient) GetOutboundTraffic(tag string) (*TrafficStats, error) {
	if err := c.Connect(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 获取上传流量
	uploadResp, err := c.client.GetStats(ctx, &command.GetStatsRequest{
		Name: fmt.Sprintf("outbound>>>%s>>>traffic>>>uplink", tag),
	})
	upload := int64(0)
	if err == nil && uploadResp != nil && uploadResp.Stat != nil {
		upload = uploadResp.Stat.Value
	}

	// 获取下载流量
	downloadResp, err := c.client.GetStats(ctx, &command.GetStatsRequest{
		Name: fmt.Sprintf("outbound>>>%s>>>traffic>>>downlink", tag),
	})
	download := int64(0)
	if err == nil && downloadResp != nil && downloadResp.Stat != nil {
		download = downloadResp.Stat.Value
	}

	return &TrafficStats{
		Upload:   upload,
		Download: download,
	}, nil
}

// GetAllTraffic 获取所有流量统计
func (c *StatsClient) GetAllTraffic() (map[string]*TrafficStats, error) {
	if err := c.Connect(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.QueryStats(ctx, &command.QueryStatsRequest{})
	if err != nil {
		return nil, fmt.Errorf("查询统计失败: %w", err)
	}

	stats := make(map[string]*TrafficStats)
	for _, stat := range resp.GetStat() {
		logger.Debug("统计: %s = %d", stat.Name, stat.Value)
		// 解析统计名称格式: inbound>>>tag>>>traffic>>>uplink/downlink
		// 或 outbound>>>tag>>>traffic>>>uplink/downlink
	}

	return stats, nil
}
