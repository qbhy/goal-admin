package http

import (
	"github.com/goal-web/contracts"
	"github.com/qbhy/goal-admin/resources"
)

func GetResourceList(request contracts.HttpRequest, factory resources.Factory) any {
	var params resources.ResourceQueryParams
	err := request.Parse(&params)
	if err != nil {
		return contracts.Fields{"err_msg": err.Error()}
	}
	resource := factory.Get(request.Param("name"))
	if resource == nil {
		return contracts.Fields{"err_msg": "该资源不存在"}
	}

	list, total := resource.Query(params)

	return contracts.Fields{
		"data":  list,
		"total": total,
	}
}

func GetResourceMeta(request contracts.HttpRequest, factory resources.Factory) any {
	resource := factory.Get(request.Param("name"))
	if resource == nil {
		return contracts.Fields{"err_msg": "该资源不存在"}
	}
	meta, err := resource.GetMeta()
	if err != nil {
		return contracts.Fields{"err_msg": err.Error()}
	}
	return contracts.Fields{
		"data": meta,
	}
}