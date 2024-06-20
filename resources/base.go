package resources

import (
	"github.com/goal-web/application"
	"github.com/goal-web/contracts"
	"github.com/goal-web/database/table"
)

type Base struct {
	Name   string
	RowKey string
	Title  string
}

func (base Base) GetName() string {
	return base.Name
}

func (base Base) GetRowKey() string {
	return base.RowKey
}

func (base Base) GetTitle() string {
	return base.Title
}

func (base Base) GetMeta() (ProTableProps, contracts.Exception) {
	props, err := application.Get("resources").(Factory).GetProTablePropsFromDB(base.Name)
	return props, err
}

func (base Base) Delete(id int) contracts.Exception {
	_, err := table.ArrayQuery(base.Name).Where(base.RowKey, id).DeleteE()
	return err
}

func (base Base) Update(id int, fields contracts.Fields) contracts.Exception {
	_, err := table.ArrayQuery(base.Name).Where(base.RowKey, id).UpdateE(fields)
	return err
}

func (base Base) Query(params ResourceQueryParams) (contracts.Collection[*contracts.Fields], int64) {
	query := table.ArrayQuery(base.Name).When(params.Filter != nil, func(q contracts.Query[contracts.Fields]) contracts.Query[contracts.Fields] {
		for field, condition := range params.Filter {
			q.Where(field, condition.Condition, condition.Value)
		}
		return q
	})

	for field, sort := range params.Sort {
		query.OrderBy(field, sort)
	}

	return query.Paginate(params.PageSize, params.Current)
}
