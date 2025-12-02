package resources

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/goal-web/filesystem"
	"github.com/qbhy/goal-admin/models"
	"github.com/qbhy/goal-admin/models/admin"
	"github.com/qbhy/goal-admin/utils"

	"github.com/goal-web/contracts"
	"github.com/goal-web/database/table"
	"github.com/goal-web/filesystem/adapters"
	"github.com/goal-web/validation"
)

type BatchFieldsHandler func(keyField, labelField string, ids []string) (map[string]any, error)
type PermissionHandler func(admin *models.AdminModel, action string) bool

type Base struct {
	Title      string
	Name       string
	RowKey     string
	Exportable bool
	Deleteable bool
	Creatable  bool
	Updatable  bool
	Copyable   bool
	Columns    []admin.ResourceColumn

	Query QueryFunction

	BatchFieldsHandler BatchFieldsHandler

	// Handlers for CRUD operations; set by specific resources as needed
	CreateHandler      func(fields contracts.Fields) (any, error)
	UpdateHandler      func(id int, fields contracts.Fields) (any, error)
	DeleteHandler      func(id int) error
	BatchDeleteHandler func(ids []int) error
	FindHandler        func(id int) (any, error)

	// 该资源可以进行的操作
	Actions     map[string]func(payload string) (any, error)
	ActionsMeta []admin.ActionMeta

	PermissionHandler PermissionHandler
}

type QueryFunction func(params admin.QueryParams) ([]any, uint64)

func sortColumnsRec(cols []admin.ResourceColumn) []admin.ResourceColumn {
	for i := range cols {
		if len(cols[i].Columns) > 0 {
			cols[i].Columns = sortColumnsRec(cols[i].Columns)
		}
	}
	type pair struct {
		idx int
		col admin.ResourceColumn
	}
	ps := make([]pair, len(cols))
	for i, c := range cols {
		ps[i] = pair{idx: i, col: c}
	}
	sort.SliceStable(ps, func(i, j int) bool {
		var ki int64
		var kj int64
		if ps[i].col.Order != 0 {
			ki = ps[i].col.Order
		} else {
			ki = int64(ps[i].idx)
		}
		if ps[j].col.Order != 0 {
			kj = ps[j].col.Order
		} else {
			kj = int64(ps[j].idx)
		}
		if ki == kj {
			return ps[i].idx < ps[j].idx
		}
		return ki < kj
	})
	for i := range cols {
		cols[i] = ps[i].col
	}
	return cols
}

func Basic[T any](title string, query func() *table.Table[T]) *Base {
	queryInstance := query()
	var item T
	r := Base{
		Title:      title,
		Name:       queryInstance.GetTableName(),
		RowKey:     queryInstance.GetPrimaryKeyField(),
		Exportable: true,
		Deleteable: true,
		Creatable:  true,
		Updatable:  true,
		Copyable:   true,
		Columns:    utils.GenerateColumnsFromStruct(item),
	}

	r.Query = BuildQuery(query)
	r.FindHandler = func(id int) (any, error) {
		return query().FindOrFail(int64(id)), nil
	}

	r.DeleteHandler = func(id int) error {
		_, err := query().Where(queryInstance.GetPrimaryKeyField(), "=", id).DeleteE()
		return err
	}

	r.BatchDeleteHandler = func(ids []int) error {
		_, err := query().WhereIn(queryInstance.GetPrimaryKeyField(), ids).DeleteE()
		return err
	}

	r.CreateHandler = func(fields contracts.Fields) (any, error) {
		return query().CreateE(fields)
	}

	r.UpdateHandler = func(id int, fields contracts.Fields) (any, error) {
		return query().Where(queryInstance.GetPrimaryKeyField(), "=", id).UpdateE(fields)
	}

	return &r
}

func (b Base) Can(admin *models.AdminModel, action string) bool {
	if b.PermissionHandler != nil {
		return b.PermissionHandler(admin, action)
	}

	return admin != nil
}

func (b Base) Meta() admin.ResourceMeta {
	return admin.ResourceMeta{
		Title:      b.Title,
		Name:       b.Name,
		RowKey:     b.RowKey,
		Exportable: b.Exportable,
		Deleteable: b.Deleteable,
		Creatable:  b.Creatable,
		Updatable:  b.Updatable,
		Copyable:   b.Copyable,
		Columns:    sortColumnsRec(append([]admin.ResourceColumn(nil), b.Columns...)),
		Actions:    b.ActionsMeta,
	}
}

func (b Base) GetName() string {
	return b.Name
}

//	Export 导出
//
// 1. 创建 ExportModel
// 2. 直接组装CSV字符串并上传到OSS
// 3. 获取OSS URL并更新下载链接
func (b Base) Export(params admin.QueryParams) (*models.ExportModel, error) {
	if !b.Exportable {
		return nil, fmt.Errorf("resource '%s' is not exportable", b.Name)
	}
	if b.Query == nil {
		return nil, fmt.Errorf("resource '%s' has no query function for export", b.Name)
	}

	// 创建导出记录，标记为 pending
	export := models.ExportQuery().Create(contracts.Fields{
		"admin_id": int64(0), // 无上下文时设为 0；实际业务可在服务层注入当前管理员ID
		"url":      "",
		"status":   "pending",
		"resource": b.Name,
		"params":   params,
	})

	// 生成文件名
	filename := fmt.Sprintf("%s-%d.csv", b.Name, time.Now().Unix())

	// 使用内存缓冲区组装CSV字符串
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写表头与内容
	var headerTitles []string
	var headers []admin.ResourceColumn
	var headerKeys []string
	var foreignKeys = map[string]*admin.ForeignKey{}

	// 推断列头：优先使用资源 Meta 的 Columns（标题来自 Title，数据键来自 DataIndex）；支持嵌套列展平
	if metaCols := b.Meta().Columns; len(metaCols) > 0 {
		var walk func(cols []admin.ResourceColumn)
		walk = func(cols []admin.ResourceColumn) {
			for _, c := range cols {
				if c.HideInTable {
					continue
				}
				if len(c.Columns) > 0 {
					walk(c.Columns)
					continue
				}
				if c.ForeignKey != nil {
					foreignKeys[c.DataIndex] = c.ForeignKey
				}
				headerTitles = append(headerTitles, c.Title)
				headers = append(headers, c)
				headerKeys = append(headerKeys, c.DataIndex)
			}
		}
		walk(metaCols)
	}

	_ = writer.Write(headerTitles)

	var foreignFields = map[string]contracts.Fields{}

	// 执行查询

	for {
		list, _ := b.Query(params)
		if len(list) == 0 {
			break
		}

		for localKey, foreign := range foreignKeys {
			var foreignIds []any

			for _, item := range list {
				itemFields, ok := item.(contracts.Fields)
				if !ok {
					itemFields = item.(contracts.FieldsProvider).ToFields()
				}
				foreignIds = append(foreignIds, itemFields[localKey])
			}

			table.ArrayQuery(foreign.Model).
				Where(foreign.KeyField, "in", foreignIds).
				Select(foreign.KeyField, foreign.LabelField).
				Get().
				Foreach(func(i int, c *contracts.Fields) {
					if _, ok := foreignFields[localKey]; ok {
						foreignFields[localKey][fmt.Sprintf("%v", (*c)[foreign.KeyField])] = (*c)[foreign.LabelField]
					} else {
						foreignFields[localKey] = map[string]any{
							fmt.Sprintf("%v", (*c)[foreign.KeyField]): (*c)[foreign.LabelField],
						}
					}
				})

		}

		for _, item := range list {
			var row []string
			var itemFields, ok = item.(contracts.Fields)

			if !ok {
				itemFields = item.(contracts.FieldsProvider).ToFields()
			}

			for _, h := range headerKeys {
				if col := foreignFields[h]; col != nil {
					row = append(row, fmt.Sprintf("%v", col[fmt.Sprintf("%v", itemFields[h])]))
				} else {
					row = append(row, fmt.Sprintf("%v", itemFields[h]))
				}
			}
			_ = writer.Write(row)
		}
		params.Page++
	}

	writer.Flush()
	if writer.Error() != nil {
		_ = export.Update(contracts.Fields{"status": "failed"})
		return nil, writer.Error()
	}

	// 获取OSS实例并上传CSV内容
	oss := filesystem.Disk("oss").(*adapters.Oss)

	// 直接上传CSV字符串到OSS
	err := oss.Put(fmt.Sprintf("exports/%s", filename), buf.String())
	if err != nil {
		_ = export.Update(contracts.Fields{"status": "failed"})
		return nil, err
	}

	// 获取OSS文件URL
	url := oss.Url(fmt.Sprintf("exports/%s", filename))
	_ = export.Update(contracts.Fields{
		"status": "success",
		"url":    url,
	})

	return export, nil
}

func (b Base) Create(fields contracts.Fields) (any, error) {
	if !b.Creatable {
		return nil, fmt.Errorf("resource '%s' is not creatable", b.Name)
	}
	if b.CreateHandler != nil {
		return b.CreateHandler(fields)
	}
	return nil, fmt.Errorf("create not implemented for base resource '%s'", b.Name)
}

func (b Base) BatchDelete(ids []int) error {
	if !b.Deleteable {
		return fmt.Errorf("resource '%s' is not deleteable", b.Name)
	}
	if b.BatchDeleteHandler != nil {
		return b.BatchDeleteHandler(ids)
	}
	return fmt.Errorf("BatchDelete not implemented for base resource '%s'", b.Name)
}

func (b Base) Delete(id int) error {
	if !b.Deleteable {
		return fmt.Errorf("resource '%s' is not deleteable", b.Name)
	}
	if b.DeleteHandler != nil {
		return b.DeleteHandler(id)
	}
	return fmt.Errorf("delete not implemented for base resource '%s'", b.Name)
}

func (b Base) Update(id int, fields contracts.Fields) (any, error) {
	if !b.Updatable {
		return nil, fmt.Errorf("resource '%s' is not updatable", b.Name)
	}
	if b.UpdateHandler != nil {
		return b.UpdateHandler(id, fields)
	}
	return nil, fmt.Errorf("update not implemented for base resource '%s'", b.Name)
}

func (b Base) Find(id int) (any, error) {
	if b.FindHandler != nil {
		return b.FindHandler(id)
	}
	return nil, fmt.Errorf("find not implemented for base resource '%s'", b.Name)
}

func (b Base) List(params admin.QueryParams) ([]any, uint64) {
	if b.Query == nil {
		return []any{}, 0
	}
	return b.Query(params)
}

func (b Base) BatchFields(keyField, labelField string, ids []string) (map[string]any, error) {
	if b.BatchFieldsHandler == nil {
		return nil, fmt.Errorf("BatchFields not implemented for base resource '%s'", b.Name)
	}
	return b.BatchFieldsHandler(keyField, labelField, ids)
}

func (b Base) Action(action, payload string) (any, error) {
	if b.Actions == nil {
		return nil, fmt.Errorf("action '%s' not implemented for base resource '%s'", action, b.Name)
	}

	if handler, ok := b.Actions[action]; ok {
		return handler(payload)
	}

	return nil, fmt.Errorf("action '%s' not implemented for base resource '%s'", action, b.Name)
}

// RegisterAction 注册一个自定义操作到资源
func (b *Base) RegisterAction(action string, handler any) {
	// 初始化动作映射
	if b.Actions == nil {
		b.Actions = make(map[string]func(payload string) (any, error))
	}

	hv := reflect.ValueOf(handler)
	ht := hv.Type()

	// 必须是函数
	if hv.Kind() != reflect.Func {
		panic(fmt.Sprintf("handler for action '%s' must be a function", action))
	}

	// 必须仅有一个参数，且为 struct 或 *struct
	if ht.NumIn() != 1 {
		panic(fmt.Sprintf("handler for action '%s' must have exactly 1 parameter", action))
	}
	pt := ht.In(0)
	if !(pt.Kind() == reflect.Struct || (pt.Kind() == reflect.Ptr && pt.Elem().Kind() == reflect.Struct)) {
		panic(fmt.Sprintf("handler for action '%s' parameter must be a struct or pointer to struct", action))
	}

	// 返回值必须是 (any, error)
	if ht.NumOut() != 2 || !ht.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		panic(fmt.Sprintf("handler for action '%s' must return (any, error)", action))
	}

	// 生成 ActionsMeta: 从参数结构体提取字段并映射为 ResourceColumn
	var structType reflect.Type
	if pt.Kind() == reflect.Ptr {
		structType = pt.Elem()
	} else {
		structType = pt
	}

	var columns []admin.ResourceColumn
	var batch bool
	var idFound bool
	var idTypeOk bool
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		// 跳过未导出的字段
		if field.PkgPath != "" { // 非导出字段
			continue
		}

		jsonName := field.Tag.Get("json")
		if jsonName == "-" {
			continue
		}
		// 处理带有逗号选项的 json 标签，如 "id,omitempty"
		if jsonName == "" {
			jsonName = strings.ToLower(field.Name)
		} else if idx := strings.IndexByte(jsonName, ','); idx > 0 {
			jsonName = jsonName[:idx]
		}

		// 校验并记录 id 字段类型是否合法（int64 或 []int64）以及是否为批量
		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if jsonName == "id" {
			idFound = true
			switch ft.Kind() {
			case reflect.Int64, reflect.Int, reflect.Uint64, reflect.Uint32:
				idTypeOk = true
			case reflect.Slice, reflect.Array:
				if ft.Elem().Kind() == reflect.Int64 {
					idTypeOk = true
					batch = true
				}
			}
		}
	}

	// 统一使用 GenerateColumnsFromStruct 生成列，并在 ActionsMeta 中忽略 id 字段
	gen := utils.GenerateColumnsFromStruct(reflect.New(structType).Interface())
	for _, c := range gen {
		if c.DataIndex == "id" {
			continue
		}
		columns = append(columns, c)
	}

	// 必须存在合法的 id 字段（int64 或 []int64）
	if !idFound || !idTypeOk {
		batch = true
		// 暂时支持没有 id 的操作
		//panic(fmt.Sprintf("handler for action '%s' must contain field 'id' of type int64 or []int64", action))
	}

	// 写入或更新动作的元信息
	// 若已有相同 action 名称的元信息，则覆盖其 columns
	var updated bool
	for i := range b.ActionsMeta {
		if b.ActionsMeta[i].Name == action {
			b.ActionsMeta[i].Columns = columns
			b.ActionsMeta[i].Batch = batch
			updated = true
			break
		}
	}
	if !updated {
		b.ActionsMeta = append(b.ActionsMeta, admin.ActionMeta{Name: action, Columns: columns, Batch: batch})
	}

	// 注册包装后的调用器
	b.Actions[action] = func(payload string) (any, error) {
		// 构造参数指针并反序列化 payload
		var paramPtr reflect.Value
		if pt.Kind() == reflect.Ptr {
			paramPtr = reflect.New(pt.Elem())
		} else {
			paramPtr = reflect.New(pt)
		}

		if len(payload) > 0 {
			if err := json.Unmarshal([]byte(payload), paramPtr.Interface()); err != nil {
				return nil, err
			}
		}

		if err := validation.Struct(paramPtr); err != nil {
			return nil, err
		}

		// 组装调用参数（指针或值）
		var arg reflect.Value
		if pt.Kind() == reflect.Ptr {
			arg = paramPtr
		} else {
			arg = paramPtr.Elem()
		}

		// 调用并返回 (any, error)
		results := hv.Call([]reflect.Value{arg})
		res := results[0].Interface()
		if !results[1].IsNil() {
			return res, results[1].Interface().(error)
		}
		return res, nil
	}
}
