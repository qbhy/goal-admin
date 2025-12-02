package admin

import (
	"fmt"

	"github.com/goal-web/config"
	"github.com/goal-web/contracts"
	goaladmin "github.com/qbhy/goal-admin"
	admin2 "github.com/qbhy/goal-admin/models/admin"
	"github.com/qbhy/goal-admin/requests/admin"
	admin0 "github.com/qbhy/goal-admin/results/admin"
)

func init() {
	MenuServiceDefine.GetMenu = func(req *admin.MenuListReq, ctx contracts.Context) (*admin0.MenuListResult, error) {

		list := config.Get("admin").(goaladmin.Config).Resources
		menus := []admin2.Menu{
			{
				Name: "首页",
				Path: "/",
			},
			{
				Name: "管理员管理",
				Path: "/admin/admins",
			},
			{
				Name: "导出记录",
				Path: "/admin/resources/exports",
			},
		}

		for _, item := range list {
			menus = append(menus, admin2.Menu{
				Name: item.Meta().Title,
				Path: fmt.Sprintf("/admin/resources/%s", item.GetName()),
			})
		}

		return &admin0.MenuListResult{
			Menus: menus,
		}, nil
	}
}
