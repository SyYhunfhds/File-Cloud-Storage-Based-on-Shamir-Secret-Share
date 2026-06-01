package item

import (
	"backend/internal/claims"
	"backend/internal/dao"
	"backend/internal/model/entity"
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"backend/api/item/v1"
)

func (c *ControllerV1) ItemRequireView(ctx context.Context, req *v1.ItemRequireViewReq) (res *v1.ItemRequireViewRes, err error) {
	_, span := gtrace.NewSpan(ctx, "处理条目查看申请")
	defer span.End()

	ac, err := claims.AuthClaimsFromCtx(ctx)
	if err != nil { // code 61 Token过期或无法解析
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "用户未登录")
	}
	_ = ac

	res = &v1.ItemRequireViewRes{}

	// 先记个数
	toVerifyIds := make([]int, 0, len(req.Requirements))
	for _, req := range req.Requirements {
		toVerifyIds = append(toVerifyIds, req.ItemId)
	}

	var (
		count         int
		verifiedItems []entity.Items
	)
	err = dao.Items.Ctx(ctx).
		Fields("id").
		WhereIn("id", toVerifyIds).
		ScanAndCount(&verifiedItems, &count, true)
	if err != nil {
		span.SetAttributes(
			attribute.String("db.error", err.Error()),
		)
		span.SetStatus(codes.Error, "数据库处理异常")
		return nil, gerror.NewCode(gcode.CodeInternalError, "无法处理用户请求, 请联系管理员")
	}
	res.Failure = len(toVerifyIds) - count

	// 构建插入列表
	insertLists := make(g.List, 0, count)
	for _, item := range verifiedItems {
		insertLists = append(insertLists, g.Map{
			"item_id":   item.Id,
			"member_id": ac.Id,
			"role":      "member",
			"status":    2, // 2: 待审核
		})
	}

	sqlres, err := dao.ItemMembers.Ctx(ctx).
		Data(insertLists).
		OnConflict("item_id", "member_id").
		Save()
	if err != nil {
		span.SetAttributes(
			attribute.String("db.error", err.Error()),
			attribute.String("db.query.integrants", fmt.Sprintf("%v", insertLists)),
		)

		span.SetStatus(codes.Error, "数据库处理异常")
		return nil, gerror.NewCode(gcode.CodeInternalError, "无法处理用户请求, 请联系管理员")
	}
	success, err := sqlres.RowsAffected()
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, "无法统计成功提交的条目个数")
	}
	res.Success = int(success)
	res.Failure = len(req.Requirements) - res.Success
	res.Message = "申请提交成功"

	span.SetStatus(codes.Ok, "请求处理完毕")
	return res, nil
}
