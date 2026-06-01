package user

import (
	"context"

	"backend/api/user/v2"

	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV2) GetOne(ctx context.Context, req *v2.GetOneReq) (res *v2.GetOneRes, err error) {
	res = &v2.GetOneRes{}
	err = g.DB().Raw(`
		select 
		    employees.id as employee_id, u.username as username, u.email as email, j.role_privilege as role_weight, j.job as job
		    from employees 
			left join public.users u on u.id = employees.user_id 
			left join public.jobs j on j.id = employees.job_id
		    where public.employees.id = ?
`, req.Id).Scan(&res)

	return
}
