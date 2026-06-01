package config

type CookieConfig struct {
	Key    string `json:"key" yaml:"key"`
	Value  string `json:"value" yaml:"value"`
	Domain string `json:"domain" yaml:"domain"`
	Path   string `json:"path" yaml:"path"`

	// MaxAge Cookie过期时间, 单位: 秒
	MaxAge   int64 `json:"max_age" yaml:"max_age"`
	HttpOnly bool  `json:"http_only" yaml:"http_only"`

	// JWT 配置

	// Secret JWT签名密钥
	Secret        string `json:"jwt_secret" yaml:"jwt_secret"`
	SigningMethod string `json:"jwt_signing_method" yaml:"jwt_signing_method"`
}
