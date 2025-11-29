package goal_admin

import (
	"github.com/qbhy/goal-admin/models"
	"github.com/qbhy/goal-admin/models/admin"

	"github.com/goal-web/contracts"
)

type Factory interface {
	Register(resource Resource)
	Get(name string) (Resource, bool)
}

// Resource 定义一个可提供元数据的资源
type Resource interface {
	GetName() string
	Meta() admin.ResourceMeta

	Can(admin *models.AdminModel, action string) bool

	Export(params admin.QueryParams) (*models.ExportModel, error)
	Create(fields contracts.Fields) (any, error)
	BatchDelete(ids []int) error
	Delete(id int) error
	Update(id int, fields contracts.Fields) (any, error)
	Find(id int) (any, error)
	List(params admin.QueryParams) ([]any, uint64) // 返回列表和页数
	BatchFields(keyField, labelField string, ids []string) (map[string]any, error)
	Action(action, payload string) (any, error)
}
