package admin

import (
	"github.com/goal-web/config"
	"github.com/goal-web/contracts"
	"github.com/qbhy/goal-admin/requests/admin"
	admin0 "github.com/qbhy/goal-admin/results/admin"
)

func init() {
	SiteServiceDefine.GetSiteConfig = func(req *admin.GetSiteConfigReq, ctx contracts.Context) (*admin0.SiteConfigResult, error) {

		return &admin0.SiteConfigResult{
			Logo:   config.StringOptional("admin.logo", "https://ui-avatars.com/api/?name=Goal&background=0d8abc&color=fff&size=128&bold=true"),
			Title:  config.StringOptional("admin.title", "Goal Admin"),
			Footer: config.StringOptional("admin.footer", "Powered by Goal"),
		}, nil
	}
}
