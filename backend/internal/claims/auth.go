package claims

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

type Auth struct {
	Id       int
	Username string
	// 用户权限
	Privilege int
	// 用户独一无二的份额坐标, coor = maphash(user.id)
	Coordinate uint32
}

func (auth *Auth) InjectContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, "userid", auth.Id)
	ctx = context.WithValue(ctx, "username", auth.Username)
	ctx = context.WithValue(ctx, "coordinate", auth.Coordinate)
	ctx = context.WithValue(ctx, "privilege", auth.Privilege)
	return ctx
}
func AuthClaimsFromCtx(ctx context.Context) (claims Auth, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = gerror.NewCode(gcode.CodeInvalidRequest)
		}
	}()

	claims = Auth{}
	{
		claims.Id = ctx.Value("userid").(int)
		claims.Username = ctx.Value("username").(string)
		claims.Coordinate = ctx.Value("coordinate").(uint32)
		claims.Privilege = ctx.Value("privilege").(int)
	}

	return
}
