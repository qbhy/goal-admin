package resources

import (
	"encoding/json"
	"fmt"
	"github.com/goal-web/contracts"
	"github.com/goal-web/supports/commands"
	"strings"
)

func ScanResources(app contracts.Application) contracts.Command {
	return &scanResources{
		Command:    commands.Base("scan:resources", "扫描资源"),
		connection: app.Get("db").(contracts.DBConnection),
		resources:  app.Get("resources").(Factory),
		app:        app,
		fs:         app.Get("filesystem").(contracts.FileSystemFactory).Disk("resources"),
	}
}

type scanResources struct {
	commands.Command
	connection contracts.DBConnection
	app        contracts.Application
	fs         contracts.FileSystem
	resources  Factory
}

func fieldTypeToGoType(fieldType string) string {
	switch fieldType {
	case "int", "bigint", "float", "double":
		return "number"
	case "json":
		return "object"
	case "binary":
		return "string"
	case "date", "datetime", "timestamp":
		return "dateTime"
	case "boolean":
		return "boolean"
	default:
		if strings.HasPrefix(fieldType, "varchar") ||
			strings.HasPrefix(fieldType, "nvarchar") ||
			strings.HasPrefix(fieldType, "text") {
			return "string"
		}
		return "any" // Fallback type
	}
}

func (cmd scanResources) Handle() any {
	//list1, err1 := cmd.resources.GetProTablePropsListFromFs()
	//fmt.Println(list1, err1)
	list, err := cmd.resources.GetProTablePropsListFromDB()
	if err != nil {
		panic(err)
	}
	for _, props := range list {
		propsBytes, jsonErr := json.Marshal(props)
		if jsonErr != nil {
			panic(jsonErr)
		}
		fsErr := cmd.fs.Put(fmt.Sprintf("%s.json", props.HeaderTitle), string(propsBytes))
		if fsErr != nil {
			panic(fsErr)
		}
	}
	return nil
}
