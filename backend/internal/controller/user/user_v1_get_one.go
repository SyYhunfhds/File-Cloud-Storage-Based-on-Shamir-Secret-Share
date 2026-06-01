package user

import (
	"backend/internal/dao"
	"context"

	"backend/api/user/v1"
)

const (
	stringWithStars = "******" // 原本用于手动脱敏, 后来发现gf orm提供了FieldsEx接口来排除字段
)

func (c *ControllerV1) GetOne(ctx context.Context, req *v1.GetOneReq) (res *v1.GetOneRes, err error) {
	res = &v1.GetOneRes{}
	err = dao.Users.Ctx(ctx).
		WherePri(req.Id).
		FieldsEx("password"). // 排除密码字段
		Scan(&res.User)

	return
}
