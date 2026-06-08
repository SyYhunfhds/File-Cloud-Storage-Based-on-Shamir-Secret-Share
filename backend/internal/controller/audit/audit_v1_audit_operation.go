package audit

import (
	v1 "backend/api/audit/v1"
	"backend/internal/claims"
	"backend/internal/dao"
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type Scope struct {
	OwnerId  int
	AuditId  []int // 申请ID列表, 防止改到其他东西
	Scoped   []int // 属于他的申请者ID列表
	Unscoped []int // 不属于他的申请者ID列表
}

// OperationTempValue 用于业务过程中暂存值
type OperationTempValue struct {
	Scopes             []Scope // 所有者ID和对应的申请者名称
	VerifiedConfirmIds []int   // 校验过的条目列表, 防止越权篡改
}

// 检查请求里面的目标申请者ID有哪些是正确的, 哪些不是
// 参数依次为 用户ID、用户ID、用户ID和请求ID列表
const verifyControlBypass = `
select
    items.owner_id,
    array_agg(im.id) as audit_id,
    array_remove(array_agg(CASE WHEN items.owner_id != ? THEN u.id END), NULL) as unscoped,
    array_remove(array_agg(CASE WHEN items.owner_id = ? THEN u.id END), NULL) as scoped
from public.items
    left join public.item_members im on items.id = im.item_id
    left join public.users u on im.member_id = u.id
where member_id != ? and im.id in (?)
group by items.owner_id;`

func (c *ControllerV1) AuditOperation(ctx context.Context, req *v1.AuditOperationReq) (res *v1.AuditOperationRes, err error) {
	ctx, span := gtrace.NewSpan(ctx, "AuditOperationHandler")
	defer span.End()

	var msg = "解析用户Claims"
	span.AddEvent(msg)
	ac, err := claims.AuthClaimsFromCtx(ctx)
	if err != nil {
		msg = "无法获取用户Claims"
		span.SetStatus(codes.Error, msg)
		span.SetAttributes(attribute.String("request.claims.validation.error", err.Error()))

		err = gerror.NewCode(gcode.CodeNotAuthorized, "无法确认用户登录状态")
		return
	}
	var temp = OperationTempValue{}

	span.AddEvent("校验申请者ID, 避免水平越权")
	_ctx, _span := gtrace.NewSpan(ctx, "校验申请者ID, 避免水平越权")
	var auditIdList = make([]int, 0, len(req.Operations))
	for id, _ := range req.Operations {
		auditIdList = append(auditIdList, id)
	}
	err = g.DB().Ctx(_ctx).Raw(verifyControlBypass, ac.Id, ac.Id, ac.Id, auditIdList).Scan(&temp.Scopes)
	if err != nil {
		span.SetStatus(codes.Error, "申请者ID校验失败")
		span.SetAttributes(attribute.String("db.query.error", err.Error()))

		err = gerror.NewCode(gcode.CodeInternalError, "校验失败")
		_span.End()
		return nil, err
	}
	if len(temp.Scopes) > 1 || temp.Scopes[0].OwnerId != ac.Id {
		span.SetStatus(codes.Error, "检测到水平越权")
		span.SetAttributes(
			attribute.String("request.audit.validation.error", "检测到水平越权"),
			attribute.String("audit.scope.format_string", fmt.Sprintf("%v", temp.Scopes)),
		)

		err = gerror.NewCode(gcode.CodeNotAuthorized, "检测到水平越权")
		_span.End()
		return nil, err
	} // 匹配到了多个所有者, 表现为水平越权
	if len(temp.Scopes) == 0 {
		span.SetStatus(codes.Error, "未检测到有效的审计列表")

		err = gerror.NewCode(gcode.CodeNotFound, "未检测到有效的审计列表")
		_span.End()
		return nil, err
	}
	_span.End()
	temp.VerifiedConfirmIds = temp.Scopes[0].AuditId

	_ctx, _span = gtrace.NewSpan(ctx, "更新审计列表")
	var updateList = make(g.List, len(req.Operations))
	var idx = 0
	for _, id := range temp.VerifiedConfirmIds {
		updateList[idx] = g.Map{
			"id":     id,
			"status": v1.OperationTypeToInt(req.Operations[id]),
		}
		idx++
	}
	// spew.Dump(list)
	result, err := dao.ItemMembers.Ctx(_ctx).Data(updateList).Save()
	if err != nil { // 有些时候更新失败, 可能就是因为发生了水平越权
		span.SetStatus(codes.Error, "更新失败")
		span.SetAttributes(attribute.String("db.update.error", err.Error()))

		err = gerror.NewCode(gcode.CodeInternalError, "检测到水平越权!")
		_span.End()
		return nil, err
	}
	_span.End()

	affected, _ := result.RowsAffected() // 有些时候可能没法计数
	res = &v1.AuditOperationRes{
		Affected: affected,
	}

	msg = "请求处理完毕"
	span.SetStatus(codes.Ok, msg)
	return res, nil
}

func (c *ControllerV1) verifyRecoveryCode(code string, hash string) (bool, error) {
	return c.hu.HashVerify(code, hash)
}
