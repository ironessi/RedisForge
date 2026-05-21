package cmd

import (
	"context"
	"redis-demo/internal/controller/auth"
	"redis-demo/internal/controller/user"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				// auth 相关接口不需要登录，例如注册和登录。
				group.Bind(
					auth.NewV1(),
				)
				// user 相关接口需要先通过 JWT 鉴权。
				group.Group("/", func(group *ghttp.RouterGroup) {
					group.Middleware(middleware.Auth) // JWT 鉴权中间件，验证用户身份并把用户信息写入请求上下文
					group.Bind(
						user.NewV1(),
					)
				})
			})
			s.Run()
			return nil
		},
	}
)
