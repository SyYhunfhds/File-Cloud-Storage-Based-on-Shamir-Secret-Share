package audit

import (
	v1 "backend/api/audit/v1"
	"backend/internal/claims"
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// 参数分别为
// 用户ID、用户ID和搜索范围
//
// API变动说明: 布尔表达式统一使用 ::int 转为整数数组
// GoFrame的Scan()无法正确处理PostgreSQL boolean[] -> Go []bool 的映射,
// 底层驱动返回的"{false}"文本被gconv当作非空字符串全部转为true。
const queryAuditList = `
select
    items.id as item_id,
    array_agg((im.id::int)) as audit_id,
    filename as item_name, array_agg(im.status) as audit_status, array_agg(u.username) as applicant,
    array_agg((im.status = 0)::int) as rejected,
    array_agg((im.status = 1)::int) as approved,
    array_agg((im.joined_at is not null)::int) as member_confirmed,
    array_agg(im.created_at) as created_at, array_agg(im.updated_at) as updated_at, array_agg(im.joined_at) as joined_at,

    array_agg((im.joined_at is not null)::int) as allow_download,
    array_agg((s.share_base64 != '')::int) as can_download
from public.items
         left join public.item_members im on items.id = im.item_id
         left join public.users u on im.member_id = u.id
         left join public.shares s on items.id = s.item_id
where items.owner_id = ? and im.member_id != ? and s.share_type = 'auth' and im.status in (?)
group by public.items.id, filename, im.id
order by array_agg(im.updated_at) desc
`
const countAuditList = `select
    count(1) as total
from public.items
         left join public.item_members im on items.id = im.item_id
         left join public.users u on im.member_id = u.id
         left join public.shares s on items.id = s.item_id
where items.owner_id = ? and im.member_id != ? and s.share_type = 'auth' and im.status in (?)`

func (c *ControllerV1) AuditList(ctx context.Context, req *v1.AuditListReq) (res *v1.AuditListRes, err error) {
	ctx, span := gtrace.NewSpan(ctx, "AuditListHandler")
	defer span.End()

	span.AddEvent("解析用户Claims")
	ac, err := claims.AuthClaimsFromCtx(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "无法获取用户Claims")
		span.SetAttributes(attribute.String("request.claims.validation.error", err.Error()))

		err = gerror.NewCode(gcode.CodeNotAuthorized, "无法确认用户登录状态")
		return
	}
	res = &v1.AuditListRes{
		List: []v1.LessDetailedAudit{},
	}

	var (
		limit  = req.Size
		offset = (req.Page - 1) * limit
	)

	span.AddEvent("获取审计列表")
	err = g.DB().Ctx(ctx).Raw(queryAuditList, ac.Id, ac.Id, req.Scope).
		Limit(limit).
		Offset(offset).
		Scan(&res.List)
	if err != nil {
		span.SetStatus(codes.Error, "查询审核列表失败")
		span.SetAttributes(attribute.String("db.query.error", err.Error()))

		err = gerror.NewCode(gcode.CodeInternalError, "查询审核列表失败")
		return nil, err
	}
	span.AddEvent("获取审计条目总数")
	err = g.DB().Ctx(ctx).Raw(countAuditList, ac.Id, ac.Id, req.Scope).Scan(&res)
	if err != nil {
		span.SetStatus(codes.Error, "查询审核列表总数失败")
		span.SetAttributes(attribute.String("db.count.error", err.Error()))

		err = gerror.NewCode(gcode.CodeInternalError, "查询审核列表总数失败")
		return nil, err
	}

	span.SetStatus(codes.Ok, "查询审核列表成功")
	return res, nil
}
