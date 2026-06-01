package auth

import (
	"backend/internal/config"
	"backend/internal/logic"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gmode"
)

func PortableCookiePrintHandler(cc *config.CookieConfig) ghttp.HandlerFunc {
	cu := logic.NewCookieUtils()
	cu.BuildWithConfig(cc)

	return func(r *ghttp.Request) {
		if gmode.Mode() != gmode.DEVELOP {
			return
		}

		tokenStr := cu.GetCookie(r.Cookie)
		token, valid, err1 := cu.ParseAndVerify(tokenStr)
		claims := &jwtAuthClaims{}
		err2 := claims.FromIncomingCtx(r.GetCtx())

		r.Response.WriteJson(g.Map{
			"jwt":                 tokenStr,
			"claims":              token,
			"claims-from-context": claims,
			"jwt-valid":           valid,
			"err":                 []error{err1, err2},
		})
	}
}
