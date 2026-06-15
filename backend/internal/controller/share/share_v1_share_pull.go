package share

import (
	"backend/internal/claims"
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
	"go.opentelemetry.io/otel/codes"

	"backend/api/share/v1"
)

type SimpleShare struct {
	ShareId   int
	IsExpired bool
}
type StepTwo struct {
	ItemIds   []int
	Filenames []string
	Share     []string // 都是base64编码, 可以直接发给用户
}
type PullTempValue struct {
	StepOne []SimpleShare
	StepTwo StepTwo
}

// 参数为用户ID 和 请求携带的目标份额ID
const queryShareIfExpire = `
select
    s.id as share_id,
    (s.expire_at < now() at time zone 'UTC-8') as is_expired
from public.shares s
left join public.items i on s.item_id = i.id
left join public.item_members im on s.item_id = im.item_id
where
    s.share_type = 'device'
    and s.user_id = ?
    and s.id in (?)
	and i.is_public = true
	and im.status = 1 -- 通过审核
`

// 参数为用户ID 和 上一步过滤得到的ID
const retrieveShare = `
select
    array_agg(s.item_id) as item_ids, -- 用于第四步定向更新JoinedAt (但也可能不需要更新)
    array_agg(i.filename) as filenames,
    array_agg(s.share_base64) as share
from public.shares s
left join public.items i on s.item_id = i.id
where
    s.share_type = 'device'
  and s.user_id = ?
  and s.id in (?)
  and s.expire_at >= now() at time zone 'UTC-8'
`

//	(删除份额)
//
// 参数为上一步得到的ID列表 、 用户ID 、[固定为ShareTypeDevice]
const deleteShares = `
delete from public.shares s
where s.id in (?) and s.user_id = ? and s.share_type = 'device';
`

// SharePull
// 获取份额并删除份额缓存
func (c *ControllerV1) SharePull(ctx context.Context, req *v1.SharePullReq) (res *v1.SharePullRes, err error) {
	ctx, span := gtrace.NewSpan(ctx, "SharePullHandler")
	defer span.End()

	ac, err := claims.AuthClaimsFromCtx(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "用户未登录")
		span.RecordError(err)

		err = gerror.NewCode(gcode.CodeNotAuthorized, "用户未登录")
		return nil, err
	}

	var v PullTempValue
	span.AddEvent("查询未过期的份额")
	_ctx, _span := gtrace.NewSpan(ctx, "查询可能有效的份额, 并在应用层按照是否过期进行过滤")
	err = g.DB().Ctx(_ctx).Raw(queryShareIfExpire, ac.Id, req.ShareIds).Scan(&v.StepOne)
	if err != nil {
		_span.RecordError(err)
	}
	_span.End()
	if err != nil {
		span.SetStatus(codes.Error, "查询份额失败")
		err = gerror.NewCode(gcode.CodeInternalError, "查询份额失败")
		return nil, err
	}

	span.AddEvent("业务层过滤未过期的份额")
	var notExpiredIds = make([]int, 0, len(req.ShareIds))
	for _, v := range v.StepOne {
		if !v.IsExpired {
			notExpiredIds = append(notExpiredIds, v.ShareId)
		}
	}
	if len(notExpiredIds) == 0 {
		err = gerror.NewCode(gcode.CodeNotFound, "所有份额全部过期, 请重新提交申请来刷新份额状态")
		return nil, err
	}

	span.AddEvent("从数据库中检索份额")
	_ctx, _span = gtrace.NewSpan(ctx, "从数据库中检索份额")
	err = g.DB().Ctx(_ctx).Raw(retrieveShare, ac.Id, notExpiredIds).Scan(&v.StepTwo)
	if err != nil {
		_span.RecordError(err)
	}
	_span.End()
	if err != nil {
		span.SetStatus(codes.Error, "检索份额失败")
		err = gerror.NewCode(gcode.CodeInternalError, "检索份额失败")
		return nil, err
	}

	span.AddEvent("删除获取的份额")
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_ctx, _span := gtrace.NewSpan(ctx, "删除获取的份额")
		defer _span.End()

		_, err := g.DB().Exec(_ctx, deleteShares, notExpiredIds, ac.Id)
		if err != nil {
			_span.RecordError(err)
			return err
		}
		return nil
	})
	if err != nil {
		span.SetStatus(codes.Error, "删除份额失败")
		return nil, err
	}

	res = &v1.SharePullRes{
		List: make([]v1.Share, 0, len(v.StepTwo.Share)),
	}
	for i, share := range v.StepTwo.Share {
		res.List = append(res.List, v1.Share{
			ItemId:      v.StepTwo.ItemIds[i],
			Filename:    v.StepTwo.Filenames[i],
			DeviceShare: share,
		})
	}
	return res, nil
}
