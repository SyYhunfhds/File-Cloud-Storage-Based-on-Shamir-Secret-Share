package auth

import (
	"backend/internal/claims"
	"backend/internal/config"
	"backend/internal/dao"
	"backend/internal/logic"
	"backend/internal/model/entity"
	"context"
	"errors"

	"backend/api/auth/v1"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/util/gvalid"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var (
	ErrUsernameAndEmailEmpty = gerror.NewCode(
		gcode.CodeInvalidRequest,
		"用户名和邮箱不能同时为空",
	)

	ErrPasswordMismatch = gerror.NewCode(
		gcode.CodeInvalidRequest,
		"密码错误",
	)
	ErrJustCannotLogin = gerror.NewCode(
		gcode.CodeInternalError,
		"无法登录, 请联系管理员",
	)
)

type jwtAuthClaims struct {
	claims.Auth
	jwt.RegisteredClaims
}

func (j *jwtAuthClaims) GetUserId() int {
	return j.Id
}

func (j *jwtAuthClaims) GetUsername() string {
	return j.Username
}

func (j *jwtAuthClaims) GetUserCoor() uint32 {
	return j.Coordinate
}

func (j *jwtAuthClaims) SetRegisteredClaims(rc jwt.RegisteredClaims) {
	j.RegisteredClaims = rc
}

func (j *jwtAuthClaims) InjectRequestCtx(r *ghttp.Request) {
	r.SetCtx(j.InjectContext(r.GetCtx()))
}
func (j *jwtAuthClaims) FromIncomingCtx(ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = gerror.NewCode(gcode.CodeInternalError)
		}
	}()

	j.Id = ctx.Value("userid").(int)
	j.Username = ctx.Value("username").(string)
	j.Coordinate = ctx.Value("coordinate").(uint32)

	return nil
}

func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	if req.Username == "" && req.Email == "" {
		return nil, ErrUsernameAndEmailEmpty
	}

	if req.Username != "" {
		// 用 用户名+密码登录
		var user = &entity.Users{}
		err = dao.Users.Ctx(ctx).Where("username", req.Username).Scan(user)
		if err != nil {
			return nil, ErrJustCannotLogin
		}

		match, err := c.Utilities.hashVerify(req.Password, user.Password)
		if err != nil {
			return nil, ErrJustCannotLogin
		}
		if match {
			return &v1.LoginRes{
				Message: "登录成功",
			}, nil
		}
		return nil, ErrPasswordMismatch
	}

	if req.Email != "" {
		// 用 用户名+密码登录
		var user = &entity.Users{}
		err = dao.Users.Ctx(ctx).Where("email", req.Email).Scan(user)
		if err != nil {
			return nil, ErrJustCannotLogin
		}

		match, err := c.Utilities.hashVerify(req.Password, user.Password)
		if err != nil {
			return nil, ErrJustCannotLogin
		}
		if match {
			return &v1.LoginRes{
				Message: "登录成功",
			}, nil
		}
		return nil, ErrPasswordMismatch
	}

	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func PortableLoginHandler(ac *config.ArgonConfig, cc *config.CookieConfig) ghttp.HandlerFunc {
	hu := logic.NewHashUtils()
	hu.BuildWithConfig(ac)
	cu := logic.NewCookieUtils()
	cu.BuildWithConfig(cc)

	return func(r *ghttp.Request) {
		ctx, span := gtrace.NewSpan(r.GetCtx(), "LoginHandler")
		defer span.End()

		var req = &v1.LoginReq{}
		span.AddEvent("解析并校验参数")
		if err := r.Parse(req); err != nil {
			var v gvalid.Error
			if errors.As(err, &v) {
				span.SetStatus(codes.Error, "参数校验失败")
				span.SetAttributes(
					attribute.String("request.validation.error", v.Error()),
				)

				r.Response.WriteJson(v1.LoginRes{
					Code:    gcode.CodeInvalidParameter.Code(),
					Message: v.FirstError().Error(),
				})
				return
			}

			span.SetStatus(codes.Error, "参数解析错误")
			span.SetAttributes(
				attribute.String("request.parse.error", err.Error()),
			)
			r.Response.WriteJson(v1.LoginRes{
				Message: err.Error(),
			})
			return
		}

		span.SetAttributes(
			// 记录请求信息
			attribute.String("request.login.parameter.username", req.Username),
			attribute.String("request.login.parameter.email", req.Email),
		)

		if req.Username == "" && req.Email == "" {
			span.SetStatus(codes.Error, "无效请求")
			span.SetAttributes(
				attribute.String("request.validation.error", "用户名和邮箱不能同时为空"),
			)

			r.Response.WriteJson(v1.LoginRes{
				Code:    gcode.CodeMissingParameter.Code(),
				Message: ErrUsernameAndEmailEmpty.Error(),
			})
			return
		}

		var user = &entity.Users{}
		if req.Username != "" { // 如果用户名不为空就用用户名查询账号信息
			span.AddEvent("通过用户名查询用户信息")

			// 用 用户名+密码登录
			err := dao.Users.Ctx(ctx).Where("username", req.Username).Scan(user)
			if err != nil {
				span.SetStatus(codes.Error, "无法通过用户名搜索用户信息")
				span.SetAttributes(
					attribute.String("db.query.error", err.Error()),
				)

				r.Response.WriteJson(v1.LoginRes{
					Code:    gcode.CodeNotFound.Code(),
					Message: "用户名不存在",
				})
				return
			}
		}

		if req.Username == "" && req.Email != "" { // 仅当用户名为空而邮箱不为空时用邮箱查询账号信息
			span.AddEvent("通过邮箱查询用户信息")

			// 用 用户名+密码登录
			err := dao.Users.Ctx(ctx).Where("email", req.Email).Scan(user)
			if err != nil {
				span.SetStatus(codes.Error, "无法通过邮箱搜索用户信息")
				span.SetAttributes(
					attribute.String("db.query.error", err.Error()),
				)

				r.Response.WriteJson(v1.LoginRes{
					Code:    gcode.CodeNotFound.Code(),
					Message: "邮箱不存在",
				})
				return
			}
		}

		span.AddEvent("校验用户密码")
		// 进行哈希校验
		match, err := hu.HashVerify(req.Password, user.Password)
		if err != nil {
			span.SetStatus(codes.Error, "无法校验密码哈希")
			span.SetAttributes(
				attribute.String("hash(argon).validation.error", err.Error()),
			)

			r.Response.WriteJson(v1.LoginRes{ // hash校验函数失败
				Code:    gcode.CodeInternalError.Code(),
				Message: "密码错误",
			})
			return
		}
		if !match {
			span.SetStatus(codes.Error, "密码不正确")

			r.Response.WriteJson(v1.LoginRes{
				Code:    gcode.CodeNotAuthorized.Code(),
				Message: "密码错误",
			})
			return
		}

		span.AddEvent("检索用户的员工信息")
		var employee = &entity.Employees{}
		err = dao.Employees.Ctx(ctx).Fields("user_id", "job_id").Where("user_id", user.Id).Scan(employee)
		if err != nil {
			span.SetStatus(codes.Error, "无法获取员工信息")
			span.SetAttributes(attribute.String("db.query.error", err.Error()))

			r.Response.WriteJson(v1.LoginRes{
				Code:    gcode.CodeInternalError.Code(),
				Message: "无法获取员工信息",
			})
			return
		}

		span.AddEvent("检索用户的职位权限")
		var job = &entity.Jobs{}
		err = dao.Jobs.Ctx(ctx).Fields("role_privilege").Where("id", employee.JobId).Scan(job)
		if err != nil {
			span.SetStatus(codes.Error, "无法获取员工的职位信息, 因此无法获取权限信息")
			span.SetAttributes(attribute.String("db.query.error", err.Error()))

			r.Response.WriteJson(v1.LoginRes{
				Code:    gcode.CodeInternalError.Code(),
				Message: "无法获取员工职位信息",
			})
			return
		}
		// spew.Dump(job)

		// 校验成功
		span.AddEvent("生成Token")
		// 生成token
		authClaims := jwtAuthClaims{
			Auth: claims.Auth{
				Id:         user.Id,
				Username:   user.Username,
				Privilege:  job.RolePrivilege,
				Coordinate: uint32(user.ShareCoor),
			},
		}
		token := cu.NewWithClaims(&authClaims)
		signedString, err := cu.GetSignedString(token)
		if err != nil { // Token生成失败
			span.SetStatus(codes.Error, "无法生成token")
			span.SetAttributes(
				attribute.String("token.generation.error", err.Error()),
			)

			r.Response.WriteJson(v1.LoginRes{
				Code:    gcode.CodeInternalError.Code(),
				Message: "无法生成token",
			})
			return
		}

		span.SetStatus(codes.Ok, "登录成功")
		cu.SetCookie(r.Cookie, signedString)
		r.Response.WriteJson(v1.LoginRes{
			Code:    gcode.CodeOK.Code(),
			Message: "登录成功",
			Data:    signedString, // 适配桌面端
		})
		return
	}
}
