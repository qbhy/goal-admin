package resources

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/goal-web/contracts"
	"github.com/goal-web/supports/commands"
)

func ScanResources(app contracts.Application) contracts.Command {
	return &scanResources{
		Command:   commands.Base("scan:resources", "扫描资源"),
		resources: app.Get("resources").(Factory),
		app:       app,
	}
}

type scanResources struct {
	commands.Command
	app       contracts.Application
	resources Factory
}

func fieldTypeToGoType(fieldType string) string {
	switch fieldType {
	case "json":
		return "object"
	case "date", "datetime", "timestamp":
		return "dateTime"
	case "boolean":
		return "boolean"
	default:
		return "" // Fallback type
	}
}

func (cmd scanResources) Handle() any {
	list, err := cmd.resources.GetProTablePropsListFromDB()
	if err != nil {
		panic(err)
	}
	for _, props := range list {
		base := Base{
			Name:   props.HeaderTitle,
			RowKey: props.RowKey,
			Title:  props.HeaderTitle,
		}

		var columns []string
		for _, col := range props.Columns {
			columns = append(columns, col.DataIndex)
		}

		askErr := survey.Ask([]*survey.Question{
			{
				Name:   "title",
				Prompt: &survey.Input{Message: fmt.Sprintf("[%s] 请输入标题", props.HeaderTitle), Default: props.HeaderTitle},
			},
			{
				Name:   "hideInTable",
				Prompt: &survey.MultiSelect{Message: fmt.Sprintf("[%s] 需要在列表中隐藏的字段", props.HeaderTitle), Options: columns},
			},
		}, &base)

		if askErr != nil {
			panic(askErr)
		}

		err = cmd.resources.SaveResource(base)
		if err != nil {
			panic(err)
		}
	}
	return nil
}
