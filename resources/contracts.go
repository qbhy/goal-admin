package resources

import (
	"database/sql"
	"github.com/goal-web/contracts"
)

// ProTableColumn 定义了一个与 Pro Table 列对应的 Go 结构体
type ProTableColumn struct {
	ValueTypeParams    map[string]any `json:"valueTypeParams,omitempty"`
	Title              string         `json:"title,omitempty"`     // 列的标题
	DataIndex          string         `json:"dataIndex,omitempty"` // 数据索引
	ValueType          string         `json:"valueType,omitempty"` // 值类型
	Ellipsis           bool           `json:"ellipsis"`
	Tooltip            string         `json:"tooltip,omitempty"`
	Copyable           bool           `json:"copyable,omitempty"`
	ValueEnum          map[string]any `json:"valueEnum,omitempty"`
	Order              int            `json:"order,omitempty"` // 查询表单中的权重，权重大排序靠前
	Search             bool           `json:"search"`
	ColSize            int            `json:"colSize,omitempty"`
	HideInSearch       bool           `json:"hideInSearch"`
	HideInTable        bool           `json:"hideInTable"`
	Sorter             bool           `json:"sorter"`
	HideInForm         bool           `json:"hideInForm"`
	HideInDescriptions bool           `json:"hideInDescriptions"`
	Filters            bool           `json:"filters"`
	InitialValue       string         `json:"initialValue,omitempty"`
	Disable            bool           `json:"disable"`
	Render             string         `json:"render,omitempty"` // 渲染函数
}

// Pagination 定义了分页配置
type Pagination struct {
	ShowQuickJumper bool `json:"showQuickJumper,omitempty"`
	PageSize        int  `json:"pageSize,omitempty"`
	Current         int  `json:"current,omitempty"`
	Total           int  `json:"total,omitempty"`
}

// ProTableProps 定义了 Pro Table 的所有属性
type ProTableProps struct {
	Actions       []string               `json:"actions,omitempty"`
	Columns       []*ProTableColumn      `json:"columns,omitempty"`
	RowKey        string                 `json:"rowKey,omitempty"`
	Pagination    *Pagination            `json:"pagination,omitempty"`
	Search        map[string]interface{} `json:"search,omitempty"`
	DateFormatter string                 `json:"dateFormatter,omitempty"`
	HeaderTitle   string                 `json:"headerTitle,omitempty"`
	SubTitle      string                 `json:"subTitle,omitempty"`
	Options       map[string]interface{} `json:"options,omitempty"`
	Params        map[string]interface{} `json:"params,omitempty"`
}

type ResourceQueryParams struct {
	Current  int64                          `json:"current"`
	PageSize int64                          `json:"pageSize"`
	Sort     map[string]contracts.OrderType `json:"sort"`
	Filter   map[string]any                 `json:"filter"`
}

type ResourceQueryFilter struct {
	Condition string `json:"condition"`
	Value     any    `json:"value"`
}

type Resource interface {
	GetTitle() string
	GetName() string
	GetRowKey() string
	GetMeta() (*ProTableProps, contracts.Exception)
	Delete(id int) contracts.Exception
	Update(id int, fields contracts.Fields) contracts.Exception
	Values(value, label string) contracts.Collection[*contracts.Fields]
	Query(params ResourceQueryParams) (contracts.Collection[*contracts.Fields], int64)
}

type Factory interface {
	ExtendResource(resource Resource)
	GetResource(name string) (Resource, contracts.Exception)
	GetResourceListFromDB() ([]*Base, contracts.Exception)
	GetProTablePropsListFromFs() ([]*ProTableProps, contracts.Exception)
	GetResourceFromDB(table string) (*Base, contracts.Exception)
	GetMenuList() []MenuDataItem
	SaveMenuList(list []MenuDataItem) contracts.Exception
	SaveResource(resource Resource) contracts.Exception
}

// ColumnInfo 结构体用于存储表的列信息
type ColumnInfo struct {
	Field   string         `db:"Field"`
	Type    string         `db:"Type"`
	Null    string         `db:"Null"`
	Key     string         `db:"Key"`
	Default sql.NullString `db:"Default"`
	Extra   string         `db:"Extra"`
}

// MenuDataItem represents a menu item with various attributes.
type MenuDataItem struct {
	Children            []MenuDataItem         `json:"children,omitempty"`
	HideChildrenInMenu  bool                   `json:"hideChildrenInMenu,omitempty"`
	HideInMenu          bool                   `json:"hideInMenu,omitempty"`
	Icon                string                 `json:"icon,omitempty"`
	Locale              string                 `json:"locale,omitempty"`
	Name                string                 `json:"name,omitempty"`
	Key                 string                 `json:"key,omitempty"`
	ProLayoutParentKeys []string               `json:"pro_layout_parentKeys,omitempty"`
	Path                string                 `json:"path,omitempty"`
	ParentKeys          []string               `json:"parentKeys,omitempty"`
	ExtraFields         map[string]interface{} `json:"-"` // For additional fields
}
