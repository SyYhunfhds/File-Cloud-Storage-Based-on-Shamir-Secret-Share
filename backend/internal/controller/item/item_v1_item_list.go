package item

import (
	"backend/internal/claims"
	"backend/internal/dao"
	"context"

	"backend/api/item/v1"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) ItemList(ctx context.Context, req *v1.ItemListReq) (res *v1.ItemListRes, err error) {
	res = &v1.ItemListRes{}
	var (
		limit  = req.Size
		offset = (req.Page - 1) * req.Size
	)

	ac, err := claims.AuthClaimsFromCtx(ctx)
	// spew.Dump(ac)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "用户未登录")
	}

	err = dao.Items.Ctx(ctx).
		Raw(`
select
    filename, uploaded_at, changed_at, 
    ownerUsers.username as owner, uploaderUsers.username as uploader
    from public.items i
left join public.users ownerUsers on ownerUsers.id = i.owner_id
left join public.users uploaderUsers on uploaderUsers.id = i.uploader_id
where i.owner_id = ? OR i.minimum_privilege < ?
`, ac.Id, ac.Privilege).
		Offset(offset).
		Limit(limit).
		Scan(&res.Items)

	if err != nil {
		// TODO: 封装error
		return nil, err
	}
	res.Count = len(res.Items)

	return res, err
}
