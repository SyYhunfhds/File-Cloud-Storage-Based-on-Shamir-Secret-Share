package auth

import (
	"backend/internal/config"
	"backend/internal/logic"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/os/gctx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func InjectCookieIntoContext(cc *config.CookieConfig) ghttp.HandlerFunc {
	cu := logic.NewCookieUtils()
	cu.BuildWithConfig(cc)

	return func(r *ghttp.Request) {
		// fmt.Println("Cookie")
		// spew.Dump(r.Cookie)
		// fmt.Println("Header")
		// spew.Dump(r.Header.Values("authorization"))
		// spew.Dump(r.Header.Values("cookie"))
		// spew.Dump(r.Header.Values("Authorization"))
		// spew.Dump(r.Header.Values("Cookie"))

		// 在Trace中注入用户信息
		ctx, span := gtrace.NewSpan(r.Context(), "ProtectedPaths") // 受保护路由
		defer span.End()

		tokenStr := cu.GetCookie(r.Cookie)
		if tokenStr == "" {
			// 如果Cookie里没有就去Header里找
			tokenStr = r.GetHeader("Authorization")
			if strings.HasPrefix(tokenStr, "Bearer") { // 为了兼容不同的前端习惯
				tokenStr = tokenStr[7:] // 跳过一个额外的空格或冒号
			}
		}

		// spew.Dump(tokenStr)
		token, valid, err := cu.ParseAndVerify(tokenStr, &jwtAuthClaims{})
		// JWT无效/无法序列化/没有JWT Token
		if err != nil {
			span.SetStatus(codes.Error, "用户JWT Token解析失败")
			span.SetAttributes(attribute.String("token.parse.error", err.Error()))

			r.Response.WriteJson(g.Map{
				"code":    gcode.CodeInvalidRequest.Code(), // code 66
				"message": "请先登录",
			})
			return
		}
		// JWT过期
		if !valid {
			span.SetStatus(codes.Error, "用户非法携带无法校验的Token")
			span.SetAttributes(attribute.Bool("token.expired", true))

			r.Response.WriteJson(g.Map{
				"code":    gcode.CodeNotAuthorized.Code(), // code 61
				"message": "Token无效",
			})
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
		span.SetAttributes( // 这些参数竟然是最后才看到的
			attribute.String("user.username", claims.Username),
			attribute.Int("user.privilege.level", claims.Privilege),
		)
		ctx = gctx.WithSpan(ctx, "放行带有有效Token的请求") // 替换为新span的context
		r.SetCtx(ctx)
		claims.InjectRequestCtx(r)
		// span.End()
		r.Middleware.Next() // 进入下一个路由
	}
}
