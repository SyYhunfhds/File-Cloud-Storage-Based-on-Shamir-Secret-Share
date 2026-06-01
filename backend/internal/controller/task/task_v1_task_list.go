package task

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"backend/api/task/v1"
)

func (c *ControllerV1) TaskList(ctx context.Context, req *v1.TaskListReq) (res *v1.TaskListRes, err error) {
	res = &v1.TaskListRes{}
	var (
		limit  = req.Size
		offset = (req.Page - 1) * req.Size
	)

	err = g.DB().Raw(
		`
select
    id as task_id, filename, type, 
    is_succeed as success, failed_for_why as cause, created_at, finished_at
    from public.tasks order by tasks.finished_at desc 
	limit ? offset ? `, limit, offset,
	).ScanAndCount(&res.List, &res.Count, true)

	if err != nil {
		return nil, err
	}

	return res, nil
}
