package logic

import (
	"backend/internal/config"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v5"
)

type nopClaim struct {
	jwt.RegisteredClaims
}

type WrapperClaims interface {
	jwt.Claims // 带上原本的Claims接口

	GetUserId() int
	GetUsername() string
	GetUserCoor() uint32

	SetRegisteredClaims(rc jwt.RegisteredClaims) // 允许动态修改JWT签名载荷, Handler调用时完全不需要知道JWT密钥
}

type CookieUtils struct {
	// JWT配置
	secret        []byte
	signingMethod jwt.SigningMethod
	expiresAt     *jwt.NumericDate

	// Cookie配置
	key      string
	domain   string
	path     string
	maxAge   time.Duration
	httpOnly bool
}

func (c *CookieUtils) NewWithClaims(claims WrapperClaims, opts ...jwt.TokenOption) *jwt.Token {
	claims.SetRegisteredClaims(jwt.RegisteredClaims{
		ExpiresAt: c.expiresAt,
	})
	return jwt.NewWithClaims(c.signingMethod, claims, opts...)
}
func (c *CookieUtils) GetSignedString(token *jwt.Token) (string, error) {
	return token.SignedString(c.secret)
}

// ParseAndVerify
//
// 想要啥类型自己传指针进去给jwt库识别
func (c *CookieUtils) ParseAndVerify(tokenStr string, claimType ...jwt.Claims) (token *jwt.Token, valid bool, err error) {
	var claim jwt.Claims
	if len(claimType) > 0 {
		claim = claimType[0]
	} else {
		claim = &nopClaim{}
	}

	token, err = jwt.ParseWithClaims(tokenStr, claim, func(token *jwt.Token) (interface{}, error) {
		return c.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, false, err
	}

	return token, true, nil
}

func (c *CookieUtils) SetCookie(r *ghttp.Cookie, value string) {
	r.SetCookie(c.key, value, c.domain, c.path, c.maxAge)
}
func (c *CookieUtils) GetCookie(r *ghttp.Cookie) string {
	return r.Get(c.key).String()
}
func (c *CookieUtils) Remove(r *ghttp.Cookie) {
	r.RemoveCookie(c.key, c.domain, c.path)
}

var str2SigningMethod = map[string]jwt.SigningMethod{
	"HS256": jwt.SigningMethodHS256,
	"HS384": jwt.SigningMethodHS384,
	"HS512": jwt.SigningMethodHS512,
	"RS256": jwt.SigningMethodRS256,
	"RS384": jwt.SigningMethodRS384,
	"RS512": jwt.SigningMethodRS512,
	"ES256": jwt.SigningMethodES256,
	"ES384": jwt.SigningMethodES384,
	"ES512": jwt.SigningMethodES512,
	"PS256": jwt.SigningMethodPS256,
	"PS384": jwt.SigningMethodPS384,
	"PS512": jwt.SigningMethodPS512,
	"EdDSA": jwt.SigningMethodEdDSA,
	"none":  jwt.SigningMethodNone,
}

func NewCookieUtils() *CookieUtils {
	return &CookieUtils{
		// signingMethod: jwt.SigningMethodHS256,
	}
}
func (c *CookieUtils) BuildWithConfig(cfg *config.CookieConfig) {
	// 设置Cookie配置
	{
		c.key = cfg.Key
		c.domain = cfg.Domain
		c.path = cfg.Path
		c.httpOnly = cfg.HttpOnly
		c.maxAge = time.Duration(cfg.MaxAge) * time.Second
	}

	// 设置JWT配置
	{
		// 设置密钥
		c.secret = []byte(cfg.Secret)
		// 设置签名方法
		if method, exists := str2SigningMethod[cfg.SigningMethod]; exists {
			c.signingMethod = method
		} else {
			c.signingMethod = jwt.SigningMethodHS256
		}
		c.expiresAt = jwt.NewNumericDate(time.Now().Add(c.maxAge))
	}
}
