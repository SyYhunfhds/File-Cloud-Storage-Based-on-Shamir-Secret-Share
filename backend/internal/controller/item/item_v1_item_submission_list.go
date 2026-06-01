package item

import (
	"backend/internal/claims"
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"backend/api/item/v1"
)

func (c *ControllerV1) ItemSubmissionList(ctx context.Context, req *v1.ItemSubmissionListReq) (res *v1.ItemSubmissionListRes, err error) {
	_, span := gtrace.NewSpan(ctx, "查看用户关联的条目查看申请")
	defer span.End()
	ac, err := claims.AuthClaimsFromCtx(ctx)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "用户未登录")
	}

	var (
		limit  = req.Size
		offset = (req.Page - 1) * req.Size
	)

	res = &v1.ItemSubmissionListRes{}
	err = g.DB().Ctx(ctx).Raw(`
select
    item_members.id as submission_id,
    i.filename as target_item,
    su.username as submitter,
    status,

    item_members.updated_at as submitted_at,
    item_members.joined_at is not null as is_approved,
    item_members.joined_at as approved_at

from public.item_members
         left join public.items i on item_members.item_id = i.id
         left join public.users su on item_members.member_id = su.id
    where item_id in (
    select id from public.items i where i.owner_id = ?
    ) and member_id != ?
    order by item_members.updated_at desc
limit ? offset ?
`, ac.Id, ac.Id, limit, offset).ScanAndCount(&res.Submissions, &res.Total, false)
	if err != nil {
		span.SetAttributes(
			attribute.String("db.error", err.Error()),
		)
		span.SetStatus(codes.Error, "数据库处理异常")
		return nil, gerror.NewCode(gcode.CodeInternalError, "无法处理用户请求, 请联系管理员")
	}

	return res, nil
}
