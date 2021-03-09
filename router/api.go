package routers

import (
	"github.com/1024casts/snake/config"
	snake "github.com/1024casts/snake/pkg"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger" //nolint: goimports
	"github.com/swaggo/gin-swagger/swaggerFiles"

	"github.com/1024casts/snake/app/api"
	"github.com/1024casts/snake/app/api/http/v1/user"

	// import swagger handler
	_ "github.com/1024casts/snake/docs" // docs is generated by Swag CLI, you have to import it.
	"github.com/1024casts/snake/pkg/middleware"
)

// Load loads the middlewares, routes, handlers.
func Load(g *gin.Engine, mw ...gin.HandlerFunc) *gin.Engine {
	// 使用中间件
	g.Use(middleware.NoCache)
	g.Use(middleware.Options)
	g.Use(middleware.Secure)
	g.Use(middleware.Logging())
	g.Use(middleware.RequestID())
	g.Use(middleware.Prom(nil))
	g.Use(middleware.Trace())
	g.Use(mw...)

	// 404 Handler.
	g.NoRoute(api.RouteNotFound)
	g.NoMethod(api.RouteNotFound)

	// 静态资源，主要是图片
	g.Static("/static", "./static")

	// 返回404，仅在test环境下开启，线上关闭
	if config.Conf.App.Mode == snake.ModeDebug {
		// swagger api docs
		g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		// pprof router 性能分析路由
		// 默认关闭，开发环境下可以打开
		// 访问方式: HOST/debug/pprof
		// 通过 HOST/debug/pprof/profile 生成profile
		// 查看分析图 go tool pprof -http=:5000 profile
		// see: https://github.com/gin-contrib/pprof
		pprof.Register(g)
	} else {
		// disable swagger docs for release  env=release
		g.GET("/swagger/*any", ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "env"))
	}

	// 认证相关路由
	g.POST("/v1/register", user.Register)
	g.POST("/v1/login", user.Login)
	g.POST("/v1/login/phone", user.PhoneLogin)
	g.GET("/v1/vcode", user.VCode)

	// 用户
	g.GET("/v1/users/:id", user.Get)

	u := g.Group("/v1/users")
	u.Use(middleware.AuthMiddleware())
	{
		u.PUT("/:id", user.Update)
		u.POST("/follow", user.Follow)
		u.GET("/:id/following", user.FollowList)
		u.GET("/:id/followers", user.FollowerList)
	}

	return g
}
