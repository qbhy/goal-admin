package goal_admin

import "github.com/qbhy/goal-admin/resources"

type Config struct {
	Logo  string
	Title string

	PermissionHandler resources.PermissionHandler
}
