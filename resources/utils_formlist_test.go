package resources

import (
	"testing"

	"github.com/qbhy/goal-admin/models"
	adminModel "github.com/qbhy/goal-admin/models/admin"
	"github.com/qbhy/goal-admin/utils"
)

// findColTop 查找顶层列（不递归）
func findColTop(cols []adminModel.ResourceColumn, di string) *adminModel.ResourceColumn {
	for i := range cols {
		if cols[i].DataIndex == di {
			return &cols[i]
		}
	}
	return nil
}

func TestGenerateColumnsFromCompositionRecipeModel(t *testing.T) {
	cols := utils.GenerateColumnsFromStruct(models.ExportModel{})
	if len(cols) == 0 {
		t.Fatalf("no columns generated")
	}

	// 顶层字段基本断言
	if id := findColTop(cols, "id"); id == nil || id.ValueType != "digit" {
		t.Fatalf("id column missing or type not digit: %+v", id)
	}
	if name := findColTop(cols, "name"); name == nil || name.ValueType != "text" {
		t.Fatalf("name column missing or type not text: %+v", name)
	}
	if nm := findColTop(cols, "new_model_id"); nm == nil || nm.ValueType != "foreign" || nm.ForeignKey == nil || nm.ForeignKey.Model != "collectibles" {
		t.Fatalf("new_model_id foreign key parse failed: %+v", nm)
	}

	// formList: rules 为 repeated CompositionRecipeItem，应生成 formList，并包含子列 model_id、quantity
	rules := findColTop(cols, "rules")
	if rules == nil || rules.ValueType != "formList" {
		t.Fatalf("rules column missing or type not formList: %+v", rules)
	}
	if rules != nil {
		// formList 的 items 需要先包一层 group
		var group *adminModel.ResourceColumn
		for i := range rules.Columns {
			if rules.Columns[i].ValueType == "group" {
				group = &rules.Columns[i]
				break
			}
		}
		if group == nil {
			t.Fatalf("rules formList should contain a group item")
		}
		// 子项字段不带父路径
		mi := findColRecursive(group.Columns, "model_id")
		if mi == nil || mi.ValueType != "foreign" || mi.ForeignKey == nil || mi.ForeignKey.Model != "collectibles" {
			t.Fatalf("rules.model_id missing or not foreign: %+v", mi)
		}
		qty := findColRecursive(group.Columns, "quantity")
		if qty == nil || qty.ValueType != "digit" {
			t.Fatalf("rules.quantity missing or not digit: %+v", qty)
		}
	}

	// 时间字段
	if ca := findColTop(cols, "created_at"); ca == nil || ca.ValueType != "dateTime" {
		t.Fatalf("created_at missing or type not dateTime: %+v", ca)
	}
	if ua := findColTop(cols, "updated_at"); ua == nil || ua.ValueType != "dateTime" {
		t.Fatalf("updated_at missing or type not dateTime: %+v", ua)
	}
}
