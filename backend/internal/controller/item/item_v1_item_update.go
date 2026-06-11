package item

import (
	"backend/internal/claims"
	"backend/internal/dao"
	"context"
	"time"

	"backend/api/item/v1"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) ItemUpdate(ctx context.Context, req *v1.ItemUpdateReq) (res *v1.ItemUpdateRes, err error) {
	ac, err := claims.AuthClaimsFromCtx(ctx)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "用户未登录")
	}

	if req.MinimumPrivilege > ac.Privilege {
		return nil, gerror.NewCode(gcode.CodeInvalidOperation, "权限设置过高, 不予批准")
	}

	// 这个时候会有水平越权, 但并不会执行
	model := dao.Items.Ctx(ctx).Where("owner_id", ac.Id).WherePri(req.ItemId).
		Data(g.Map{
			"minimum_privilege": req.MinimumPrivilege,
			"filename":          req.NewFilename, // 现在只是改个字段而已的事了
			"is_public":         req.EnablePublic,
			"changed_at":        time.Now(),
		})

	_, err = model.Update()
	if err != nil {
		switch gerror.Code(err) {
		case gcode.CodeMissingParameter: // updating table with empty data
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, "不允许用空数据更新")
		default:
			return nil, err
		}
	}

	res = &v1.ItemUpdateRes{}
	res.Message = "更新成功"

	return res, nil
}
