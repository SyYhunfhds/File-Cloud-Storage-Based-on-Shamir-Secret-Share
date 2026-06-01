package auth

import (
	"backend/internal/config"
	"backend/internal/logic"
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"backend/api/auth/v1"
)

func (c *ControllerV1) Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func PortableLogoutHandler(cc *config.CookieConfig) ghttp.HandlerFunc {
	cu := logic.NewCookieUtils()
	cu.BuildWithConfig(cc)

	return func(r *ghttp.Request) {
		cu.Remove(r.Cookie)

		r.Response.WriteJson(v1.LogoutRes{
			Code:    gcode.CodeOK.Code(),
			Message: "登出成功",
		})
	}
}
