package resources

import (
	"fmt"

	"github.com/qbhy/goal-admin/models"
	"github.com/qbhy/goal-admin/utils"

	"github.com/goal-web/contracts"
)

// ExportRecordsResource 导出记录资源（只读）
type ExportRecordsResource struct{ Base }

// exportsFind 获取导出记录详情
func exportsFind(id int) (any, error) {
	item := models.ExportQuery().FindOrFail(int64(id))
	return item, nil
}

func NewExportResource() *ExportRecordsResource {
	r := ExportRecordsResource{Base: Base{
		Title:      "导出记录",
		Name:       "exports",
		RowKey:     "id",
		Exportable: false,
		Deleteable: false,
		Creatable:  false,
		Updatable:  false,
		Copyable:   false,
		Columns:    utils.GenerateColumnsFromStruct(models.ExportModel{}),
	}}

	r.Query = BuildQuery(models.ExportQuery)
	r.FindHandler = exportsFind

	r.BatchFieldsHandler = func(keyField, labelField string, ids []string) (map[string]any, error) {
		var fields = map[string]any{}
		models.ExportQuery().When(len(ids) > 0, func(q contracts.QueryBuilder[models.ExportModel]) contracts.QueryBuilder[models.ExportModel] {
			return q.WhereIn(keyField, ids)
		}).Select(keyField, labelField).Get().Foreach(func(i int, m *models.ExportModel) {
			fields[fmt.Sprintf("%v", m.Get(keyField))] = m.Get(labelField)
		})
		return fields, nil
	}
	// 只读资源：不设置 Create/Update/Delete 处理器

	return &r
}
