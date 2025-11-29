package resources

import (
	"testing"

	"github.com/qbhy/goal-admin/models"
	adminModel "github.com/qbhy/goal-admin/models/admin"
	"github.com/qbhy/goal-admin/utils"
)

// findCol 根据 dataIndex 查找列
func findCol(cols []adminModel.ResourceColumn, di string) *adminModel.ResourceColumn {
	for i := range cols {
		if cols[i].DataIndex == di {
			return &cols[i]
		}
	}
	return nil
}

func TestHasFields(t *testing.T) {
	params := adminModel.QueryParams{
		Params: []adminModel.QueryParam{
			{Key: "status"},
			{Key: "id"},
			{Key: "name"},
		},
	}

	if !HasFields(params, "status", "id") {
		t.Errorf("expected HasFields to return true when all keys present")
	}

	if HasFields(params, "status", "missing") {
		t.Errorf("expected HasFields to return false when a key is missing")
	}

	if !HasFields(params) {
		t.Errorf("expected HasFields to return true when no fields provided")
	}
}

// findColRecursive 递归查找列（支持 group 下的子列）
func findColRecursive(cols []adminModel.ResourceColumn, di string) *adminModel.ResourceColumn {
	for i := range cols {
		if cols[i].DataIndex == di {
			return &cols[i]
		}
		if len(cols[i].Columns) > 0 {
			if c := findColRecursive(cols[i].Columns, di); c != nil {
				return c
			}
		}
	}
	return nil
}

func TestGenerateColumnsFromExportModel(t *testing.T) {
	cols := utils.GenerateColumnsFromStruct(models.ExportModel{})
	if len(cols) == 0 {
		t.Fatalf("no columns generated")
	}

	// 打印观察生成结果
	for _, c := range cols {
		t.Logf("%s => dataIndex=%s, type=%s", c.Title, c.DataIndex, c.ValueType)
	}

	// 基本字段
	if id := findCol(cols, "id"); id == nil || id.ValueType != "digit" {
		t.Fatalf("id column missing or type not digit: %+v", id)
	}

	// 外键解析（来自 ExportModel.AdminId 的 foreign 标签）
	if fk := findCol(cols, "admin_id"); fk == nil || fk.ValueType != "foreign" || fk.ForeignKey == nil || fk.ForeignKey.Model != "admins" {
		t.Fatalf("admin_id foreign key parse failed: %+v", fk)
	}

	// 结构体字段 QueryParams：应生成父级 group，并在 Columns 中包含子列（子列 dataIndex 不加父路径）
	if pg := findCol(cols, "params"); pg == nil || pg.ValueType != "group" {
		t.Fatalf("params group missing or type not group: %+v", pg)
	}
	if pp := findColRecursive(cols, "page"); pp == nil || pp.ValueType != "digit" {
		t.Fatalf("params.page missing or type not digit: %+v", pp)
	}
	if ps := findColRecursive(cols, "sorters"); ps == nil || ps.ValueType != "json" {
		t.Fatalf("params.sorters missing or type not json: %+v", ps)
	}

	// 时间字符串字段
	if ca := findCol(cols, "created_at"); ca == nil || ca.ValueType != "dateTime" {
		t.Fatalf("created_at missing or type not dateTime: %+v", ca)
	}
	if ua := findCol(cols, "updated_at"); ua == nil || ua.ValueType != "dateTime" {
		t.Fatalf("updated_at missing or type not dateTime: %+v", ua)
	}
}
