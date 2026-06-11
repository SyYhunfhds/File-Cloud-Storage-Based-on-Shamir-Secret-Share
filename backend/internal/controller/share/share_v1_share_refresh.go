package share

import (
	"backend/internal/claims"
	"backend/internal/dao"
	"backend/internal/logic"
	"backend/pkg/shamir/v3"
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/os/gtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"backend/api/share/v1"
)

// OperationTempValue 用于业务过程中暂存值
type OperationTempValue struct {
	EncodedRecoveryShare []byte
	RecoveryCodeHash     string

	EncodedUserShare []byte // Base64编码的User Share, 用于密钥重建和再分发
	UserShare        shamir.Share

	IsOwner          bool   // 用户是否为条目所有者
	EncodedAuthShare []byte // 用于密钥重建和再分发 (后端持有) // 从数据中取出来后要得到Base64解码后的值
	AuthShare        shamir.Share

	FileKey     []byte   // 文件密钥 (还原得到)
	Usernames   []string // 用于调试 审计成员的用户名
	UserIds     []int    // 用于对照解析份额
	Coordinates []uint32 // 用户坐标, 用于密钥重建和再分发
}

// 参数分别为 用户ID和条目ID
const queryRecoveryCodeHash = `
select
    s.share_base64 as encoded_recovery_share,
    s.code_hash as recovery_code_hash
from public.items
left join public.shares s on items.id = s.item_id
where items.owner_id = ? and item_id = ? and s.share_type = 'recovery'
`

// 参数分别为 用户ID和条目ID
const queryAuthShare = `select
    items.owner_id = ? as is_owner,
    s.share_base64 as encoded_auth_share
from public.items
left join public.shares s on items.id = s.item_id
where item_id = ? and s.share_type = 'auth'`

// 检索用户坐标, 参数分别为当前用户的ID、当前用户ID和条目ID
//
// 分别用于防止水平越权、避免检索到用户自己的坐标以及防止给其他用户分配到未公开项目的份额
const queryUserCoordinates = `select
    array_agg(distinct (username)) as usernames,
    array_agg(users.id) as user_ids, array_agg(share_coor) as coordinates
from public.users
left join public.item_members im on users.id = im.member_id
left join public.items i on im.item_id = i.id
where 
    users.share_coor != 0 
  and i.owner_id = ?
  and im.id != 0 
  and users.id != ? 
  and item_id = ? 
  and im.status = 1
  and i.is_public = true
`

func (c *ControllerV1) ShareRefresh(ctx context.Context, req *v1.ShareRefreshReq) (res *v1.ShareRefreshRes, err error) {
	ctx, span := gtrace.NewSpan(ctx, "ShareRefreshHandler")
	defer span.End()
	request := g.RequestFromCtx(ctx)
	if request == nil {
		span.SetStatus(codes.Error, "无法获取请求对象")
		err = gerror.NewCode(gcode.CodeInternalError, "无法建立流式响应")
		return
	}
	response := request.Response
	c.beginSSE(response) // 开启SSE

	var msg = "解析用户Claims"
	span.AddEvent(msg)
	c.sendMsgAndFlush(response, msg)
	ac, err := claims.AuthClaimsFromCtx(ctx)
	if err != nil {
		msg = "无法获取用户Claims"
		span.SetStatus(codes.Error, msg)
		span.SetAttributes(attribute.String("request.claims.validation.error", err.Error()))
		c.sendMsgAndFlush(response, msg)

		err = gerror.NewCode(gcode.CodeNotAuthorized, "无法确认用户登录状态")
		return
	}
	c.sendAndFlush(response, Msg{Progress: 10, Message: "用户Claims解析成功"})

	res = &v1.ShareRefreshRes{}

	var temp = &OperationTempValue{}
	msg = "校验参数有效性"
	c.sendAndFlush(response, Msg{Progress: 20, Message: "开始校验参数有效性"})
	span.AddEvent(msg)
	c.sendMsgAndFlush(response, msg)
	if req.RecoveryCode != "" {
		span.AddEvent("用户提交了恢复码, 优先使用RecoveryShare进行密钥重建")
		span.AddEvent("从数据库中取出Recovery Code的哈希值")
		_ctx, _span := gtrace.NewSpan(ctx, "从数据库中取出Recovery Code的哈希值和Recovery Share并进行校验")
		err = g.DB().Ctx(_ctx).Raw(queryRecoveryCodeHash, ac.Id, req.ItemId).Scan(&temp)
		if err != nil {
			_span.SetStatus(codes.Error, "从数据库中取出Recovery Code的哈希值时发生异常")
			_span.SetAttributes(attribute.String("db.query.error", err.Error()))

			err = gerror.NewCode(gcode.CodeInternalError, "无法确认恢复码是否有效, 这不是用户的问题")
			_span.End()
			return nil, err
		}

		if ok, err := c.verifyRecoveryCode(req.RecoveryCode, temp.RecoveryCodeHash); !ok {
			_span.SetStatus(codes.Error, "恢复码验证失败")
			_span.SetAttributes(attribute.String("request.recovery_code.validation.error", err.Error()))

			err = gerror.NewCode(gcode.CodeSecurityReason, "用户的Recovery Code不正确")
			_span.End()
			return nil, err
		}
		res.IsRecoveryCodeReGenerated = false // 恢复码有效, 所以不会修改恢复码

		_span.AddEvent("使用Recovery Code对Recovery Share进行解密")
		/*
			temp.EncodedUserShare, err = c.cu.ChainBeforeDecrypt(
					ctx, []byte(req.RecoveryCode), temp.EncodedRecoveryShare,
				)
			if err != nil {
					_span.SetStatus(codes.Error, "Recovery Share解密失败")
					_span.SetAttributes(attribute.String("file.key.recovery_share.decryption.error", err.Error()))

					err = gerror.NewCode(gcode.CodeSecurityReason, "Recovery Share可能已失效, 这不是用户的问题")
					_span.End()
					return
				}
		*/
		temp.EncodedUserShare = temp.EncodedRecoveryShare // 现在RS也是Base64编码的而已
		_span.End()
	}
	if req.DeviceShare == "" && req.RecoveryCode == "" {
		span.SetStatus(codes.Error, "用户未提交Device Share, 无法进行密钥重建和再分发")
		span.SetAttributes(attribute.String("request.device_share.validation.error", "用户未提交Device Share"))

		err = gerror.NewCode(gcode.CodeInvalidRequest, "用户未提交Device Share")
		return
	}
	if req.RecoveryCode == "" && req.DeviceShare != "" {
		span.AddEvent("用户未提交Recovery Code, 但提交了Device Share, 将使用Auth Share与Device Share进行密钥重建")
		temp.EncodedUserShare = []byte(req.DeviceShare)
		req.DeviceShare = "" // 置空字符串

		res.IsRecoveryCodeReGenerated = true // 用户未提供恢复码, 将重新生成恢复码
	}

	msg = "提取份额并进行密钥重建"
	c.sendAndFlush(response, Msg{Progress: 40, Message: "开始提取份额并进行密钥重建"})
	span.AddEvent("从数据库中取出Auth Share")
	_ctx, _span := gtrace.NewSpan(ctx, "从数据库中取出Auth Share")
	err = g.DB().Ctx(_ctx).Raw(queryAuthShare, ac.Id, req.ItemId).Scan(&temp)
	if err != nil {
		span.SetStatus(codes.Error, "从数据库中取出Auth Share时发生异常")
		span.SetAttributes(attribute.String("db.query.error", err.Error()))

		err = gerror.NewCode(gcode.CodeInternalError, "无法确认Auth Share是否有效")
		_span.End()
		return
	} else if !temp.IsOwner {
		_span.SetStatus(codes.Error, "用户不是条目所有者")

		err = gerror.NewCode(gcode.CodeNotAuthorized, "用户不是条目所有者")
		return
	}
	_span.End()

	msg = "查找其他用户的坐标, 用于计算份额"
	span.AddEvent(msg)
	c.sendAndFlush(response, Msg{Progress: 50, Message: msg})
	_ctx, _span = gtrace.NewSpan(ctx, "查找其他用户的坐标用于计算份额")
	err = g.DB().Ctx(_ctx).Raw(queryUserCoordinates, ac.Id, ac.Id, req.ItemId).Scan(&temp)
	if err != nil {
		msg = "无法获取Member的信息, 将不会为其他用户计算份额"
		c.sendAndFlush(response, Msg{Progress: 55, Message: msg})
		_span.SetStatus(codes.Error, "无法获取其他用户的坐标")
		_span.SetAttributes(attribute.String("db.query.error", err.Error()))

	}
	_span.End()
	span.SetAttributes(attribute.StringSlice("key.share.re_split.members", temp.Usernames))

	span.AddEvent("使用Auth Share和Device Share重建密钥")
	temp.EncodedUserShare, _ = gbase64.Decode(temp.EncodedUserShare)
	temp.EncodedAuthShare, _ = gbase64.Decode(temp.EncodedAuthShare)
	temp.FileKey, err = c.cu.RecoverFromJson(ctx,
		temp.EncodedUserShare,
		temp.EncodedAuthShare,
	) // 密钥重建
	if err != nil {
		span.SetStatus(codes.Error, "密钥重建失败")
		span.SetAttributes(attribute.String("file.key.recovery.error", err.Error()))

		c.sendAndFlush(response, Msg{Progress: 55, Message: "密钥重建失败"})
		err = gerror.NewCode(gcode.CodeSecurityReason, "密钥重建失败")
		return
	}
	c.sendAndFlush(response, Msg{Progress: 60, Message: "密钥重建结束"})

	coordinates := make([]uint32, 0, len(temp.Coordinates)+3)
	coordinates = append(coordinates, ac.Coordinate)              // 用户主坐标
	coordinates = append(coordinates, rand.Uint32N(shamir.Prime)) // 用户随机坐标, 用于生成新的Auth Share
	coordinates = append(coordinates, rand.Uint32N(shamir.Prime)) // 用户随机坐标, 用于生成新的Recovery Share
	coordinates = append(coordinates, temp.Coordinates...)        // 其他用户的坐标

	c.sendAndFlush(response, Msg{Progress: 70, Message: "重新计算份额"})
	shares, err := c.cu.SplitToJson(
		ctx, temp.FileKey, coordinates...,
	)
	var (
		ds          = shares[0]
		as          = shares[1]
		rs          = shares[2]
		otherShares = shares[3:]
	)
	defer cleanup(temp.FileKey, otherShares, as, rs)
	if err != nil {
		span.SetStatus(codes.Error, "份额重新计算失败")
		span.SetAttributes(attribute.String("file.key.re_split.error", err.Error()))

		c.sendMsgAndFlush(response, "无法重新计算份额")
		err = gerror.NewCode(gcode.CodeInternalError, "无法重新计算份额")
		return
	}
	/*
		// TODO: 启用份额加密
		span.AddEvent("加密份额并存入数据库")
		encAS, err := c.cu.EncryptAuthShare(ctx, as)
			encRS, code, err := c.cu.EncryptRecoveryShare(rs, nil)
			if err != nil {
				span.SetStatus(codes.Error, "份额加密失败")
				span.SetAttributes(attribute.String("file.key.share.encryption.error", err.Error()))

				c.sendAndFlush(response, Msg{Progress: 75, Message: "份额加密失败"})
				err = gerror.NewCode(gcode.CodeInternalError, "份额加密失败")
				return
			}
	*/
	code := c.cu.StringKey() // Recovery Code
	codeHash, _ := c.hu.HashGen(code)
	span.AddEvent("将Auth Share、Recovery Share以及其他用户的Device Share存入数据库")
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Auth Share与Recovery Share入库
		_ctx, _span = gtrace.NewSpan(_ctx, "将Auth Share和Recovery Share存入数据库")
		_, err = dao.Shares.Ctx(_ctx).Data(g.List{
			g.Map{
				"item_id":      req.ItemId,
				"user_id":      ac.Id,
				"owner_id":     ac.Id,
				"owner":        ac.Username,
				"share_type":   dao.ShareTypeAuth,
				"share_base64": gbase64.EncodeToString(as),
				"status":       dao.ShareStatusActive,

				"updated_at": gtime.Now(),
			},
			g.Map{
				"item_id":      req.ItemId,
				"user_id":      ac.Id,
				"owner_id":     ac.Id,
				"owner":        ac.Username,
				"share_type":   dao.ShareTypeRecovery,
				"share_base64": gbase64.EncodeToString(rs),
				"code_hash":    codeHash,
				"status":       dao.ShareStatusActive,

				"updated_at": gtime.Now(),
			},
		}).OnConflict("item_id", "user_id", "share_type", "status").Save()
		if err != nil {
			return err
		}

		// 其他份额入库
		_span.AddEvent("缓存审计成员的Device Share")
		if len(otherShares) == 0 {
			_span.AddEvent("没有可被备份的Device Share, 提前退出事务")
			return nil
		}

		var updateList = make(g.List, 0, len(otherShares))
		_span.AddEvent(fmt.Sprintf("共有%d位成员的Device Share需要缓存", len(otherShares)))
		for i, share := range otherShares {
			updateList = append(updateList, g.Map{
				"user_id":      temp.UserIds[i],     // 成员ID
				"item_id":      req.ItemId,          // 条目ID
				"owner_id":     ac.Id,               // 条目所有者 (不是 该份额的所有者)ID
				"owner":        ac.Username,         // 条目所有者的用户名
				"share_type":   dao.ShareTypeDevice, // 用户需要尽快登录上来
				"share_base64": gbase64.EncodeToString(share),
				"status":       dao.ShareStatusActive,

				"updated_at": gtime.Now(), // 手动刷新
			})
		}
		_, err = dao.Shares.Ctx(_ctx).
			Data(updateList).
			OnConflict("item_id", "user_id", "share_type", "status").
			Save()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		span.SetStatus(codes.Error, "份额入库失败")
		span.SetAttributes(attribute.String("db.save.trsaction.error", err.Error()))

		c.sendMsgAndFlush(response, "份额入库失败")
		_span.End()
		return
	}
	_span.End()
	c.sendAndFlush(response, Msg{Progress: 90, Message: "份额入库完毕"})

	msg = "请求处理完毕"
	res.DeviceShare = gbase64.EncodeToString(ds)
	res.RecoveryCode = code
	span.SetStatus(codes.Ok, msg)
	span.AddEvent("请求处理成功")
	c.sendAndFlush(response, Msg{Progress: 100, Message: msg, Data: res}) // 响应在这里

	return nil, nil
}

func (c *ControllerV1) verifyRecoveryCode(code string, hash string) (bool, error) {
	return c.hu.HashVerify(code, hash)
}
func cleanup(key []byte, shares [][]byte, others ...[]byte) func() {
	return func() {
		for _, share := range shares {
			logic.Memclr(share)
		}
		for _, share := range others {
			logic.Memclr(share)
		}
		logic.Memclr(key)
	}
}
