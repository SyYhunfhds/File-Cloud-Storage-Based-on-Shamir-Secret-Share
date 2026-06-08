package item

import (
	"backend/internal/claims"
	"backend/internal/dao"
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/os/gtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"backend/api/item/v1"
)

type TempValue struct {
	ItemId []int `json:"item_ids" form:"item_ids" query:"item_ids"` // 校验过之后的有效ID
}

// 查询有效的条目ID, 参数为当前用户的ID
const queryValidItemIds = `select 
    array_agg(items.id) as item_id 
from public.items 
where owner_id != ? and is_public = true`

func (c *ControllerV1) ApplyForViewing(ctx context.Context, req *v1.ApplyForViewingReq) (res *v1.ApplyForViewingRes, err error) {
	ctx, span := gtrace.NewSpan(ctx, "ApplyForViewingHandler")
	defer span.End()

	ac, err := claims.AuthClaimsFromCtx(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "无法获取用户信息")
		span.SetAttributes(attribute.String("request.claims.validation.error", err.Error()))

		err = gerror.NewCode(gcode.CodeNotAuthorized, "用户未登录")
		return
	}

	span.SetAttributes(
		attribute.IntSlice("request.item_ids", req.ItemIds),
	)

	var temp = TempValue{}
	_ctx, _span := gtrace.NewSpan(ctx, "校验有效的ItemId")
	err = g.DB().Ctx(_ctx).Raw(queryValidItemIds, ac.Id).Scan(&temp)
	if err != nil {
		_span.SetStatus(codes.Error, "无法查询有效的ItemId")
		_span.SetAttributes(attribute.String("db.examine.error", err.Error()))
	}
	if len(temp.ItemId) != len(req.ItemIds) {
		_span.SetStatus(codes.Error, "检查到水平越权")
		err = gerror.NewCode(gcode.CodeInvalidOperation, "检测到可能的水平越权")
		_span.End()
		return
	}
	_span.End()
	if err != nil {
		err = gerror.NewCode(gcode.CodeInternalError, "无法查询有效的ItemId")
		return
	}

	var updateList = make(g.List, 0, len(temp.ItemId))
	for _, id := range temp.ItemId {
		updateList = append(updateList, g.Map{
			"item_id":   id,
			"member_id": ac.Id,

			"updated_at": gtime.Now(),
		})
	}
	_ctx, _span = gtrace.NewSpan(ctx, "更新申请记录")
	result, err := dao.ItemMembers.Ctx(_ctx).Data(updateList).OnConflict("item_id", "member_id").Save()
	if err != nil {
		_span.SetStatus(codes.Error, "无法更新申请记录")
		_span.SetAttributes(attribute.String("db.update.error", err.Error()))
	}
	_span.End()
	if err != nil {
		err = gerror.NewCode(gcode.CodeInternalError, "无法更新申请记录")
		return
	}

	affected, _ := result.RowsAffected() // Psql支持这个的, 所以应该不需要处理错误
	res = &v1.ApplyForViewingRes{
		TotalApplied: affected,
	}

	return res, nil
}
