package http

import (
	"encoding/json"
	"fmt"
	"github.com/goal-web/contracts"
	"github.com/qbhy/goal-admin/resources"
)

func GetResourcesList(factory resources.Factory) any {
	data, exception := factory.GetProTablePropsListFromDB()
	if exception != nil {
		return contracts.Fields{"err_msg": exception.Error()}
	}

	return contracts.Fields{
		"data": data,
	}
}

func GetResourceList(request contracts.HttpRequest, factory resources.Factory) any {
	var params resources.ResourceQueryParams
	err := json.Unmarshal([]byte(fmt.Sprintf(`{"current": %d, "pageSize": %d, "sort": %s, "filter": %s}`,
		request.IntOptional("current", 1),
		request.IntOptional("pageSize", 1),
		request.StringOptional("sort", "{}"),
		request.StringOptional("filter", "{}"),
	)), &params)
	if err != nil {
		return contracts.Fields{"err_msg": err.Error()}
	}
	resource, exception := factory.GetResource(request.Param("name"))
	if exception != nil {
		return contracts.Fields{"err_msg": exception.Error()}
	}

	list, total := resource.Query(params)

	return contracts.Fields{
		"data":  list.ToArray(),
		"total": total,
	}
}

func GetResourceMeta(request contracts.HttpRequest, factory resources.Factory) any {
	resource, exception := factory.GetResource(request.Param("name"))
	if exception != nil {
		return contracts.Fields{"err_msg": exception.Error()}
	}
	meta, err := resource.GetMeta()
	if err != nil {
		return contracts.Fields{"err_msg": err.Error()}
	}
	return contracts.Fields{
		"data": meta,
	}
}
