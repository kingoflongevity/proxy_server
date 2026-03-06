package router

import (
	"proxy_server/internal/handler"
	"proxy_server/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter(
	subHandler *handler.SubscriptionHandler,
	nodeHandler *handler.NodeHandler,
	ruleHandler *handler.RuleHandler,
	systemHandler *handler.SystemHandler,
) *gin.Engine {
	r := gin.New()
	
	// 使用中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(cors.Default())
	
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Service is running",
		})
	})
	
	// API路由组
	api := r.Group("/api")
	{
		// 订阅管理
		subscriptions := api.Group("/subscriptions")
		{
			subscriptions.POST("", subHandler.Create)
			subscriptions.GET("", subHandler.GetAll)
			subscriptions.GET("/:id", subHandler.GetByID)
			subscriptions.PUT("/:id", subHandler.Update)
			subscriptions.DELETE("/:id", subHandler.Delete)
			subscriptions.POST("/:id/refresh", subHandler.Refresh)
		}
		
		// 节点管理
		nodes := api.Group("/nodes")
		{
			nodes.GET("", nodeHandler.GetList)
			nodes.GET("/:id", nodeHandler.GetByID)
			nodes.PUT("/:id", nodeHandler.Update)
			nodes.POST("/:id/test", nodeHandler.Test)
			nodes.POST("/connect", nodeHandler.Connect)
			nodes.POST("/disconnect", nodeHandler.Disconnect)
		}
		
		// 规则管理
		rules := api.Group("/rules")
		{
			rules.GET("", ruleHandler.GetAll)
			rules.POST("", ruleHandler.Create)
			rules.GET("/:id", ruleHandler.GetByID)
			rules.PUT("/:id", ruleHandler.Update)
			rules.DELETE("/:id", ruleHandler.Delete)
		}
		
		// 系统状态
		api.GET("/status", systemHandler.GetStatus)
		api.GET("/traffic", systemHandler.GetTraffic)
		api.GET("/logs", systemHandler.GetLogs)
	}
	
	return r
}
