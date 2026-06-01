package item

import (
	"backend/internal/claims"
	"backend/internal/config"
	"backend/internal/dao"
	"backend/internal/logic"
	"backend/internal/model/do"
	"context"

	"backend/api/item/v1"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/os/gfile"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (c *ControllerV1) ItemSubmit(ctx context.Context, req *v1.ItemSubmitReq) (res *v1.ItemSubmitRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func PortableItemSubmit(ic *config.Item, ac *config.ArgonConfig) ghttp.HandlerFunc {
	fu := logic.NewFileUtils()
	fu.BuildWithConfig(ic)
	crypU := logic.NewCryptoUtils()
	crypU.BuildWithConfig(ic)
	hu := logic.NewHashUtils()
	hu.BuildWithConfig(ac)

	path := "business/item/encrypted" // 会被闭包函数捕获, 逃逸到堆上
	var itemId int64

	return func(r *ghttp.Request) {
		ctx, span := gtrace.NewSpan(r.GetCtx(), "ItemUploadHandler")
		defer span.End()

		span.AddEvent("解析用户Claims")
		ac, err := claims.AuthClaimsFromCtx(ctx)
		if err != nil {
			span.SetStatus(codes.Error, "无法获取用户Token")
			span.SetAttributes(attribute.String("request.validation.error", err.Error()))

			r.Response.WriteJson(v1.ItemSubmitRes{
				Code:    gcode.CodeNotAuthorized.Code(),
				Message: "无法读取用户信息",
				Cause:   err.Error(),
			})
			return
		}
		// TODO: 移除高危Span属性 (用户份额坐标)
		span.SetAttributes(attribute.Int64("request.user.coordinate", int64(ac.Coordinate)))

		span.AddEvent("从请求中获取上传文件")
		file := r.GetUploadFile("item")
		// 分割份额
		// 需要拿到用户坐标
		span.SetAttributes(
			attribute.String("request.file.name", file.Filename),
			attribute.String("request.file.size", gfile.FormatSize(file.Size)),
		)

		// 读取并加密保存文件
		span.AddEvent("读取并加密保存文件")
		ciphertext, key, name, err := fu.EncryptAndSaveFile(file, path)
		span.SetAttributes(
			attribute.String("file.encryption.filename", name),
			attribute.String("file.encryption.filesize", gfile.FormatSize(int64(len(ciphertext)))),
		)
		logic.Memclr(ciphertext) // 主动置空密文
		if err != nil {
			span.SetStatus(codes.Error, "文件加密失败, 已置空密钥, 且没有保存文件")
			span.AddEvent("已置空密钥, 且未保存文件")
			span.SetAttributes(attribute.String("file.encryption.error", err.Error()))

			logic.Memclr(key) // 密钥清空
			r.Response.WriteJson(v1.ItemSubmitRes{
				Code:    gcode.CodeInternalError.Code(),
				Message: "文件保存失败",
			})
			return
		}
		// TODO: 移除高危Span属性 (文件加密密钥原始值)
		span.SetAttributes(attribute.String("file.encryption.key.base64encode", gbase64.EncodeToString(key)))

		// 调试用途
		// 立即解密文件
		/*
			glog.Debugf(gctx.New(), "加密文件时的密钥为: %x", key)
				_, err = fu.DecryptBytes(ciphertext, key, false)
				if err != nil {
					glog.Errorf(gctx.New(), "无法立即解密文件: %v", err)
					logic.Memclr(ciphertext)
				}
		*/

		// 生成份额
		span.AddEvent("对随机生成的密钥进行份额切割计算")
		deviceShare, authShare, recoveryShare, err := crypU.SplitShare(key, ac.Coordinate)
		logic.Memclr(key) // 密钥清空
		// 对称加密认证份额
		encryptedAuthShare, err := crypU.SymmetricEncrypt(authShare, nil, true)
		// 随机密钥加密恢复份额, 并返回密文和32字节密钥
		encryptedRecovery, recoveryCode, err := crypU.SymEncryptWithRandKey(recoveryShare)
		recoveryCodeHash, err := hu.HashGen(recoveryCode)
		if err != nil { // 份额生成失败
			_ = fu.Delete(file, path) // 删除上传的文件
			// logic.Memclr(encryptedAuthShare) // 置空加密的AuthShare
			// logic.Memclr(encryptedRecovery) // 置空加密的RecoveryShare
			span.SetStatus(codes.Error, "密钥切割失败")
			span.AddEvent("份额生成失败, 已删除加密保存的文件, 并置空密钥")

			r.Response.WriteJson(v1.ItemSubmitRes{
				Code:    gcode.CodeInternalError.Code(),
				Message: "密钥切割失败",
				Cause:   err.Error(),
			})
			return
		}
		// TODO: 移除高危Span属性 (份额原始值)
		span.SetAttributes(
			attribute.String("file.key.device_share.base64encode", deviceShare),
			attribute.String("file.key.auth_share.base64encode", gbase64.EncodeToString(encryptedAuthShare)),
			attribute.String("file.key.recovery_share.base64encode", gbase64.EncodeToString(encryptedRecovery)),
		)

		// 数据库存档
		err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			itemId, err = dao.Items.Ctx(ctx).Data(do.Items{
				Filename:   name,
				OwnerId:    ac.Id,
				UploaderId: ac.Id,

				MinimumPrivilege: ac.Privilege,
				IsPublic:         false, // 默认不可被公开搜索
			}).InsertAndGetId()
			if err != nil {
				return err
			}

			_, err = dao.Shares.Ctx(ctx).Data(do.Shares{
				ItemId: itemId,

				AuthShare:        encryptedAuthShare,
				RecoveryShare:    encryptedRecovery,
				RecoveryCodeHash: recoveryCodeHash,
			}).Insert()
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			_ = fu.Delete(file, path) // 删除上传的文件
			span.SetStatus(codes.Error, "份额或条目存档失败")
			span.SetAttributes(attribute.String("db.insert.error", err.Error()))

			r.Response.WriteJson(v1.ItemSubmitRes{
				Code:    gcode.CodeInternalError.Code(),
				Message: "存档失败",
				Cause:   err.Error(),
			})
			return
		}
		res := v1.ItemSubmitRes{
			Code:    gcode.CodeOK.Code(),
			Message: "上传成功",
		}

		res.Data = struct {
			ItemId       int    `json:"item_id"`
			Name         string `json:"name" dc:"上传后的文件名; 可能会因为存在同名文件而被重命名"`
			Share        string `json:"share" dc:"Base64编码的明文Device Share"`
			RecoveryCode string `json:"recovery_code" dc:"Recovery Share加密时用的随机32位可打印字符密钥"`
		}{ItemId: int(itemId), Name: name, Share: deviceShare, RecoveryCode: recoveryCode}

		span.AddEvent("条目上传成功")
		span.SetStatus(codes.Ok, "条目上传成功")
		r.Response.WriteJson(res)
		return
	}
}
