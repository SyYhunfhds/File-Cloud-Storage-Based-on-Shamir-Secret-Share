package audit

import (
	"backend/internal/claims"
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/util/gmode"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"backend/api/audit/v1"
)

// 参数为当前用户的ID
const queryAllAuditLines = `
select array_agg(distinct (im.id)) as audit_ids
from public.item_members im
         left join public.shares s on im.item_id = s.item_id
where s.owner_id = ?
`

// 参数为上一步获得的ID列表
const passAllAuditLines = `
update public.item_members set status = 1
    where id in (?);
`

// 参数为当前用户的ID
const countAllPass = `
select count(distinct (im.id))
from public.item_members im
         left join public.shares s on im.item_id = s.item_id
where s.owner_id = ? and im.status = 1;
`

type PassAllTempValue struct {
	AuditIds []int `json:"audit_ids"`
	Count    int   `json:"pass_count"` // 通过的数目
}

func (c *ControllerV1) PassAll(ctx context.Context, req *v1.PassAllReq) (res *v1.PassAllRes, err error) {
	if gmode.IsStaging() || gmode.IsProduct() { // 只在开发或测试模式下允许使用这个API
		return nil, gerror.NewCode(gcode.CodeNotImplemented)
	}
	ctx, span := gtrace.NewSpan(ctx, "PassAllMemberHandler")
	defer span.End()

	span.AddEvent("解析并校验用户Claims")
	ac, err := claims.AuthClaimsFromCtx(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "用户未登录")
		span.SetAttributes(attribute.String("request.token.validation.error", err.Error()))

		err = gerror.NewCode(gcode.CodeNotAuthorized, "用户未登录")
		return nil, err
	}

	var v PassAllTempValue
	span.AddEvent("查询用户关联的审核条目")
	err = g.DB().Ctx(ctx).Raw(queryAllAuditLines, ac.Id).Scan(&v)
	if err != nil {
		span.SetStatus(codes.Error, "无法获取用户关联的审核条目")
		span.SetAttributes(attribute.String("db.query.audit_list.error", err.Error()))

		err = gerror.NewCode(gcode.CodeInternalError, "无法获取用户关联的审核条目")
		return nil, err
	}

	span.AddEvent("使用上一步获得的ID列表 更新条目状态")
	result, err := g.DB().Exec(ctx, passAllAuditLines, v.AuditIds) // 这一步是没有结果的
	if err != nil {
		span.SetStatus(codes.Error, "无法更新条目")
		span.SetAttributes(attribute.String("db.update.pass_all.error", err.Error()))

		err = gerror.NewCode(gcode.CodeInternalError, "无法更新条目")
		return nil, err
	}
	var affected, _ = result.RowsAffected()

	res = &v1.PassAllRes{
		TotalPassed: int(affected),
	}
	span.AddEvent("返回结果")
	span.SetStatus(codes.Ok, "所有条目的状态都已修改为Passed")
	return res, nil
}
