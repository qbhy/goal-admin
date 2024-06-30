package http

import (
	"encoding/json"
	"fmt"
	"github.com/goal-web/contracts"
	"github.com/goal-web/database/table"
	"github.com/goal-web/supports/utils"
	"github.com/qbhy/goal-admin/resources"
)

func GetResourcesList(factory resources.Factory) any {
	data, exception := factory.GetResourceListFromDB()
	if exception != nil {
		return contracts.Fields{"err_msg": exception.Error()}
	}

	return contracts.Fields{
		"data": data,
	}
}

func SaveResource(factory resources.Factory, request contracts.HttpRequest) any {
	var base resources.Base
	err := request.Parse(&base)
	if err != nil {
		return contracts.Fields{"err_msg": err.Error()}
	}

	exception := factory.SaveResource(base)
	if exception != nil {
		return contracts.Fields{"err_msg": exception.Error()}
	}

	return contracts.Fields{"data": nil}
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

	with := map[string][]*contracts.Fields{}

	meta, err := resource.GetMeta()
	if err != nil {
		return contracts.Fields{"err_msg": err.Error()}
	}

	for _, column := range meta.Columns {
		if column.ValueType == "database" {
			var relations []any

			value := utils.GetStringField(column.ValueTypeParams, "value")
			label := utils.GetStringField(column.ValueTypeParams, "label")
			list.Foreach(func(i int, fields *contracts.Fields) {
				var itemRelations []any
				switch v := (*fields)[column.DataIndex].(type) {
				case int64:
					itemRelations = append(itemRelations, v)
				case string:
					if err := json.Unmarshal([]byte(v), &itemRelations); err != nil {
						itemRelations = append(itemRelations, (*fields)[column.DataIndex])
					}
				}
				relations = append(relations, itemRelations...)
			})

			with[column.DataIndex] = table.ArrayQuery(utils.GetStringField(column.ValueTypeParams, "table")).
				WhereIn(value, resources.RemoveDuplicates(relations)).
				Select(value, label).
				Get().Each(func(i int, c *contracts.Fields) *contracts.Fields {
				return &contracts.Fields{
					"value": (*c)[value],
					"label": (*c)[label],
				}
			}).ToArray()
		}
	}

	return contracts.Fields{
		"data":  list.ToArray(),
		"total": total,
		"with":  with,
	}
}

func GetResourceValueEnums(request contracts.HttpRequest, factory resources.Factory) any {
	resource, exception := factory.GetResource(request.Param("name"))
	if exception != nil {
		return contracts.Fields{"err_msg": exception.Error()}
	}
	return contracts.Fields{
		"data": resource.Values(request.QueryParam("value"), request.QueryParam("label")).ToArray(),
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
