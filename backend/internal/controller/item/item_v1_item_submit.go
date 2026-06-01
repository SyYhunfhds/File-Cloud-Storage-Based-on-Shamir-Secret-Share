package item

import (
	"backend/internal/claims"
	"backend/internal/config"
	"backend/internal/dao"
	"backend/internal/logic"
	"backend/internal/model/do"
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"backend/api/item/v1"
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

	return func(r *ghttp.Request) {
		file := r.GetUploadFile("item")

		// 分割份额
		// 需要拿到用户坐标
		ac, err := claims.AuthClaimsFromCtx(r.GetCtx())
		if err != nil {
			r.Response.WriteJson(v1.ItemSubmitRes{
				Code:    gcode.CodeNotAuthorized.Code(),
				Message: "无法读取用户信息",
				Cause:   err.Error(),
			})
			return
		}

		// 读取并加密保存文件
		key, name, err := fu.EncryptAndSaveFile(file, path)
		if err != nil {
			logic.Memclr(key) // 密钥清空
			r.Response.WriteJson(v1.ItemSubmitRes{
				Code:    gcode.CodeInternalError.Code(),
				Message: "文件保存失败",
				Cause:   err.Error(),
			})
			return
		}

		// 生成份额
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

			r.Response.WriteJson(v1.ItemSubmitRes{
				Code:    gcode.CodeInternalError.Code(),
				Message: "密钥切割失败",
				Cause:   err.Error(),
			})
			return
		}

		// 数据库存档
		err = g.DB().Transaction(r.GetCtx(), func(ctx context.Context, tx gdb.TX) error {
			itemId, err := dao.Items.Ctx(ctx).Data(do.Items{
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
		res.Data.Name = name
		res.Data.Share = deviceShare
		res.Data.RecoveryCode = recoveryCode

		r.Response.WriteJson(res)
		return
	}
}
