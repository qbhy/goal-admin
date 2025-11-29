package admin

import (
	"fmt"
	"reflect"

	"github.com/goal-web/auth"
	goaladmin "github.com/qbhy/goal-admin"
	admin2 "github.com/qbhy/goal-admin/enums/admin"
	"github.com/qbhy/goal-admin/models"
	adminReq "github.com/qbhy/goal-admin/requests/admin"
	adminRes "github.com/qbhy/goal-admin/results/admin"

	"github.com/goal-web/contracts"
	"github.com/spf13/cast"
)

// helper: 获取资源实例
func getResource(model string) (goaladmin.Resource, error) {
	r, ok := goaladmin.Default().Get(model)
	if !ok {
		return nil, fmt.Errorf("未找到资源模型: %s", model)
	}
	return r, nil
}

// helper: 归一化 fields
func toFields(v any) contracts.Fields {
	if v == nil {
		return contracts.Fields{}
	}
	if f, ok := v.(contracts.Fields); ok {
		return f
	}
	if m, ok := v.(map[string]any); ok {
		return contracts.Fields(m)
	}
	return contracts.Fields{}
}

// helper: 提取对象中的 ID 字段
func extractID(v any) int64 {
	if v == nil {
		return 0
	}
	if m, ok := v.(map[string]any); ok {
		if id, ok := m["id"]; ok {
			return cast.ToInt64(id)
		}
	}
	if f, ok := v.(contracts.Fields); ok {
		if id, ok := f["id"]; ok {
			return cast.ToInt64(id)
		}
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.IsValid() && rv.Kind() == reflect.Struct {
		field := rv.FieldByName("Id")
		if field.IsValid() {
			return cast.ToInt64(field.Interface())
		}
	}
	return cast.ToInt64(v)
}

func Admin(ctx contracts.Context) *models.AdminModel {
	return auth.GetGuard("admin_jwt", ctx).User().(*models.AdminModel)
}

func init() {
	// 绑定通用资源服务方法
	ResourceServiceDefine.GetResourceMeta = func(req *adminReq.GetResourceMetaReq, ctx contracts.Context) (*adminRes.GetResourceMetaResult, error) {
		r, err := getResource(req.Model)
		if err != nil {
			return nil, err
		}
		if admin := Admin(ctx); !r.Can(admin, admin2.ActionTypeshow.String()) {
			return nil, fmt.Errorf("无权访问资源: %s", req.Model)
		}

		meta := r.Meta()
		return &adminRes.GetResourceMetaResult{Meta: meta}, nil
	}

	ResourceServiceDefine.ListResource = func(req *adminReq.ListResourceReq, ctx contracts.Context) (*adminRes.ListResourceResult, error) {
		r, err := getResource(req.Model)
		if err != nil {
			return nil, err
		}

		if admin := Admin(ctx); !r.Can(admin, admin2.ActionTypeshow.String()) {
			return nil, fmt.Errorf("无权访问资源: %s", req.Model)
		}

		list, total := r.List(req.Query)
		var results []any
		for _, item := range list {
			if fieldsProvider, ok := item.(contracts.FieldsProvider); ok {
				item = fieldsProvider.ToFields()
				results = append(results, item)
			} else {
				results = append(results, item)
			}
		}
		return &adminRes.ListResourceResult{Total: int64(total), List: results}, nil
	}

	ResourceServiceDefine.GetResourceDetail = func(req *adminReq.GetResourceDetailReq, ctx contracts.Context) (*adminRes.GetResourceDetailResult, error) {
		r, err := getResource(req.Model)
		if err != nil {
			return nil, err
		}

		if admin := Admin(ctx); !r.Can(admin, admin2.ActionTypeshow.String()) {
			return nil, fmt.Errorf("无权访问资源: %s", req.Model)
		}

		item, err := r.Find(int(req.Id))
		if err != nil {
			return nil, err
		}
		return &adminRes.GetResourceDetailResult{Item: item}, nil
	}

	ResourceServiceDefine.CreateResource = func(req *adminReq.CreateResourceReq, ctx contracts.Context) (*adminRes.CreateResourceResult, error) {
		r, err := getResource(req.Model)
		if err != nil {
			return nil, err
		}

		if admin := Admin(ctx); !r.Can(admin, admin2.ActionTypecreate.String()) {
			return nil, fmt.Errorf("无权创建资源: %s", req.Model)
		}

		created, err := r.Create(toFields(req.Fields))
		if err != nil {
			return nil, err
		}
		return &adminRes.CreateResourceResult{Id: extractID(created)}, nil
	}

	ResourceServiceDefine.UpdateResource = func(req *adminReq.UpdateResourceReq, ctx contracts.Context) (*adminRes.UpdateResourceResult, error) {
		r, err := getResource(req.Model)
		if err != nil {
			return nil, err
		}

		if admin := Admin(ctx); !r.Can(admin, admin2.ActionTypeedit.String()) {
			return nil, fmt.Errorf("无权编辑资源: %s", req.Model)
		}

		_, err = r.Update(int(req.Id), toFields(req.Fields))
		if err != nil {
			return &adminRes.UpdateResourceResult{Success: false, Message: err.Error()}, nil
		}
		return &adminRes.UpdateResourceResult{Success: true, Message: "ok"}, nil
	}

	ResourceServiceDefine.DeleteResource = func(req *adminReq.DeleteResourceReq, ctx contracts.Context) (*adminRes.DeleteResourceResult, error) {
		r, err := getResource(req.Model)
		if err != nil {
			return nil, err
		}

		if admin := Admin(ctx); !r.Can(admin, admin2.ActionTypedelete.String()) {
			return nil, fmt.Errorf("无权删除资源: %s", req.Model)
		}

		if err = r.Delete(int(req.Id)); err != nil {
			return &adminRes.DeleteResourceResult{Success: false, Message: err.Error()}, nil
		}
		return &adminRes.DeleteResourceResult{Success: true, Message: "ok"}, nil
	}

	ResourceServiceDefine.BatchDeleteResource = func(req *adminReq.BatchDeleteResourceReq, ctx contracts.Context) (*adminRes.BatchDeleteResourceResult, error) {
		r, err := getResource(req.Model)
		if err != nil {
			return nil, err
		}

		if admin := Admin(ctx); !r.Can(admin, admin2.ActionTypebatchDelete.String()) {
			return nil, fmt.Errorf("无权批量删除资源: %s", req.Model)
		}

		ids := make([]int, 0, len(req.Ids))
		for _, id := range req.Ids {
			ids = append(ids, int(id))
		}
		if err = r.BatchDelete(ids); err != nil {
			return &adminRes.BatchDeleteResourceResult{Success: false, Count: int64(len(ids)), Message: err.Error()}, nil
		}
		return &adminRes.BatchDeleteResourceResult{Success: true, Count: int64(len(ids)), Message: "ok"}, nil
	}

	ResourceServiceDefine.ExportResource = func(req *adminReq.ExportResourceReq, ctx contracts.Context) (*adminRes.ExportResourceResult, error) {
		r, err := getResource(req.Model)
		if err != nil {
			return nil, err
		}

		if admin := Admin(ctx); !r.Can(admin, admin2.ActionTypeexport.String()) {
			return nil, fmt.Errorf("无权导出资源: %s", req.Model)
		}

		export, err := r.Export(req.Query)
		if err != nil {
			return nil, err
		}
		return &adminRes.ExportResourceResult{ExportId: export.Id, Status: export.Status, Url: export.Url}, nil
	}

	// 批量获取指定字段，返回 map[id]=>map[field]=>value
	ResourceServiceDefine.BatchFetchFields = func(req *adminReq.BatchFetchFieldsReq, ctx contracts.Context) (*adminRes.BatchFetchFieldsResult, error) {
		r, err := getResource(req.Model)
		if err != nil {
			return nil, err
		}

		if admin := Admin(ctx); !r.Can(admin, admin2.ActionTypeshow.String()) {
			return nil, fmt.Errorf("无权访问资源: %s", req.Model)
		}

		result, err := r.BatchFields(req.KeyField, req.LabelField, req.Keys)
		if err != nil {
			return nil, err
		}
		return &adminRes.BatchFetchFieldsResult{Fields: result}, nil
	}

	// 执行自定义资源操作
	ResourceServiceDefine.ResourceAction = func(req *adminReq.ResourceActionReq, ctx contracts.Context) (*adminRes.ResourceActionResult, error) {
		r, err := getResource(req.Model)
		if err != nil {
			return nil, err
		}

		if admin := Admin(ctx); !r.Can(admin, req.Action) {
			return nil, fmt.Errorf("无权操作资源: %s", req.Model)
		}

		result, err := r.Action(req.Action, req.Payload)
		if err != nil {
			return nil, err
		}
		return &adminRes.ResourceActionResult{Result: result}, nil
	}

}
