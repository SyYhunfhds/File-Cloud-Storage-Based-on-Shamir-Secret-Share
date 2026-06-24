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
	"github.com/gogf/gf/v2/net/gtrace"
)

// MaxShares
//
// 根据给定成员数目计算最大份额数, 公式为 2 + x (下同, 只不过变为1 + x)
// 即所有者有一份Device Share, 服务端持有唯一一份Auth Share和Recovery Share, 其他每个审计成员各拥有一份Device Share
func MaxShares(members int) (m int) {
	return members + 2
}
func MaxThreshold(members int) (m int) {
	return members + 1
}

func (c *ControllerV1) ItemUpdate(ctx context.Context, req *v1.ItemUpdateReq) (res *v1.ItemUpdateRes, err error) {
	ctx, span := gtrace.NewSpan(ctx, "ItemUpdateHandler")
	defer span.End()

	ac, err := claims.AuthClaimsFromCtx(ctx)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "用户未登录")
	}

	if req.MinimumPrivilege > ac.Privilege {
		return nil, gerror.NewCode(gcode.CodeInvalidOperation, "权限设置过高, 不予批准")
	}

	// 静默纠正请求阈值
	members, err := dao.CountMembers(ctx, ac.Id, req.ItemId)
	if err != nil {
		return nil, gerror.NewCodef(gcode.CodeNotAuthorized, "请勿访问他人的文件")
	}
	maxThreshold := MaxThreshold(members)
	req.Threshold = min(maxThreshold, req.Threshold) // 选一个较小的值
	req.Threshold = max(2, req.Threshold)            // 再修正一下, 至少为2

	// 这个时候会有水平越权, 但并不会执行
	model := dao.Items.Ctx(ctx).Where("owner_id", ac.Id).WherePri(req.ItemId).
		Data(g.Map{
			"minimum_privilege": req.MinimumPrivilege,
			"filename":          req.NewFilename, // 现在只是改个字段而已的事了
			"is_public":         req.EnablePublic,
			"threshold":         req.Threshold,
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
