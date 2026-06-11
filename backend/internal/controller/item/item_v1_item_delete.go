package item

import (
	v1 "backend/api/item/v1"
	"backend/internal/claims"
	"backend/internal/dao"
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// 参数依次为 用户ID和条目ID列表
const queryGroupItemByOwner = `
select
    array_agg(items.id) as item_ids
from public.items
where items.owner_id = ? and items.id in (?) 
`
const deleteItem = `
delete from public.items where id in (?) and owner_id = ?
`
const queryItemSavename = `
select array_agg(savename) as savenames from public.items where id in(?)
`

type DeleteHandlerTempValue struct {
	ItemIds   []int
	Savenames []string
}

func (c *ControllerV1) ItemDelete(ctx context.Context, req *v1.ItemDeleteReq) (res *v1.ItemDeleteRes, err error) {
	ctx, span := gtrace.NewSpan(ctx, "ItemDeleteHandler")
	defer span.End()

	ac, err := claims.AuthClaimsFromCtx(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "无法获取用户Claims")
		err = gerror.NewCode(gcode.CodeNotAuthorized, "用户未登录")
		return
	}
	_ = ac

	var v DeleteHandlerTempValue
	span.AddEvent("检查水平越权")
	_ctx, _span := gtrace.NewSpan(ctx, "检查是否存在水平越权")
	err = g.DB().Ctx(_ctx).Raw(queryGroupItemByOwner, ac.Id, req.ItemIds).Scan(&v)
	if err != nil {
		_span.SetStatus(codes.Error, "查询数据库失败")
		_span.SetAttributes(attribute.String("db.query.item_id_list.error", err.Error()))

		err = gerror.NewCode(gcode.CodeInternalError, "后端发生未知错误")
	}
	if err != nil {
		span.RecordError(err)
		_span.End()
		return nil, err
	}

	span.SetAttributes(
		attribute.IntSlice("request.item_id_list.to_delete", req.ItemIds),
		attribute.IntSlice("db.examine.item_id_list", v.ItemIds),
	)
	if len(v.ItemIds) != len(req.ItemIds) {
		span.SetStatus(codes.Error, "存在水平越权")

		err = gerror.NewCode(gcode.CodeSecurityReason, "不允许删除其他用户的条目")
		_span.End()
		return nil, err
	}

	span.AddEvent("检索保存路径")
	_ctx, _span = gtrace.NewSpan(ctx, "检索保存路径")
	err = g.DB().Ctx(_ctx).Raw(queryItemSavename, req.ItemIds).Scan(&v)
	if err != nil {
		_span.SetStatus(codes.Error, "无法检索到文件的保存路径")
		_span.SetAttributes(attribute.String("db.query.savename_list.error", err.Error()))

		_span.End()
		err = gerror.NewCode(gcode.CodeInternalError, "后端发生未知错误")
		return nil, err
	}
	_span.End()

	var affected int64
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_ctx, _span = gtrace.NewSpan(ctx, "删除条目")
		defer _span.End()
		_span.AddEvent("先删数据库记录")
		_, err := dao.Items.Ctx(_ctx).Delete("id in (?)", req.ItemIds)
		if err != nil {
			_span.SetStatus(codes.Error, "删除条目失败")
			_span.SetAttributes(attribute.String("db.delete.error", err.Error()))

			return err
		}

		_span.AddEvent("再删文件")
		for _, savename := range v.Savenames {
			if err = c.fu.DeleteItem(savename); err != nil {
				span.SetAttributes(attribute.String("file.delete.error", err.Error()))
			} else {
				affected++ // 删除成功才算影响行数
			}
		}
		return nil
	})
	if err != nil {
		span.RecordError(err)
		err = gerror.NewCode(gcode.CodeInternalError, "后端发生未知错误")
		return nil, err
	}

	res = &v1.ItemDeleteRes{
		Deleted: affected,
	}

	return res, nil
}
