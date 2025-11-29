package admin

import (
	"github.com/goal-web/contracts"
	admin2 "github.com/qbhy/goal-admin/models/admin"
	"github.com/qbhy/goal-admin/requests/admin"
	admin0 "github.com/qbhy/goal-admin/results/admin"
)

func init() {
	MenuServiceDefine.GetMenu = func(req *admin.MenuListReq, ctx contracts.Context) (*admin0.MenuListResult, error) {

		return &admin0.MenuListResult{
			Menus: []admin2.Menu{
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
			},
		}, nil
	}
}
