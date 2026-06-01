package user

import (
	"backend/internal/claims"
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"backend/api/user/v3"
)

func (c *ControllerV3) Me(ctx context.Context, req *v3.MeReq) (res *v3.MeRes, err error) {
	ac, err := claims.AuthClaimsFromCtx(ctx)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "用户未登录")
	}
	_ = ac

	res = &v3.MeRes{}
	err = g.DB().Ctx(ctx).Raw(`
select
    u.username, u.email, u.created_at as registered_at, j.job, j.role_privilege as privilege
from public.users u
left join public.employees e on e.user_id = u.id
left join public.jobs j on e.job_id = j.id
where u.id = ?
`, ac.Id).Scan(&res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
