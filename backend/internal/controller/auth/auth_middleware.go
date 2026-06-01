package auth

import (
	"backend/internal/config"
	"backend/internal/logic"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/gtrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func InjectCookieIntoContext(cc *config.CookieConfig) ghttp.HandlerFunc {
	cu := logic.NewCookieUtils()
	cu.BuildWithConfig(cc)

	return func(r *ghttp.Request) {
		// 在Trace中注入用户信息
		ctx, span := gtrace.NewSpan(r.Context(), "ProtectedPaths") // 受保护路由
		defer span.End()

		tokenStr := cu.GetCookie(r.Cookie)
		// spew.Dump(tokenStr)
		token, valid, err := cu.ParseAndVerify(tokenStr, &jwtAuthClaims{})
		// JWT无效/无法序列化/没有JWT Token
		if err != nil {
			r.Response.WriteJson(g.Map{
				"code":    gcode.CodeInvalidRequest.Code(), // code 66
				"message": "请先登录",
			})
			// span.SetStatus(codes.Error, "用户未登录, 未携带Token或Token已过期")
			return
		}
		// JWT过期
		if !valid {
			r.Response.WriteJson(g.Map{
				"code":    gcode.CodeNotAuthorized.Code(), // code 61
				"message": "Token无效",
			})
			span.SetStatus(codes.Error, "用户非法携带无法校验的Token")
			return
		}
		// JWT类型不是AuthClaims (比如拿了其他地方的jwt过来)
		claims, ok := token.Claims.(*jwtAuthClaims)
		if !ok {
			r.Response.WriteJson(g.Map{
				"code":    gcode.CodeInvalidOperation.Code(), // code 55
				"message": "请携带正确的Token类型",
			})
			span.SetStatus(codes.Error, "用户非法携带类型错误的Token")
			return
		}
		// spew.Dump(claims)

		// 注入Context
		r.SetCtx(ctx)
		claims.InjectRequestCtx(r)
		span.SetAttributes( // 这些参数竟然是最后才看到的
			attribute.String("user.username", claims.Username),
			attribute.Int("user.privilege.level", claims.Privilege),
		)
		// span.End()
		r.Middleware.Next() // 进入下一个路由
	}
}
