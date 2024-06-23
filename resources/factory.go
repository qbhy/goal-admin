package resources

import (
	"encoding/json"
	"fmt"
	"github.com/goal-web/contracts"
	"github.com/goal-web/supports/exceptions"
	"github.com/goal-web/supports/logs"
	"sync"
)

type ResourceFactory struct {
	db        contracts.DBConnection
	fs        contracts.FileSystem
	mutex     sync.Mutex
	resources map[string]Resource
}

func NewFactory(connection contracts.DBConnection, fs contracts.FileSystem) Factory {
	factory := &ResourceFactory{
		mutex:     sync.Mutex{},
		resources: make(map[string]Resource),
		db:        connection,
		fs:        fs,
	}

	for _, file := range fs.Files("resources") {
		var resource Base
		if err := json.Unmarshal(file.Read(), &resource); err != nil {
			logs.Default().WithField("file", file.Name()).WithError(err).Error("Failed to unmarshal resource")
		} else {
			factory.resources[resource.GetName()] = resource
		}
	}

	return factory
}

func (factory *ResourceFactory) ExtendResource(resource Resource) {
	factory.mutex.Lock()
	defer factory.mutex.Unlock()
	factory.resources[resource.GetName()] = resource
}

func (factory *ResourceFactory) GetResource(name string) (Resource, contracts.Exception) {
	resource, exists := factory.resources[name]
	if exists {
		return resource, nil
	}
	return nil, exceptions.New(fmt.Sprintf("resource [%s] is not exists", name))
}

func (factory *ResourceFactory) GetProTablePropsListFromFs() ([]*ProTableProps, contracts.Exception) {
	var list []*ProTableProps

	for _, file := range factory.fs.Files("") {
		var props ProTableProps
		err := json.Unmarshal(file.Read(), &props)
		if err != nil {
			return nil, exceptions.WithError(err)
		}
		list = append(list, &props)
	}
	return list, nil
}

func (factory *ResourceFactory) SaveMenuList(list []MenuDataItem) contracts.Exception {
	listBytes, _ := json.Marshal(list)
	err := factory.fs.Put("menu/list.json", string(listBytes))
	if err != nil {
		logs.Default().WithError(err).Error("Failed to save menu list")
	}
	return exceptions.WithError(err)
}

func (factory *ResourceFactory) GetMenuList() []MenuDataItem {
	listStr, err := factory.fs.Get("menu/list.json")
	var menus []MenuDataItem
	if err == nil {
		err = json.Unmarshal([]byte(listStr), &menus)
		if err != nil {
			logs.Default().WithError(err).Error("Failed to parse menu list from menu/list.json")
		} else {
			return menus
		}
	}

	for name, res := range factory.resources {
		menus = append(menus, MenuDataItem{
			Name: res.GetTitle(),
			Path: fmt.Sprintf(`/resource/%s/list`, name),
		})
	}
	return menus
}

func (factory *ResourceFactory) GetProTablePropsListFromDB() ([]*ProTableProps, contracts.Exception) {
	var tables []string
	err := factory.db.Select(&tables, "show tables;")
	if err != nil {
		return nil, err
	}

	var list []*ProTableProps
	for _, table := range tables {
		pro, dbErr := factory.GetProTablePropsFromDB(table)
		if dbErr != nil {
			return nil, dbErr
		}
		list = append(list, pro)
	}
	return list, nil
}

func (factory *ResourceFactory) SaveResource(resource Resource) contracts.Exception {
	jsonBytes, err := json.Marshal(resource)
	if err != nil {
		logs.Default().WithError(err).Error("Failed to encode json")
		return exceptions.WithError(err)
	}
	err = factory.fs.Put(fmt.Sprintf("resources/%s.json", resource.GetName()), string(jsonBytes))
	if err != nil {
		logs.Default().WithError(err).Error("Failed to save resource")
	}
	return exceptions.WithError(err)
}

func (factory *ResourceFactory) GetProTablePropsFromDB(table string) (*ProTableProps, contracts.Exception) {
	var columns []ColumnInfo
	err := factory.db.Select(&columns, fmt.Sprintf("describe `%s`", table))
	if err != nil {
		return nil, err
	}
	return makeProTableProps(table, columns), nil
}

func makeProTableProps(table string, columns []ColumnInfo) *ProTableProps {
	var pro = ProTableProps{
		HeaderTitle: table,
	}

	for _, column := range columns {
		if column.Key == "PRI" {
			pro.RowKey = column.Field
		}
		pro.Columns = append(pro.Columns, &ProTableColumn{
			Title:        column.Field,
			DataIndex:    column.Field,
			ValueType:    fieldTypeToGoType(column.Type),
			InitialValue: column.Default.String,
		})
	}

	return &pro
}
