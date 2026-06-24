package item

import (
	"backend/internal/claims"
	"backend/internal/dao"
	"context"

	"backend/api/item/v1"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/gtrace"
)

func (c *ControllerV1) GetOneItem(ctx context.Context, req *v1.GetOneItemReq) (res *v1.GetOneItemRes, err error) {
	ctx, span := gtrace.NewSpan(ctx, "ItemDetailHandler")
	defer span.End()

	ac, err := claims.AuthClaimsFromCtx(ctx)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "请先登录")
	}

	members, err := dao.CountMembers(ctx, ac.Id, req.ItemId)
	if err != nil {
		span.RecordError(err)
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "请勿访问他人文件")
	}

	res = &v1.GetOneItemRes{}
	err = dao.Items.Ctx(ctx).Raw(`
select
    i.id as item_id, filename, uploaded_at, changed_at,
    ownerUsers.username as owner, uploaderUsers.username as uploader,
    i.threshold as curr_threshold,
    i.minimum_privilege, i.is_public
from public.items i
         left join public.users ownerUsers on ownerUsers.id = i.owner_id
         left join public.users uploaderUsers on uploaderUsers.id = i.uploader_id
where i.owner_id = ? and i.id = ?
`, ac.Id, req.ItemId).
		Limit(1).
		Scan(&res)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "找不到条目信息")
	}

	{
		res.CurrMembers = members
		res.MinThreshold = 2
		res.MaxThreshold = MaxThreshold(members) // 最大可设置的份额数
		res.Shares = MaxShares(members)
	}

	return res, nil
}
