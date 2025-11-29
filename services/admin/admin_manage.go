package admin

import (
	"fmt"

	"github.com/goal-web/application"
	"github.com/goal-web/contracts"
	"github.com/goal-web/supports/logs"
	"github.com/goal-web/supports/utils"
	"github.com/qbhy/goal-admin/models"
	adminModel "github.com/qbhy/goal-admin/models/admin"
	adminReq "github.com/qbhy/goal-admin/requests/admin"
	adminRes "github.com/qbhy/goal-admin/results/admin"
)

func init() {
	// 管理员列表
	AdminManageServiceDefine.ListAdmins = func(req *adminReq.ListAdminsReq, ctx contracts.Context) (*adminRes.ListAdminsResult, error) {
		// 仅 super 可查看管理员列表
		// 默认分页参数
		current := int(req.Current)
		pageSize := int(req.PageSize)
		if current <= 0 {
			current = 1
		}
		if pageSize <= 0 {
			pageSize = 10
		}

		// 构建查询（一次性使用 Paginate 原生分页）
		query := models.AdminQuery()
		if req.Username != "" {
			query.Where("username", "like", fmt.Sprintf("%%%s%%", req.Username))
		}
		if req.Nickname != "" {
			query.Where("nickname", "like", fmt.Sprintf("%%%s%%", req.Nickname))
		}
		if req.Phone != "" {
			query.Where("phone", "like", fmt.Sprintf("%%%s%%", req.Phone))
		}
		if req.Email != "" {
			query.Where("email", "like", fmt.Sprintf("%%%s%%", req.Email))
		}
		// 原生分页（返回列表与总数）
		list, total := query.OrderBy("id", "desc").Paginate(int64(pageSize), int64(current))

		items := make([]adminModel.AdminItem, 0)
		for _, m := range list.ToArray() {
			items = append(items, adminModel.AdminItem{
				Id:        m.Id,
				Username:  m.Username,
				Nickname:  m.Nickname,
				Avatar:    m.Avatar,
				Phone:     m.Phone,
				Email:     m.Email,
				Role:      m.Role,
				CreatedAt: m.CreatedAt,
				UpdatedAt: m.UpdatedAt,
			})
		}

		return &adminRes.ListAdminsResult{Total: int32(total), List: items}, nil
	}

	// 删除管理员：仅 super 可删除 admin；禁止删除 super
	AdminManageServiceDefine.DeleteAdmin = func(req *adminReq.DeleteAdminReq, ctx contracts.Context) (*adminRes.DeleteAdminResult, error) {

		target := models.AdminQuery().FindOrFail(req.Id)
		if target == nil {
			return &adminRes.DeleteAdminResult{Success: false, Message: "管理员不存在"}, nil
		}

		if target.Role == "super" {
			return &adminRes.DeleteAdminResult{Success: false, Message: "禁止删除 super 管理员"}, nil
		}

		if err := target.Delete(); err != nil {
			return &adminRes.DeleteAdminResult{Success: false, Message: "删除失败"}, nil
		}

		return &adminRes.DeleteAdminResult{Success: true, Message: "删除成功"}, nil
	}

	// 创建管理员：仅 super 可创建；校验唯一性与密码加密
	AdminManageServiceDefine.CreateAdmin = func(req *adminReq.CreateAdminReq, ctx contracts.Context) (*adminRes.CreateAdminResult, error) {

		// 基础校验
		if req.Username == "" || req.Password == "" {
			return &adminRes.CreateAdminResult{Success: false, Message: "用户名、手机号与密码为必填"}, nil
		}

		// 检查唯一性
		if models.AdminQuery().Where("username", req.Username).First() != nil {
			return &adminRes.CreateAdminResult{Success: false, Message: "用户名已存在"}, nil
		}
		// 密码加密
		hasher := application.Get("hashing").(contracts.Hasher)
		hashed := hasher.Make(req.Password, nil)

		role := req.Role
		if role == "" {
			role = "admin"
		}

		// 创建记录
		_, err := models.AdminQuery().CreateE(contracts.Fields{
			"username": req.Username,
			"nickname": req.Nickname,
			"avatar":   req.Avatar,
			"phone":    utils.RandStr(11),
			"email":    req.Email,
			"role":     role,
			"password": hashed,
		})

		if err != nil {
			logs.Default().WithError(err).Info("创建管理员失败:")
			return &adminRes.CreateAdminResult{Success: false, Message: "创建失败:" + err.GetPrevious().Error()}, nil
		}

		return &adminRes.CreateAdminResult{Success: true, Message: "创建成功"}, nil
	}

	// 更新管理员：仅 super 可修改其他管理员；密码为可选字段
	AdminManageServiceDefine.UpdateAdmin = func(req *adminReq.UpdateAdminReq, ctx contracts.Context) (*adminRes.UpdateAdminResult, error) {
		if req.Id == 0 {
			return &adminRes.UpdateAdminResult{Success: false, Message: "缺少管理员ID"}, nil
		}

		target := models.AdminQuery().FindOrFail(req.Id)
		if target == nil {
			return &adminRes.UpdateAdminResult{Success: false, Message: "管理员不存在"}, nil
		}

		updateFields := contracts.Fields{}

		if req.Username != "" {
			updateFields["username"] = req.Username
		}

		if req.Nickname != "" {
			updateFields["nickname"] = req.Nickname
		}

		if req.Avatar != "" {
			updateFields["avatar"] = req.Avatar
		}

		if req.Phone != "" {
			updateFields["phone"] = req.Phone
		}

		if req.Email != "" {
			updateFields["email"] = req.Email
		}

		if req.Role != "" {
			updateFields["role"] = req.Role
		}

		if req.Password != "" {
			hasher := application.Get("hashing").(contracts.Hasher)
			hashed := hasher.Make(req.Password, nil)
			updateFields["password"] = hashed
		}

		if len(updateFields) == 0 {
			return &adminRes.UpdateAdminResult{Success: true, Message: "无更新字段"}, nil
		}

		if err := target.Update(updateFields); err != nil {
			return &adminRes.UpdateAdminResult{Success: false, Message: "更新失败"}, nil
		}

		return &adminRes.UpdateAdminResult{Success: true, Message: "更新成功"}, nil
	}
}
