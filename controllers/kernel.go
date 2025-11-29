package controllers

import (
	"github.com/goal-web/contracts"
	"github.com/qbhy/goal-admin/controllers/admin"
)

// Register 注册路由函数
func Register(router contracts.HttpRouter) {
	admin.AuthServiceRouter(router)
	admin.AdminManageServiceRouter(router)
	admin.ResourceServiceRouter(router)
	admin.MenuServiceRouter(router)
	admin.SiteServiceRouter(router,
	)

	// 在这里添加您的路由注册逻辑
}
