package item

import (
	"backend/internal/claims"
	"backend/internal/dao"
	"backend/internal/logic"
	"backend/internal/model/do"
	"context"

	v1 "backend/api/item/v1"

	"github.com/gogf/gf/v2/database/gdb"
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
	var itemId int64

	r := ghttp.RequestFromCtx(ctx)
	ctx, span := gtrace.NewSpan(ctx, "ItemUploadHandler")
	defer span.End()
	if r == nil {
		span.SetStatus(codes.Error, "无法从上下文中获取请求对象")
		err = gerror.NewCode(gcode.CodeInternalError, "无法获取请求对象")
		return
	}

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
	// 移除高危Span属性 (用户份额坐标)
	// span.SetAttributes(attribute.Int64("request.user.coordinate", int64(ac.Coordinate)))

	span.AddEvent("从请求中获取上传文件")
	file := r.GetUploadFile("item")
	if file == nil {
		span.SetStatus(codes.Error, "无法获取上传文件")
		err = gerror.NewCode(gcode.CodeInternalError, "无法获取上传文件")

		return
	}
	// 分割份额
	// 需要拿到用户坐标
	span.SetAttributes(
		attribute.String("request.file.name", file.Filename),
		attribute.String("request.file.size", gfile.FormatSize(file.Size)),
	)

	// 读取并加密保存文件
	span.AddEvent("读取并加密保存文件")
	ciphertext, key, filepath, err := c.fu.EncryptAndSaveFile(file)
	span.SetAttributes(
		attribute.String("file.encryption.filename", gfile.Basename(filepath)),
		attribute.String("file.encryption.filepath", filepath),
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

	// 生成份额
	span.AddEvent("对随机生成的密钥进行份额切割计算")
	// 返回份额切片和恢复码 (已Base64编码)
	shares, code, err := c.cu.SplitForOneUser(key, 2, 3) // TODO 更换为动态门限对
	var (
		as = shares[0]
		ds = shares[1]
		rs = shares[2]
	)
	logic.Memclr(key) // 密钥清空
	// 对称加密认证份额

	// 计算恢复码的哈希; 现在恢复码是16*3/4长的字符串了
	recoveryCodeHash, _ := c.hu.HashGen(code)

	// 数据库存档
	var (
		filename string
	)
	if req.Filename != "" {
		filename = req.Filename
	} else {
		filename = file.Filename
	} // 获取请求参数

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		itemId, err = dao.Items.Ctx(ctx).Data(do.Items{
			Filename:   filename,
			Savename:   gfile.Basename(filepath),
			OwnerId:    ac.Id,
			UploaderId: ac.Id,

			MinimumPrivilege: ac.Privilege,
			IsPublic:         req.EnablePublic,
		}).InsertAndGetId()
		if err != nil {
			return err
		}

		_, err = dao.Shares.Ctx(ctx).Data(g.List{
			g.Map{
				"item_id":    itemId,
				"user_id":    ac.Id,
				"owner_id":   ac.Id,
				"owner":      ac.Username,
				"share_type": dao.ShareTypeAuth,
				"status":     dao.ShareStatusActive,

				"share_base64": string(as),
			},
			g.Map{
				"item_id":    itemId,
				"user_id":    ac.Id,
				"owner_id":   ac.Id,
				"owner":      ac.Username,
				"share_type": dao.ShareTypeRecovery,
				"status":     dao.ShareStatusActive,

				"share_base64": string(rs),
				"code_hash":    recoveryCodeHash,
			},
		}).OnConflict("item_id", "user_id", "share_type", "status").Save()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		_ = gfile.RemoveFile(filepath) // 删除上传的文件
		span.SetStatus(codes.Error, "份额或条目存档失败")
		span.SetAttributes(attribute.String("db.insert.error", err.Error()))

		r.Response.WriteJson(v1.ItemSubmitRes{
			Code:    gcode.CodeInternalError.Code(),
			Message: "存档失败",
			Cause:   err.Error(),
		})
		return
	}
	res = &v1.ItemSubmitRes{
		Code:    gcode.CodeOK.Code(),
		Message: "上传成功",
	}

	res.Data = struct {
		ItemId       int    `json:"item_id"`
		Name         string `json:"name" dc:"上传后的文件名; 可能会因为存在同名文件而被重命名"`
		Share        string `json:"share" dc:"Base64编码的明文Device Share"`
		RecoveryCode string `json:"recovery_code" dc:"Recovery Share加密时用的随机32位可打印字符密钥"`
	}{ItemId: int(itemId), Name: file.Filename, Share: string(ds), RecoveryCode: code}

	span.AddEvent("条目上传成功")
	span.SetStatus(codes.Ok, "条目上传成功")
	r.Response.WriteJson(res)

	return nil, nil
}

func saveFileBackground(ctx context.Context) {}
