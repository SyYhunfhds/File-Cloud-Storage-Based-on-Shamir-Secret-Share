package user

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"backend/api/user/v2"
)

func (c *ControllerV2) GetList(ctx context.Context, req *v2.GetListReq) (res *v2.GetListRes, err error) {
	offset := (req.Page - 1) * req.Size
	size := req.Page * req.Size
	{
		_ = offset
		_ = size
	}

	res = &v2.GetListRes{}
	err = g.DB().Raw(`
		select 
		    employees.id as employee_id, u.username as username, u.email as email, j.role_privilege as role_weight, j.job as job
		    from employees 
			left join public.users u on u.id = employees.user_id 
			left join public.jobs j on j.id = employees.job_id
				 limit ? offset ?
`, size, offset).Scan(&res.List)
	if err != nil {
		return
	}

	return
}
