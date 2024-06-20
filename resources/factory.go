package resources

import (
	"encoding/json"
	"fmt"
	"github.com/goal-web/contracts"
	"github.com/goal-web/supports/exceptions"
	"sync"
)

type ResourceFactory struct {
	db        contracts.DBConnection
	fs        contracts.FileSystem
	mutex     sync.Mutex
	resources map[string]Resource
}

func NewFactory(connection contracts.DBConnection, fs contracts.FileSystem) Factory {
	return &ResourceFactory{
		mutex:     sync.Mutex{},
		resources: make(map[string]Resource),
		db:        connection,
		fs:        fs,
	}
}

func (factory *ResourceFactory) Extend(resource Resource) {
	factory.mutex.Lock()
	defer factory.mutex.Unlock()
	factory.resources[resource.GetName()] = resource
}

func (factory *ResourceFactory) Get(name string) Resource {
	return factory.resources[name]
}

func (factory *ResourceFactory) GetResources() map[string]Resource {
	return factory.resources
}

func (factory *ResourceFactory) GetProTablePropsListFromFs() ([]ProTableProps, contracts.Exception) {
	var list []ProTableProps

	for _, file := range factory.fs.Files("") {
		var props ProTableProps
		err := json.Unmarshal(file.Read(), &props)
		if err != nil {
			return nil, exceptions.WithError(err)
		}
		list = append(list, props)
	}
	return list, nil
}

func (factory *ResourceFactory) GetProTablePropsListFromDB() ([]ProTableProps, contracts.Exception) {
	var tables []string
	err := factory.db.Select(&tables, "show tables;")
	if err != nil {
		return nil, err
	}

	var list []ProTableProps
	for _, table := range tables {
		pro, dbErr := factory.GetProTablePropsFromDB(table)
		if dbErr != nil {
			return nil, dbErr
		}
		list = append(list, pro)
	}
	return list, nil
}

func (factory *ResourceFactory) GetProTablePropsFromDB(table string) (ProTableProps, contracts.Exception) {
	var columns []ColumnInfo
	var pro ProTableProps
	err := factory.db.Select(&columns, fmt.Sprintf("describe `%s`", table))
	if err != nil {
		return pro, err
	}
	return makeProTableProps(table, columns), nil
}

func makeProTableProps(table string, columns []ColumnInfo) ProTableProps {
	var pro = ProTableProps{
		HeaderTitle: table,
	}
	pro.HeaderTitle = table

	for _, column := range columns {
		if column.Key == "PRI" {
			pro.RowKey = column.Field
		}
		pro.Columns = append(pro.Columns, ProTableColumn{
			Title:        column.Field,
			DataIndex:    column.Field,
			ValueType:    fieldTypeToGoType(column.Type),
			InitialValue: column.Default.String,
		})
	}

	return pro
}
