package user

import (
	"backend/internal/dao"
	"backend/internal/model/do"
	"context"

	"backend/api/user/v1"
)

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	_, err = dao.Users.Ctx(ctx).Data(do.Users{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	}).WherePri(req.Id).Update()

	return
}
