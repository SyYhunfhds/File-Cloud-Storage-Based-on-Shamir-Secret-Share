package share

import (
	v1 "backend/api/share/v1"
	"backend/internal/claims"
	"backend/internal/dao"
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// 参数为用户ID、[固定为StatusApproved]
const listDeviceShares = `
select
    s.id as share_id,
    s.item_id,
    i.filename,
    s.owner,

    (s.expire_at < now() at time zone 'UTC-8') as is_expired,
    s.expire_at,
    s.updated_at
from public.shares s
left join public.users u on u.id = s.user_id
left join public.items i on s.item_id = i.id
left join public.item_members im on s.item_id = im.item_id
where s.share_type = 'device'
    and s.user_id = ?
    and i.is_public
    and im.status = ?
`

// 参数为用户ID、[固定为StatusApproved]
const countDeviceShares = `
select
   count(1) as total
from public.shares s
         left join public.users u on u.id = s.user_id
         left join public.items i on s.item_id = i.id
         left join public.item_members im on s.item_id = im.item_id
where s.share_type = 'device'
  and s.user_id = ?
  and i.is_public
  and im.status = ?	
`

// ShareList
//
// 满足如下条件其一的(Device Share)份额可被查看
// 1. 用户未确认加入某一条目且条目状态为Pass/1, 但已经在Shares表拥有对应的Device Share, 且Device Share未过期
// 2. 用户已确认加入某一条目且条目状态仍为Pass/1, 且已经在Shares表拥有对应的Device Share, 且Device Share未过期
//
// 补充
// 1. 本API没有水平越权拦截检测, 但有简单的RLS策略, 确保不会搜索到别人的份额
//
// 其他问题
// 1. 有些数据会出现UpdatedAt比ExpireAt晚8小时的现象, 这是因为早期测试时部分时间戳更新用的是Pgsql自己的特性, 没有正确同步系统时区设置
func (c *ControllerV1) ShareList(ctx context.Context, req *v1.ShareListReq) (res *v1.ShareListRes, err error) {
	ctx, span := gtrace.NewSpan(ctx, "ShareListHandler")
	defer span.End()

	span.AddEvent("解析并校验用户Claims")
	ac, err := claims.AuthClaimsFromCtx(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "用户未登录")
		span.RecordError(err)

		err = gerror.NewCode(gcode.CodeNotAuthorized, "用户未登录")
		return nil, err
	}

	var (
		limit  = req.Size
		offset = (req.Page - 1) * limit
	)

	res = &v1.ShareListRes{}
	err = g.DB().
		Ctx(ctx).
		Raw(listDeviceShares, ac.Id, dao.SubmissionApproved).
		Limit(limit).
		Offset(offset).
		Scan(&res.List)
	if err != nil {
		span.SetStatus(codes.Error, "查询失败")
		span.RecordError(err)

		err = gerror.NewCode(gcode.CodeInternalError, "查询失败")
		return nil, err
	}

	span.AddEvent("统计总数")
	err = g.DB().
		Ctx(ctx).
		Raw(countDeviceShares, ac.Id, dao.SubmissionApproved).
		Scan(&res)
	if err != nil {
		// 这个不影响API响应
		span.SetAttributes(attribute.String("db.count.error", err.Error()))
	}
	if len(res.List) == 0 { // 有时候没扫描到东西的话这个列表就是nil的
		res.List = []v1.ShareInfo{} // 给个空列表
	}

	return res, nil
}
