package http

import (
	"fmt"
	"github.com/goal-web/contracts"
	"github.com/qbhy/goal-admin/resources"
)

func GetMenuList(factory resources.Factory) any {

	var menus []contracts.Fields

	for name, res := range factory.GetResources() {
		menus = append(menus, contracts.Fields{
			"name": res.GetTitle(),
			"path": fmt.Sprintf(`/resource/%s/list`, name),
		})
	}

	return contracts.Fields{
		"data": menus,
	}
}
