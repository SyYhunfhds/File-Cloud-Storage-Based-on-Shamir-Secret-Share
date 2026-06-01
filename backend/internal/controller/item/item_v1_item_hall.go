package item

import (
	v1 "backend/api/item/v1"
	"backend/internal/dao"
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/gtrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (c *ControllerV1) ItemHall(ctx context.Context, req *v1.ItemHallReq) (res *v1.ItemHallRes, err error) {
	ctx, span := gtrace.NewSpan(ctx, "ItemHallHandler")
	defer span.End()

	var (
		limit  = req.Size
		offset = (req.Page - 1) * req.Size
	)
	res = &v1.ItemHallRes{}

	err = dao.Items.DB().Ctx(ctx).Raw(
		`
select
    items.id as item_id, ou.username as owner, uu.username as uploader,
    items.filename, items.uploaded_at, items.changed_at, items.deleted_at
from public.items
         left join public.users ou on items.owner_id = ou.id
         left join public.users uu on items.uploader_id = uu.id
where is_public=true
order by uploaded_at desc
limit ? offset ?`, limit, offset).Scan(&res.Items)

	_, res.Total, err = dao.Items.Ctx(ctx).Where("is_public", true).AllAndCount(false)
	if err != nil {
		span.SetStatus(codes.Error, "无法获取数据")
		span.SetAttributes(attribute.String("db.query.error", err.Error()))

		return nil, gerror.NewCode(gcode.CodeInvalidOperation, "搜索不到相关信息")
	}

	if temp := res.Total % limit; temp == 0 {
		res.AvailablePages = res.Total / limit
	} else {
		res.AvailablePages = res.Total/limit + 1
	}

	return res, nil
}
