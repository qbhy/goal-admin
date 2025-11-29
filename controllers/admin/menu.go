package admin

import (
	"github.com/goal-web/contracts"
	"github.com/qbhy/goal-admin/response"
)

type MenuItem struct {
	Path       string     `json:"path"`
	Name       string     `json:"name,omitempty"`
	Icon       string     `json:"icon,omitempty"`
	HideInMenu bool       `json:"hideInMenu,omitempty"`
	Children   []MenuItem `json:"children,omitempty"`
}

func AdminMenuRouter(router contracts.HttpRouter) {
	group := router.Group("/api/admin", "auth:admin_jwt")
	group.GET("/menu", GetAdminMenu)
}

func GetAdminMenu(request contracts.HttpRequest) any {
	menu := []MenuItem{
		{
			Path: "/admin/resources/user",
			Name: "用户管理",
			Icon: "user",
			Children: []MenuItem{
				{Path: "/admin/resources/user/users", Name: "用户列表", Icon: "user"},
			},
		},
	}
	return response.Success(menu)
}
