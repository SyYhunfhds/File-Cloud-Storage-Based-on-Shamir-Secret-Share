package share

import (
	"backend/internal/claims"
	"backend/internal/dao"
	"backend/internal/logic"
	"backend/pkg/shamir/v3"
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/os/gctx"
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
	UserIds     []int    // 用于对照解析份额
	Coordinates []uint32 // 用户坐标, 用于密钥重建和再分发
}

// 参数分别为 用户ID和条目ID
const queryRecoveryCodeHash = `
select
    s.recovery_share as encoded_recovery_share,
    s.recovery_code_hash
from public.items
left join public.shares s on items.id = s.item_id
where owner_id = ? and item_id = ?
`

// 参数分别为 用户ID和条目ID
const queryAuthShare = `select
    items.owner_id = ? as is_owner,
    s.auth_share as encoded_auth_share
from public.items
left join public.shares s on items.id = s.item_id
where item_id = ?`

// 检索用户坐标, 参数分别为当前用户的ID、当前用户ID和条目ID
//
// 分别用于防止水平越权、避免检索到用户自己的坐标以及防止给其他用户分配到未公开项目的份额
const queryUserCoordinates = `select
    distinct (username), users.id as user_ids, share_coor as coordinates
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
		temp.EncodedUserShare, err = c.cu.DecryptRecoveryShare(temp.EncodedRecoveryShare, []byte(req.RecoveryCode), true)
		if err != nil {
			_span.SetStatus(codes.Error, "Recovery Share解密失败")
			_span.SetAttributes(attribute.String("file.key.recovery_share.decryption.error", err.Error()))

			err = gerror.NewCode(gcode.CodeSecurityReason, "Recovery Share可能已失效, 这不是用户的问题")
			_span.End()
			return
		}
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

	/*
		span.AddEvent("解密Auth Share")
			temp.EncodedAuthShare, err = c.cu.DecryptAuthShare(gctx.New(), temp.EncodedAuthShare, true) // 使用服务器主密钥解密
			if err != nil {
				span.SetStatus(codes.Error, "Auth Share解密失败")
				span.SetAttributes(attribute.String("file.key.auth_share.decryption.error", err.Error()))

				err = gerror.NewCode(gcode.CodeSecurityReason, "Auth Share可能已失效, 请联系管理员")
				c.sendAndFlush(response, Msg{Progress: 45, Message: "Auth Share可能已失效"})
				return
			}
	*/

	// 省略部分逻辑
	span.AddEvent("反序列化Device Share")
	temp.UserShare, err = shamir.ShareFromBase64Bytes(temp.EncodedUserShare)
	if err != nil {
		span.SetStatus(codes.Error, "Device Share反序列化失败")
		span.SetAttributes(attribute.String("file.key.device_share.deserialization.error", err.Error()))

		err = gerror.NewCode(gcode.CodeInvalidRequest, "Device Share无效")
		return
	}
	logic.Memclr(temp.EncodedUserShare) // 置空用户份额字节
	span.AddEvent("反序列化Auth Share")
	temp.AuthShare, err = shamir.ShareFromBase64Bytes(temp.EncodedAuthShare)
	if err != nil {
		span.SetStatus(codes.Error, "Auth Share反序列化失败")
		span.SetAttributes(attribute.String("file.key.auth_share.deserialization.error", err.Error()))

		err = gerror.NewCode(gcode.CodeInvalidRequest, "Auth Share无效")
		return
	}
	logic.Memclr(temp.EncodedAuthShare) // 置空授权份额字节

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

		// err = gerror.NewCode(gcode.CodeInternalError, "无法获取其他用户的信息")
		// _span.End()
		// return
	}
	_span.End()

	span.AddEvent("使用Auth Share和Device Share重建密钥")
	// 这里与itemDownload的处理不同, 这里有Unpad环节
	temp.FileKey = c.cu.RecoverShare(temp.AuthShare, temp.UserShare) // 密钥重建
	c.sendAndFlush(response, Msg{Progress: 60, Message: "密钥重建结束"})

	c.sendAndFlush(response, Msg{Progress: 70, Message: "重新计算份额"})
	ds, as, rs, otherShares, err := c.cu.ResplitShare(temp.FileKey, ac.Coordinate, temp.Coordinates)
	defer cleanup(temp.FileKey, otherShares, as, rs)
	if err != nil {
		span.SetStatus(codes.Error, "份额重建失败")
		span.SetAttributes(attribute.String("file.key.re_split.error", err.Error()))

		c.sendMsgAndFlush(response, "无法重新计算份额")
		err = gerror.NewCode(gcode.CodeInternalError, "无法重新计算份额")
		return
	}
	span.AddEvent("加密份额并存入数据库")
	encAS, err := c.cu.EncryptAuthShare(gctx.New(), as)
	encRS, code, err := c.cu.EncryptRecoveryShare(rs, nil)
	if err != nil {
		span.SetStatus(codes.Error, "份额加密失败")
		span.SetAttributes(attribute.String("file.key.share.encryption.error", err.Error()))

		c.sendAndFlush(response, Msg{Progress: 75, Message: "份额加密失败"})
		err = gerror.NewCode(gcode.CodeInternalError, "份额加密失败")
		return
	}
	codeHash, _ := c.hu.HashGen(code)
	_ctx, _span = gtrace.NewSpan(ctx, "将Auth Share和Recovery Share存入数据库")
	err = g.DB().Ctx(_ctx).Transaction(_ctx, func(ctx context.Context, tx gdb.TX) error {
		// Auth Share与Recovery Share入库
		_, err = dao.Shares.Ctx(_ctx).Data(g.Map{
			"item_id":            req.ItemId,
			"user_id":            ac.Id,
			"share_type":         "auth",
			"auth_share":         encAS,
			"recovery_share":     encRS,
			"recovery_code_hash": codeHash,

			"updated_at": gtime.Now(),
		}).OnConflict("item_id", "user_id", "share_type").Save()
		if err != nil {
			return err
		}

		// 其他份额入库
		if len(otherShares) == 0 {
			return nil
		}

		var updateList = make(g.List, 0, len(otherShares))
		for i, share := range otherShares {
			updateList = append(updateList, g.Map{
				"user_id":      temp.UserIds[i],
				"item_id":      req.ItemId,
				"share_type":   "device",
				"share_base64": gbase64.EncodeToString(share),

				"updated_at": gtime.Now(),
			})
		}
		_, err = dao.Shares.Ctx(_ctx).Data(updateList).Save()
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
	res.DeviceShare = ds
	res.RecoveryCode = code
	res.RecoveryShare = gbase64.EncodeToString(encRS)
	span.SetStatus(codes.Ok, msg)
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
