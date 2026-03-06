package router

import (
	"proxy_server/internal/handler"
	"proxy_server/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	subHandler *handler.SubscriptionHandler,
	nodeHandler *handler.NodeHandler,
	ruleHandler *handler.RuleHandler,
	systemHandler *handler.SystemHandler,
	logHandler *handler.LogHandler,
) *gin.Engine {
	r := gin.New()
	
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(cors.Default())
	
	// 流量日志中间件
	r.Use(middleware.TrafficLoggerMiddleware())
	
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Service is running",
		})
	})
	
	api := r.Group("/api")
	{
		subscriptions := api.Group("/subscriptions")
		{
			subscriptions.POST("", subHandler.Create)
			subscriptions.GET("", subHandler.GetAll)
			subscriptions.GET("/:id", subHandler.GetByID)
			subscriptions.PUT("/:id", subHandler.Update)
			subscriptions.DELETE("/:id", subHandler.Delete)
			subscriptions.POST("/:id/refresh", subHandler.Refresh)
			subscriptions.POST("/:id/update", subHandler.Refresh)
			subscriptions.POST("/:id/test", subHandler.Test)
		}
		
		nodes := api.Group("/nodes")
		{
			nodes.GET("", nodeHandler.GetList)
			nodes.GET("/current", nodeHandler.GetCurrent)
			nodes.GET("/:id", nodeHandler.GetByID)
			nodes.PUT("/:id", nodeHandler.Update)
			nodes.POST("/:id/test", nodeHandler.Test)
			nodes.POST("/:id/select", nodeHandler.Select)
			nodes.POST("/connect", nodeHandler.Connect)
			nodes.POST("/disconnect", nodeHandler.Disconnect)
			nodes.GET("/:id/stats", nodeHandler.GetStats)
			nodes.POST("/test", nodeHandler.TestBatch)
			nodes.POST("/test-all", nodeHandler.TestAll)
		}
		
		rules := api.Group("/rules")
		{
			rules.GET("", ruleHandler.GetAll)
			rules.POST("", ruleHandler.Create)
			rules.GET("/:id", ruleHandler.GetByID)
			rules.PUT("/:id", ruleHandler.Update)
			rules.DELETE("/:id", ruleHandler.Delete)
			rules.PUT("/priority", ruleHandler.UpdatePriority)
		}
		
		api.GET("/rule-sets", ruleHandler.GetRuleSets)
		api.POST("/rule-sets/:id/update", ruleHandler.UpdateRuleSet)
		
		api.GET("/status", systemHandler.GetStatus)
		api.GET("/traffic", systemHandler.GetTraffic)
		api.GET("/system/logs", systemHandler.GetLogs)
		
		// 日志查询API
		logs := api.Group("/traffic-logs")
		{
			logs.GET("", logHandler.QueryLogs)
			logs.GET("/stats", logHandler.GetLogStats)
			logs.DELETE("", logHandler.ClearLogs)
		}
		
		// 流量统计API
		traffic := api.Group("/traffic-stats")
		{
			traffic.GET("/logs", logHandler.GetTrafficLogs)
			traffic.GET("/summary", logHandler.GetTrafficStats)
		}
		
		api.GET("/settings", systemHandler.GetSettings)
		api.PUT("/settings", systemHandler.UpdateSettings)
		api.GET("/connection/status", systemHandler.GetConnectionStatus)
		api.GET("/system/info", systemHandler.GetSystemInfo)
		api.POST("/system/restart", systemHandler.RestartService)
		api.GET("/config/export", systemHandler.ExportConfig)
		api.POST("/config/import", systemHandler.ImportConfig)
		api.POST("/system/clear-cache", systemHandler.ClearCache)
	}
	
	return r
}
