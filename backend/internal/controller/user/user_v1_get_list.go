package user

import (
	"backend/internal/dao"
	"context"

	"backend/api/user/v1"
)

func (c *ControllerV1) GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error) {
	offset := (req.Page - 1) * req.Size
	size := req.Page * req.Size

	res = &v1.GetListRes{}
	err = dao.Users.Ctx(ctx).
		Offset(offset).
		Limit(size).
		FieldsEx("password").
		Scan(&res.List)

	return
}
