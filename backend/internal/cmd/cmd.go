package cmd

import (
	"backend/internal/config"
	"backend/internal/controller/auth"
	filectrl "backend/internal/controller/file"
	"backend/internal/controller/hello"
	itemctrl "backend/internal/controller/item"
	userctrl "backend/internal/controller/user"
	"backend/internal/controller/version"
	watcherv2 "backend/internal/module/watcher/v2"
	"context"

	"github.com/gogf/gf/contrib/trace/otlphttp/v2"
	_ "github.com/gogf/gf/contrib/trace/otlphttp/v2"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
)

var swaggerTemplate = `
<!DOCTYPE html>
    <html>
    <head>
        <title>{SwaggerUIName}</title>
        <meta charset="utf-8"/>
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <link rel="stylesheet" type="text/css" href="/swagger-res/swagger-ui.css" >
        <style>html{box-sizing:border-box;overflow-y:scroll}*,*:after,*:before{box-sizing:inherit}body{margin:0;background:#fff}</style>
    </head>
    <body>
        <div id="swagger-ui"></div>
        <script src="/swagger-res/swagger-ui-bundle.js"></script>
        <script src="/swagger-res/swagger-ui-standalone-preset.js"></script>
        <script>
        window.onload = function() {
            window.ui = SwaggerUIBundle({
                url: "{SwaggerUIURL}",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                layout: "BaseLayout"
            });
        };
        </script>
    </body>
    </html>
`

const (
	// TODO: 迁移到配置文件中
	serviceName = "crypto-course-design"
	endpoint    = "localhost:4318"
	path        = "/v1/traces"
)

type MainAppConfig struct {
	Item   config.Item          `yaml:"item" json:"item"`
	Server config.ServerConfig  `yaml:"server" json:"server"`
	About  config.AboutSoftware `yaml:"about" json:"about"`
}

var mainConfig MainAppConfig
var mainLoader = gcfg.NewLoader[MainAppConfig](g.Cfg(), "")

func initMain() {
	// 初始化主配置
	mainLoader.MustLoadAndWatch(gctx.New(), "app-master-config") // 无法使用
	// spew.Dump(mainLoader.Get())
	mainConfig = mainLoader.Get()
	// spew.Dump(mainConfig)

	// 初始化文件API
	watcherv2.DefaultInit(
		mainConfig.Item.UploadDir,
		mainConfig.Item.EncryptedFileDir,
		mainConfig.Item.UnlockedFileDir,
		mainConfig.Item.KeySize,
	).Run(gctx.New())
}

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			initMain()
			shutdown, err := otlphttp.Init(serviceName, endpoint, path)
			if err != nil {
				glog.Errorf(ctx, "无法注册OT LP服务: %v", err)
			}
			defer shutdown(ctx)

			s := g.Server()
			s.Group("/v1", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					hello.NewV1(),
					userctrl.NewV1(
						userctrl.WithArgon2Config(mainConfig.Server.Argon),
					), // 添加路由注册
				)

				// 认证API; 需要直接操作ghttp, 所以重写了一遍
				group.POST("/auth/login", auth.PortableLoginHandler(
					&mainConfig.Server.Argon, &mainConfig.Server.Cookie,
				))
				// 帮助信息 v1/about
				group.ALL("/about", version.NewV1(version.WithAboutConfig(mainConfig.About)))
			})

			// 需要校验的路由
			s.Group("/v1/protected", func(group *ghttp.RouterGroup) {
				// 对JWT进行序列化、类型断言, 并注入到请求上下文中
				group.Middleware(auth.InjectCookieIntoContext(&mainConfig.Server.Cookie))

				// 调试API
				group.GET("/auth/debug", auth.PortableCookiePrintHandler(&mainConfig.Server.Cookie))
				// 注销API
				group.GET("/auth/logout", auth.PortableLogoutHandler(&mainConfig.Server.Cookie))

				// 条目提交API
				group.POST("/item/submit", itemctrl.PortableItemSubmit(&mainConfig.Item, &mainConfig.Server.Argon))
				// 条目下载API; 当前不支持条目Member下载条目, 因为没有为成员重新计算份额
				group.GET("/item/download", itemctrl.PortableItemDownload(&mainConfig.Item))
				group.POST("/item/download", itemctrl.PortableItemDownload(&mainConfig.Item))

				// 其他条目管理API (不需要访问原始Request对象的)
				// 需要额外绑定中间件
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					itemctrl.NewV1(),
					userctrl.NewV3(),
				)
			})

			s.Group("/v2", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					userctrl.NewV2(
						userctrl.WithArgon2Config(mainConfig.Server.Argon),
					),
				)

				// 文件路由
				group.Group("/file", func(group *ghttp.RouterGroup) {
					group.POST("/upload/batch", filectrl.BatchUploadHandler(
						mainConfig.Item.UploadDir,
					)) // 批量文件上传
				})
			})

			s.Run()

			return nil
		},
	}
)
