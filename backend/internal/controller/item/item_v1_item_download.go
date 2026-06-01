package item

import (
	"backend/api/item/v1"
	"backend/internal/claims"
	"backend/internal/config"
	"backend/internal/dao"
	"backend/internal/logic"
	"backend/pkg/shamir/v3"
	"context"
	"errors"
	"net/http"

	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/util/gvalid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (c *ControllerV1) ItemDownload(ctx context.Context, req *v1.ItemDownloadReq) (res *v1.ItemDownloadRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

// PortableItemDownload
//
// PortableXXX系列API, 用于单独实现需要自定义HTTP行为的API
func PortableItemDownload(ic *config.Item) ghttp.HandlerFunc {
	fu := logic.NewFileUtils()
	fu.BuildWithConfig(ic)
	cu := logic.NewCryptoUtils()
	cu.BuildWithConfig(ic)

	// UserAuthShare
	//
	// 用于暂存SQL查询数据, 包含用于判断(条目)为用户所有的标志字段
	type UserAuthShare struct {
		ItemId   int
		Filename string

		IsMemberValid bool // 如果是被允许查看的Member
		BelongToUser  bool // 如果是所有者

		AuthShare        []byte
		decodedAuthShare shamir.Share
	}

	return func(r *ghttp.Request) {
		ctx, span := gtrace.NewSpan(r.GetCtx(), "File recovery and download")
		defer span.End()

		span.AddEvent("获取用户Token")
		ac, err := claims.AuthClaimsFromCtx(ctx)
		if err != nil {
			span.SetStatus(codes.Error, "用户未登录")
			span.SetAttributes(attribute.String("token.validation.error", err.Error()))

			r.Response.WriteJson(v1.ItemDownloadRes{
				Code:    http.StatusUnauthorized,
				Message: "用户未登录",
			})
			return
		}

		span.AddEvent("解析并校验请求参数")
		req := &v1.ItemDownloadReq{}
		// 解析并校验参数
		if err = r.Parse(&req); err != nil {
			var v gvalid.Error
			if errors.As(err, &v) { // 参数校验不通过
				span.SetStatus(codes.Error, "参数校验不通过")
				span.SetAttributes(attribute.String("request.validation.error", err.Error()))

				r.Response.WriteJson(v1.ItemDownloadRes{
					Code:    http.StatusBadRequest,
					Message: v.String(),
				})
				return
			}

			span.SetStatus(codes.Error, "参数解析失败")
			span.SetAttributes(attribute.String("request.parse.error", err.Error()))
			r.Response.WriteJson(v1.ItemDownloadRes{ // 参数无法解析
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}

		span.SetAttributes(
			attribute.Int("request.parameter.item_id", req.ItemId),
			attribute.String("request.parameter.share.first_4_chars", req.DeviceShare[:4]),
			attribute.Int("request.parameter.share.length", len(req.DeviceShare)),
		)

		// 首先获取ItemId, 并检查是不是所有者
		span.AddEvent("检查用户是否有权下载文件")
		sqlV := UserAuthShare{}
		// 现在还只能下载完全是用户自己拥有的文件
		err = dao.Items.Ctx(ctx).Raw(`
select
    items.id as item_id,
    numeric_eq(owner_id, ?) as belong_to_user,
    filename
    from public.items where id = ?
`, ac.Id, req.ItemId).Scan(&sqlV)
		if err != nil {
			span.SetStatus(codes.Error, "无法确认用户是否为指定文件的Owner")
			span.SetAttributes(attribute.String("db.query.ownship.error", err.Error()))

			r.Response.WriteJson(v1.ItemDownloadRes{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}
		// 如果不是所有者, 再检查是不是member
		if !sqlV.BelongToUser {
			sqlV.IsMemberValid, err = dao.ItemMembers.Ctx(ctx).
				Where("id", sqlV.ItemId).
				Where("member_id", ac.Id).
				Exist()
			if err != nil {
				span.SetStatus(codes.Error, "无法确认用户是否为指定文件的Member")
				span.SetAttributes(attribute.String("db.query.membership.error", err.Error()))

				r.Response.WriteJson(v1.ItemDownloadRes{
					Code:    http.StatusInternalServerError,
					Message: err.Error(),
				})
				return
			}
		}
		span.SetAttributes(
			attribute.Bool("record.is_onwer", sqlV.BelongToUser),
			attribute.Bool("record.is_membership", sqlV.IsMemberValid),
		)
		if !sqlV.BelongToUser && !sqlV.IsMemberValid {
			span.SetStatus(codes.Error, "用户无权查看该文件")

			r.Response.WriteJson(v1.ItemDownloadRes{
				Code:    http.StatusForbidden,
				Message: "您没有权限查看该条目",
			})
			return
		}

		span.AddEvent("检查加密文件是否存在")
		if err = fu.ItemExits(sqlV.Filename); err != nil {
			span.SetStatus(codes.Error, "检测到用户寻找的文件不存在")
			span.SetAttributes(attribute.String("file.extract.error", err.Error()))

			r.Response.WriteJson(v1.ItemDownloadRes{
				Code:    http.StatusNotFound,
				Message: "条目不存在",
			})
			return
		}
		span.SetAttributes(attribute.String("file.encrypted.filesize", gfile.ReadableSize(sqlV.Filename)))

		// 然后查找Shares表获取AuthShare
		span.AddEvent("查找并GCM解密Auth Share")
		err = dao.Shares.Ctx(ctx).Raw(`
select auth_share from public.shares where item_id = ?
`, sqlV.ItemId).Scan(&sqlV)
		if err != nil {
			r.Response.WriteJson(v1.ItemDownloadRes{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}

		span.SetAttributes(
			attribute.String("file.key.auth_share.base64encode", gbase64.EncodeToString(sqlV.AuthShare)),
			attribute.String("file.key.device_share.base64encode", req.DeviceShare), // device share本来就已经是base64编码过一次了
		)

		// spew.Dump(sqlV)
		// 对AuthShare进行解码
		// 先AES-GCM解密再Base64+JSON解码
		sqlV.AuthShare, err = cu.SymmetricDecrypt(sqlV.AuthShare, nil)
		if err != nil {
			span.SetStatus(codes.Error, "检测到服务端Auth份额无法进行GCM解密")
			span.SetAttributes(attribute.String("share.decode.error", err.Error()))

			r.Response.WriteJson(v1.ItemDownloadRes{
				Code:    http.StatusInternalServerError,
				Message: "服务端认证份额已损坏, 请立即联系管理员",
			})
			return
		}
		// spew.Dump(sqlV)
		sqlV.decodedAuthShare, err = shamir.ShareFromBase64Bytes(sqlV.AuthShare)
		if err != nil {
			span.SetStatus(codes.Error, "检测到服务端Auth份额无法还原回标准类型")
			span.SetAttributes(attribute.String("auth_share.decode.error", err.Error()))

			r.Response.WriteJson(v1.ItemDownloadRes{
				Code:    http.StatusInternalServerError,
				Message: "服务端认证份额已损坏, 请立即联系管理员",
			})
			return
		}
		// spew.Dump(sqlV)

		// 然后解密DeviceShare
		span.AddEvent("还原Device Share为标准份额类型")
		deviceShare, err := shamir.ShareFromBase64(req.DeviceShare)
		req.DeviceShare = "" // 解除对原始字符串的引用
		if err != nil {
			span.SetStatus(codes.Error, "检测到用户Device份额无法还原回标准类型")
			span.SetAttributes(attribute.String("device_share.decode.error", err.Error()))

			r.Response.WriteJson(v1.ItemDownloadRes{
				Code:    http.StatusBadRequest,
				Message: "用户的Device份额已失效, 无法还原回标准份额",
			})
			return
		}
		// spew.Dump(deviceShare)
		span.AddEvent("使用Auth Share和Device Share恢复密钥")
		key := shamir.Recover([]shamir.Share{
			sqlV.decodedAuthShare, deviceShare,
		})
		_ = key
		defer logic.Memclr(key) // 置空密钥

		// 读取加密文件、解密并保存到临时目录中
		span.AddEvent("读取加密文件、解密并保存到临时目录中")
		del, err := fu.DecryptAndSaveItem(sqlV.Filename, key)
		defer del() // 删除临时文件
		if err != nil {
			span.SetStatus(codes.Error, "文件解密失败")
			span.SetAttributes(attribute.String("file.decryption.error", err.Error()))

			r.Response.WriteJson(v1.ItemDownloadRes{
				Code:    http.StatusInternalServerError,
				Message: "文件解密失败",
			})
			return
		}

		// 文件下载前会覆盖缓冲区
		// r.Response.ServeFileDownload(gfile.Join("business/item/unlocked", "plain.txt"))
		span.AddEvent("开始传输文件字节")
		err = fu.ItemDownload(r.Response, sqlV.Filename)
		if (r.Response.BufferLength() == 9 && r.Response.BufferString() == "Not Found") || err != nil {
			span.SetStatus(codes.Error, "文件下载失败, 找不到临时解密后的文件")
			span.SetAttributes(attribute.String("file.download.error", err.Error()))

			r.Response.ClearBuffer() // 清空缓冲区
			r.Response.WriteJson(v1.ItemDownloadRes{
				Code:    http.StatusNotFound,
				Message: "无法下载解密后的文件",
			})
			return
		}

		span.SetStatus(codes.Ok, "文件下载请求处理结束")
		r.Response.WriteJson(v1.ItemDownloadRes{
			Code:    http.StatusOK,
			Message: "请求处理完毕",
		})
		return
	}
}
