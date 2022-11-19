package router

import (
	"github.com/Project-Uranus-plus/all-eyes/internal/pkg/core"
)

func setApiRouter(r *resource) {

	// 需要签名验证、登录验证、RBAC 权限验证
	api := r.mux.Group("/api", core.WrapAuthHandler(r.interceptors.CheckLogin), r.interceptors.CheckSignature(), r.interceptors.CheckRBAC())
	{
		// authorized
		authorizedHandler := authorized.New(r.logger, r.db, r.cache)
		api.POST("/authorized", authorizedHandler.Create())
		api.GET("/authorized", authorizedHandler.List())
		api.PATCH("/authorized/used", authorizedHandler.UpdateUsed())
		api.DELETE("/authorized/:id", core.AliasForRecordMetrics("/api/authorized/info"), authorizedHandler.Delete())

		api.POST("/authorized_api", authorizedHandler.CreateAPI())
		api.GET("/authorized_api", authorizedHandler.ListAPI())
		api.DELETE("/authorized_api/:id", core.AliasForRecordMetrics("/api/authorized_api/info"), authorizedHandler.DeleteAPI())


	}
}
