package resources

import (
	"fmt"
	"github.com/goal-web/application"
	"github.com/goal-web/contracts"
	"github.com/goal-web/database/table"
)

type Filters map[string]Filter
type Filter func(value any, query contracts.QueryBuilder[contracts.Fields])

type Base struct {
	*ProTableProps
	Name          string                       `json:"name"`
	Labels        map[string]string            `json:"labels"`
	ValueEnum     map[string]contracts.Fields  `json:"value_enum"`
	HideInTable   []string                     `json:"hide_in_table"`
	ColumnWrapper func(column *ProTableColumn) `json:"-"`
	Filters       Filters                      `json:"-"`
}

func (base Base) GetName() string {
	return base.Name
}

func (base Base) GetRowKey() string {
	return base.RowKey
}

func (base Base) GetTitle() string {
	return base.HeaderTitle
}

func (base Base) GetMeta() (*ProTableProps, contracts.Exception) {
	if base.Columns == nil {
		dbRes, err := application.Get("resources").(Factory).GetResourceFromDB(base.Name)
		if err != nil {
			return nil, err
		}
		if dbRes != nil {
			base.Columns = dbRes.Columns
		}
	}
	for _, col := range base.Columns {
		if base.Labels != nil {
			if label, exists := base.Labels[col.DataIndex]; exists {
				col.Title = label
			}
		}
		if base.ValueEnum != nil {
			if value, exists := base.ValueEnum[col.DataIndex]; exists {
				col.ValueEnum = value
			}
		}
		for _, field := range base.HideInTable {
			if col.DataIndex == field {
				col.HideInTable = true
			}
		}
	}
	if base.ColumnWrapper != nil {
		for _, col := range base.Columns {
			base.ColumnWrapper(col)
		}
	}

	return base.ProTableProps, nil
}

func (base Base) Delete(id int) contracts.Exception {
	_, err := table.ArrayQuery(base.Name).Where(base.RowKey, id).DeleteE()
	return err
}

func (base Base) Update(id int, fields contracts.Fields) contracts.Exception {
	_, err := table.ArrayQuery(base.Name).Where(base.RowKey, id).UpdateE(fields)
	return err
}

func (base Base) Values(value, label string) contracts.Collection[*contracts.Fields] {
	return table.ArrayQuery(base.Name).
		GroupBy(value, label).
		Select(fmt.Sprintf("%s as value", value), fmt.Sprintf("%s as label", label)).
		Get()
}

func (base Base) Query(params ResourceQueryParams) (contracts.Collection[*contracts.Fields], int64) {
	query := table.ArrayQuery(base.Name).When(params.Filter != nil, func(q contracts.Query[contracts.Fields]) contracts.Query[contracts.Fields] {
		for field, condition := range params.Filter {
			if filter, exists := base.Filters[field]; exists {
				filter(condition, q)
			} else {
				switch c := condition.(type) {
				case string:
					q.Where(field, c)
				case map[string]any:
					if c["condition"] == "like" {
						c["value"] = "%" + c["value"].(string) + "%"
					}
					q.Where(field, c["condition"], c["value"])
				}
			}
		}
		return q
	})

	for field, sort := range params.Sort {
		if sort == "descend" {
			sort = contracts.Desc
		} else {
			sort = contracts.Asc
		}
		query.OrderBy(field, sort)
	}

	return query.Paginate(params.PageSize, params.Current)
}
