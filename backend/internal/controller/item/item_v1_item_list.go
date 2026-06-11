package item

import (
	"backend/internal/claims"
	"backend/internal/dao"
	"context"

	"backend/api/item/v1"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/os/gctx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const queryOnlyBelongsToUser string = `
select
    i.id as item_id, filename, uploaded_at, changed_at,
    ownerUsers.username as owner, uploaderUsers.username as uploader,
    true as can_download
from public.items i
         left join public.users ownerUsers on ownerUsers.id = i.owner_id
         left join public.users uploaderUsers on uploaderUsers.id = i.uploader_id
         left join public.item_members im on i.id = im.item_id
where i.owner_id = ? or (im.member_id = ? and im.status = 1)
order by uploaded_at desc
`
const countOnlyBelongsToUser string = `
select count(1) as total from public.items
                left join public.item_members im on items.id = im.item_id
                where owner_id = ? 
                   OR (im.member_id = ? and im.status = 1 and im.joined_at is not null)
`
const queryAll string = `
select distinct 
    i.id as item_id, filename, uploaded_at, changed_at,
    ownerUsers.username as owner, uploaderUsers.username as uploader,
    i.owner_id = ? or (im.member_id = ? and im.status = 1 and im.joined_at is not null) as can_download
from public.items i
         left join public.users ownerUsers on ownerUsers.id = i.owner_id
         left join public.users uploaderUsers on uploaderUsers.id = i.uploader_id
         left join public.item_members im on i.id = im.item_id
where i.is_public=true or i.owner_id = ?
order by uploaded_at desc
`

const countAll = `
select count(distinct items.id) as total from public.items
                         left join public.item_members im on items.id = im.item_id
where items.is_public=true or items.owner_id = ?
`

func (c *ControllerV1) ItemList(ctx context.Context, req *v1.ItemListReq) (res *v1.ItemListRes, err error) {
	ctx, span := gtrace.NewSpan(ctx, "ItemListHandler")
	defer span.End()
	span.SetAttributes(
		attribute.String("request.parameter.search_scope.comment", "搜索范围; 1为仅自己, 2为仅他人公开搜索, 3为混合; 默认只看自己的"),
		attribute.Int("request.parameter.search_scope", req.SearchScope),
	)

	res = &v1.ItemListRes{}
	result := &v1.ItemHallRes{}
	var (
		limit  = req.Size
		offset = (req.Page - 1) * req.Size
	)

	ac, err := claims.AuthClaimsFromCtx(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "无法获取用户Claims")
		span.SetAttributes(attribute.String("token.validation.error", err.Error()))

		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "用户未登录")
	}

	switch req.SearchScope {
	case v1.SearchOnlyBelongsToUser:
		goto OnlyBelongsToUser
	case v1.SearchOnlyPublic:
		goto OnlyPublic
	case v1.SearchAll:
		goto All
	default:
		span.SetStatus(codes.Error, "暂不支持其他搜索模式")
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "暂不支持其他搜索模式")
	}

OnlyBelongsToUser:
	err = dao.Items.Ctx(ctx). // 搜索自己拥有的和自己允许拥有的
					Raw(queryOnlyBelongsToUser, ac.Id, ac.Id).
					Offset(offset).
					Limit(limit).
					Scan(&res.Items)
	if err != nil {
		span.SetStatus(codes.Error, "无法获取条目列表")
		span.SetAttributes(attribute.String("db.query.error", err.Error()))

		return nil, gerror.NewCode(gcode.CodeInternalError, "查询失败")
	}

	err = g.DB().Ctx(ctx).
		Raw(countOnlyBelongsToUser, ac.Id, ac.Id).
		Scan(&res)
	if err != nil {
		span.SetStatus(codes.Error, "无法统计条目总数")
		span.SetAttributes(attribute.String("db.query.error", err.Error()))

		return nil, gerror.NewCode(gcode.CodeInternalError, "查询失败")
	}
	goto Ret

OnlyPublic:
	ctx = gctx.WithSpan(ctx, "路由到ItemHallHandler")
	result, err = c.ItemHall(ctx, &v1.ItemHallReq{
		Page: req.Page,
		Size: req.Size,
	})

	if err != nil {
		span.SetStatus(codes.Error, "无法获取条目列表")
		span.SetAttributes(attribute.String("db.query.error", err.Error()))

		return nil, gerror.NewCode(gcode.CodeInternalError, "查询失败")
	}
	res.Items = result.Items
	res.Total = result.Total

All: // 查询自己所有的条目和所有可被公开搜索的条目
	err = g.DB().Ctx(ctx).Raw(queryAll, ac.Id, ac.Id, ac.Id).
		Limit(limit).
		Offset(offset).
		Scan(&res.Items)
	if err != nil {
		span.SetStatus(codes.Error, "无法获取条目列表")
		span.SetAttributes(attribute.String("db.query.error", err.Error()))

		return nil, gerror.NewCode(gcode.CodeInternalError, "查询失败")
	}

	err = g.DB().Ctx(ctx).Raw(countAll, ac.Id).Scan(&res)
	if err != nil {
		span.SetStatus(codes.Error, "无法统计条目总数")
		span.SetAttributes(attribute.String("db.query.error", err.Error()))

		return nil, gerror.NewCode(gcode.CodeInternalError, "查询失败")
	}

Ret:
	if len(res.Items) == 0 {
		res.Items = make([]v1.ItemInfo, 0) // 给个空切片
	}
	return res, nil
}
