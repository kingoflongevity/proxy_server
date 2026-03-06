package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"proxy_server/internal/config"
	"proxy_server/internal/handler"
	"proxy_server/internal/repository"
	"proxy_server/internal/router"
	"proxy_server/internal/service"
	"proxy_server/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.GetConfig()
	if err := config.LoadConfig(config.GetDefaultConfigPath()); err != nil {
		logger.Warn("加载配置文件失败，使用默认配置: %v", err)
	}
	
	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	// 初始化数据目录
	dataDir := cfg.Database.Path
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		logger.Error("创建数据目录失败: %v", err)
		os.Exit(1)
	}
	
	// 初始化Repository层
	subRepo := repository.NewSubscriptionRepository(dataDir)
	nodeRepo := repository.NewNodeRepository(dataDir)
	ruleRepo := repository.NewRuleRepository(dataDir)
	systemRepo := repository.NewSystemRepository(dataDir)
	
	// 初始化Service层
	subService := service.NewSubscriptionService(subRepo, nodeRepo)
	nodeService := service.NewNodeService(nodeRepo, systemRepo)
	ruleService := service.NewRuleService(ruleRepo)
	systemService := service.NewSystemService(systemRepo)
	logService := service.NewLogService()
	
	// 初始化Handler层
	subHandler := handler.NewSubscriptionHandler(subService)
	nodeHandler := handler.NewNodeHandler(nodeService)
	ruleHandler := handler.NewRuleHandler(ruleService)
	systemHandler := handler.NewSystemHandler(systemService)
	logHandler := handler.NewLogHandler(logService)
	
	// 设置路由
	r := router.SetupRouter(subHandler, nodeHandler, ruleHandler, systemHandler, logHandler)
	
	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}
	
	// 启动服务器（在goroutine中）
	go func() {
		logger.Info("服务器启动在 %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("服务器启动失败: %v", err)
			os.Exit(1)
		}
	}()
	
	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("正在关闭服务器...")
	
	// 给5秒钟时间处理未完成的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务器强制关闭: %v", err)
	}
	
	logger.Info("服务器已关闭")
}
