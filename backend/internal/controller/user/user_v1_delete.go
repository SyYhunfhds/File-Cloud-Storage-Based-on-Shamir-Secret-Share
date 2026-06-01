package user

import (
	"backend/internal/dao"
	"context"

	"backend/api/user/v1"
)

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	// WherePri接受数据表主键作为参数
	_, err = dao.Users.Ctx(ctx).WherePri(req.Id).Delete()
	return
}
