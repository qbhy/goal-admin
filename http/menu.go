package http

import (
	"github.com/goal-web/contracts"
	"github.com/qbhy/goal-admin/resources"
)

func GetMenuList(factory resources.Factory) any {
	return contracts.Fields{
		"data": factory.GetMenuList(),
	}
}

type SaveMenuForm struct {
	List []resources.MenuDataItem `json:"list"`
}

func SaveMenuList(factory resources.Factory, request contracts.HttpRequest) any {
	var form SaveMenuForm
	if err := request.Parse(&form); err != nil {
		return contracts.Fields{"err_msg": err.Error()}
	}

	err := factory.SaveMenuList(form.List)
	if err != nil {
		return contracts.Fields{"err_msg": err.Error()}
	}

	return contracts.Fields{"data": nil}
}
