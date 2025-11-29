package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	adminModel "github.com/qbhy/goal-admin/models/admin"
)

// GenerateColumnsFromStruct 通过反射从结构体生成 ResourceColumn 列表（与 drivers 包逻辑保持一致）
// 参照 scheme.tsx 与 Ant Design Pro BetaSchemaForm 语义：
// - 结构体字段作为一个 valueType="group" 的列，子字段挂载到 Columns
// - 子字段 dataIndex 使用自身字段名（不拼接父路径）以保持可读性与递归能力
// - 数组/切片/映射统一按 JSON 处理（后续可升级为 formList）
// - dataIndex 优先使用 json 标签；无标签则用 snake_case 字段名
// - 外键标签：foreign:"model,key,label" → 生成 ValueType=foreign 并填充 ForeignKey
func GenerateColumnsFromStruct(v any) []adminModel.ResourceColumn {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	var columns []adminModel.ResourceColumn

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" { // 未导出字段跳过
			continue
		}

		name := jsonName(&f)
		if name == "" {
			name = camelToSnake(f.Name)
		}
		di := name

		vt, fk := valueTypeForField(&f)

		ft := f.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		if ft.Kind() == reflect.Struct {
			// 结构体以 group 呈现，递归构建子列
			groupCols := buildGroupColumns("", ft)
			// 标题优先级：独立 title 标签 > column:title > json 名称(di) > 字段名
			displayTitle := f.Name
			if t := f.Tag.Get("title"); t != "" {
				displayTitle = t
			} else if colTag := splitKVs(f.Tag.Get("column")); colTag["title"] != "" {
				displayTitle = colTag["title"]
			} else if di != "" {
				displayTitle = di
			}
			col := adminModel.ResourceColumn{
				Title:     displayTitle,
				DataIndex: di,
				ValueType: "group",
				Columns:   groupCols,
			}
			applyColumnTag(&col, &f, false) // group 类型不允许覆盖 valueType
			columns = append(columns, col)
			continue
		}

		if ft.Kind() == reflect.Slice || ft.Kind() == reflect.Array {
			// 列表类型，如果元素是结构体则作为 formList 呈现，且 items 用一个 group 包裹具体字段
			elem := ft.Elem()
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			if elem.Kind() == reflect.Struct {
				itemCols := buildGroupColumns("", elem)
				// 标题优先级：独立 title 标签 > column:title > json 名称(di) > 字段名
				displayTitle := f.Name
				if t := f.Tag.Get("title"); t != "" {
					displayTitle = t
				} else if colTag := splitKVs(f.Tag.Get("column")); colTag["title"] != "" {
					displayTitle = colTag["title"]
				} else if di != "" {
					displayTitle = di
				}
				col := adminModel.ResourceColumn{
					Title:     displayTitle,
					DataIndex: di,
					ValueType: "formList",
					Columns: []adminModel.ResourceColumn{
						{
							ValueType: "group",
							Columns:   itemCols,
						},
					},
				}
				applyColumnTag(&col, &f, false) // 顶层 formList 不允许覆盖 valueType
				columns = append(columns, col)
				continue
			}
			vt = "json"
		} else if ft.Kind() == reflect.Map {
			vt = "json"
		}

		// 标题优先级：独立 title 标签 > column:title > json 名称(di) > 字段名
		displayTitle := f.Name
		if t := f.Tag.Get("title"); t != "" {
			displayTitle = t
		} else if colTag := splitKVs(f.Tag.Get("column")); colTag["title"] != "" {
			displayTitle = colTag["title"]
		} else if di != "" {
			displayTitle = di
		}
		col := adminModel.ResourceColumn{
			Title:      displayTitle,
			DataIndex:  di,
			ValueType:  vt,
			ForeignKey: fk,
			Sorter: func() bool {
				if di == "id" || strings.HasSuffix(di, "_at") || strings.Contains(di, "time") {
					return true
				}
				return false
			}(),
			HideInSearch: di == "updated_at",
			HideInForm:   di == "id" || di == "created_at" || di == "updated_at",
		}
		applyColumnTag(&col, &f, true) // 允许通过 tag 覆盖常规字段与 valueType
		columns = append(columns, col)
	}

	return columns
}

// buildGroupColumns 递归构建结构体的子列，子列 dataIndex 仅使用自身字段名（不拼接父路径）
func buildGroupColumns(_ string, t reflect.Type) []adminModel.ResourceColumn {
	var cols []adminModel.ResourceColumn

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			continue
		}

		child := jsonName(&sf)
		if child == "" {
			child = camelToSnake(sf.Name)
		}
		childDI := child

		// 标题优先级：独立 title 标签 > column:title > json 名称(childDI) > 字段名
		displayTitle := sf.Name
		if tt := sf.Tag.Get("title"); tt != "" {
			displayTitle = tt
		} else if colTag := splitKVs(sf.Tag.Get("column")); colTag["title"] != "" {
			displayTitle = colTag["title"]
		} else if childDI != "" {
			displayTitle = childDI
		}

		svt, sfk := valueTypeForField(&sf)

		st := sf.Type
		if st.Kind() == reflect.Ptr {
			st = st.Elem()
		}

		switch st.Kind() {
		case reflect.Struct:
			// 继续递归构建子组
			sub := buildGroupColumns("", st)
			col := adminModel.ResourceColumn{
				Title:     displayTitle,
				DataIndex: childDI,
				ValueType: "group",
				Columns:   sub,
			}
			applyColumnTag(&col, &sf, false)
			cols = append(cols, col)
		case reflect.Slice, reflect.Array, reflect.Map:
			col := adminModel.ResourceColumn{
				Title:        displayTitle,
				DataIndex:    childDI,
				ValueType:    "json",
				ForeignKey:   sfk,
				Sorter:       false,
				HideInSearch: strings.HasSuffix(childDI, "_at") || strings.Contains(childDI, "time"),
				HideInForm:   childDI == "id" || childDI == "created_at" || childDI == "updated_at",
			}
			applyColumnTag(&col, &sf, true)
			cols = append(cols, col)
		default:
			col := adminModel.ResourceColumn{
				Title:      displayTitle,
				DataIndex:  childDI,
				ValueType:  svt,
				ForeignKey: sfk,
				Sorter: func() bool {
					if strings.HasSuffix(childDI, "_at") || strings.Contains(childDI, "time") {
						return true
					}
					return false
				}(),
				HideInSearch: strings.HasSuffix(childDI, "_at") || strings.Contains(childDI, "time"),
				HideInForm:   childDI == "id" || childDI == "created_at" || childDI == "updated_at",
			}
			applyColumnTag(&col, &sf, true)
			cols = append(cols, col)
		}
	}

	return cols
}

// applyColumnTag 从 struct 字段的直连标签中提取列属性并应用到列对象
// 约定：模型字段的 tag 与 ResourceColumn 的字段一一对应（小驼峰）
// 支持的键：title、valueType、width、hideInSearch、hideInTable、hideInForm、sorter、dataIndex、valueEnum（或 enum）、foreignKey
func applyColumnTag(col *adminModel.ResourceColumn, f *reflect.StructField, allowOverrideValueType bool) {
	// 优先使用直连标签（不会使用聚合的 column 标签）
	if v := f.Tag.Get("title"); v != "" {
		col.Title = v
	}

	if v := f.Tag.Get("valueType"); v != "" && allowOverrideValueType {
		col.ValueType = v
	}

	if v := f.Tag.Get("defaultValue"); v != "" {
		col.InitialValue = v
	}
	if v := f.Tag.Get("default"); v != "" {
		col.InitialValue = v
	}

	if v := f.Tag.Get("width"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil {
			col.Width = int32(iv)
		}
	}

	if v := f.Tag.Get("hideInSearch"); v != "" {
		if bv, err := strconv.ParseBool(v); err == nil {
			col.HideInSearch = bv
		}
	}

	if v := f.Tag.Get("multiple"); v != "" {
		if bv, err := strconv.ParseBool(v); err == nil {
			col.Multiple = bv
		}
	}

	if v := f.Tag.Get("hideInTable"); v != "" {
		if bv, err := strconv.ParseBool(v); err == nil {
			col.HideInTable = bv
		}
	}

	if v := f.Tag.Get("hideInForm"); v != "" {
		if bv, err := strconv.ParseBool(v); err == nil {
			col.HideInForm = bv
		}
	}

	if v := f.Tag.Get("sorter"); v != "" {
		if bv, err := strconv.ParseBool(v); err == nil {
			col.Sorter = bv
		}
	}

	if v := f.Tag.Get("dataIndex"); v != "" {
		col.DataIndex = v
	}

	if v := f.Tag.Get("valueEnum"); v != "" {
		col.ValueEnum = parseValueEnum(v)
	} else if v := f.Tag.Get("enum"); v != "" { // 兼容 enum 命名
		col.ValueEnum = parseValueEnum(v)
	}

	// 允许以 foreignKey:"model,key,label" 直接指定外键，并在允许时覆盖为 foreign 类型
	if v := f.Tag.Get("foreignKey"); v != "" {
		parts := strings.Split(v, ",")
		if len(parts) >= 3 {
			col.ForeignKey = &adminModel.ForeignKey{
				Model:      strings.TrimSpace(parts[0]),
				KeyField:   strings.TrimSpace(parts[1]),
				LabelField: strings.TrimSpace(parts[2]),
				Multiple:   col.Multiple,
			}
			if allowOverrideValueType {
				col.ValueType = "foreign"
			}
		}
	}

}

// splitKVs 支持用逗号或分号分隔的 key=value 列表
func splitKVs(s string) map[string]string {
	res := make(map[string]string)
	var parts []string
	if strings.Contains(s, ",") {
		parts = strings.Split(s, ",")
	} else if strings.Contains(s, ";") {
		parts = strings.Split(s, ";")
	} else {
		parts = []string{s}
	}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var key, val string
		if idx := strings.Index(p, "="); idx >= 0 {
			key = strings.TrimSpace(p[:idx])
			val = strings.TrimSpace(p[idx+1:])
		} else if idx := strings.Index(p, ":"); idx >= 0 {
			key = strings.TrimSpace(p[:idx])
			val = strings.TrimSpace(p[idx+1:])
		} else {
			key = p
			val = ""
		}
		if key != "" {
			res[key] = val
		}
	}
	return res
}

// parseValueEnum 将 "key:text;key:text" 或 "key=text,key=text" 解析为 valueEnum（兼容中英文分隔符）
func parseValueEnum(s string) map[string]any {
	normalized := strings.TrimSpace(s)
	normalized = strings.ReplaceAll(normalized, "；", ";")
	normalized = strings.ReplaceAll(normalized, "，", ",")
	normalized = strings.ReplaceAll(normalized, "：", ":")

	m := make(map[string]any)
	var entries []string
	if strings.Contains(normalized, ";") {
		entries = strings.Split(normalized, ";")
	} else {
		entries = strings.Split(normalized, ",")
	}
	for _, e := range entries {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		var k, t string
		if idx := strings.Index(e, ":"); idx >= 0 {
			k = strings.TrimSpace(e[:idx])
			t = strings.TrimSpace(e[idx+1:])
		} else if idx := strings.Index(e, "="); idx >= 0 {
			k = strings.TrimSpace(e[:idx])
			t = strings.TrimSpace(e[idx+1:])
		} else {
			k = e
			t = e
		}
		if k != "" {
			m[k] = t
		}
	}
	return m
}

// valueTypeForField 依据字段与标签解析值类型与外键
func valueTypeForField(f *reflect.StructField) (string, *adminModel.ForeignKey) {
	// foreign 标签优先
	if tag, ok := f.Tag.Lookup("foreign"); ok {
		parts := strings.Split(tag, ",")
		if len(parts) >= 3 {
			var foreignKey string
			if len(parts) > 3 {
				foreignKey = strings.TrimSpace(parts[3])
				if foreignKey == "" {
					foreignKey = fmt.Sprintf("%s_id", strings.ToLower(parts[1]))
				}
			}
			return "foreign", &adminModel.ForeignKey{
				Model:      strings.TrimSpace(parts[0]),
				KeyField:   strings.TrimSpace(parts[1]),
				LabelField: strings.TrimSpace(parts[2]),
				ForeignKey: foreignKey,
			}
		}
		return "json", nil
	}
	return kindValueType(f.Type.Kind(), f.Name)
}

// kindValueType 根据 Kind 与名称推断 valueType（必要时做少量启发式）
func kindValueType(k reflect.Kind, name string) (string, *adminModel.ForeignKey) {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "digit", nil
	case reflect.Bool:
		return "switch", nil
	case reflect.String:
		if strings.HasSuffix(name, "At") || strings.HasSuffix(name, "_at") || strings.Contains(strings.ToLower(name), "time") {
			return "dateTime", nil
		}
		return "text", nil
	case reflect.Struct:
		return "json", nil
	case reflect.Slice, reflect.Array, reflect.Map:
		return "json", nil
	default:
		return "text", nil
	}
}

// jsonName 提取 json 标签中的字段名
func jsonName(f *reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		return ""
	}
	if idx := strings.Index(tag, ","); idx >= 0 {
		return strings.TrimSpace(tag[:idx])
	}
	return strings.TrimSpace(tag)
}

// camelToSnake 简单的驼峰转下划线
func camelToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
