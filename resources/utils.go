package resources

import (
	"fmt"
	"strings"

	"github.com/goal-web/contracts"
	"github.com/goal-web/database/table"
	adminModel "github.com/qbhy/goal-admin/models/admin"
)

func BuildQuery[T any](queryProvider func() *table.Table[T]) QueryFunction {
	return func(params adminModel.QueryParams) ([]any, uint64) {
		page, pageSize := NormalizePagination(params)

		q := queryProvider()

		// keyword 模糊搜索（资源/下载链接）
		if kw := strings.TrimSpace(params.Keyword); kw != "" {
			like := fmt.Sprintf("%%%s%%", kw)
			// 参考 collectibles.go 使用 WhereFunc 分组 OR 条件
			q.WhereFunc(func(q contracts.QueryBuilder[T]) {
				q.Where("resource", "like", like).
					OrWhere("url", "like", like)
			})
		}

		// 处理筛选 params
		for _, p := range params.Params {
			key := p.Key
			val := p.Value
			op := p.Operator

			switch op {
			case "like":
				q.Where(key, "like", fmt.Sprintf("%%%s%%", val))
			case "neq", "!=", "<>", "":
				q.Where(key, "!=", val)
			case "in":
				q.WhereIn(key, val)
			case "between":
				if arr, ok := val.([]any); ok && len(arr) == 2 {
					q.WhereRaw(fmt.Sprintf("%s between '%v' and '%v'", key, arr[0], arr[1]))
				}
			case "gt", "<", "<=", ">=":
				q.Where(key, op, val)
			default:
				q.Where(key, val)
			}
		}

		// 排序（支持 ascend/descend → asc/desc）
		if len(params.Sorters) > 0 {
			for _, s := range params.Sorters {
				order := "asc"
				if s.Order == "descend" || s.Order == "desc" {
					order = "desc"
				}
				q.OrderBy(s.Field, contracts.OrderType(order))
			}
		}

		paginator, total := q.Paginate(int64(pageSize), int64(page))
		return paginator.ToAnyArray(), uint64(total)
	}
}

// HasFields 判断在 QueryParams 中是否同时存在指定的参数 key 列表。
// 使用示例：HasFields(params, "status", "id")
// 规则：当 fields 为空时返回 true；当任意一个 key 不存在时返回 false。
func HasFields(params adminModel.QueryParams, fields ...string) bool {
	if len(fields) == 0 {
		return true
	}

	present := make(map[string]struct{}, len(params.Params))
	for _, p := range params.Params {
		present[p.Key] = struct{}{}
	}

	for _, f := range fields {
		if _, ok := present[f]; !ok {
			return false
		}
	}
	return true
}

// NormalizePagination 统一分页参数处理：页码最小为1，页大小最小为10
func NormalizePagination(params adminModel.QueryParams) (page int, pageSize int) {
	page = int(params.Page)
	pageSize = int(params.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return
}
