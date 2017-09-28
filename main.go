package main

import (
	"go_bbs/controllers"

	"github.com/fpay/gopress"
	"go_bbs/services"
	"go_bbs/middlewares"
	. "go_bbs/conf"
)

func main() {

	s := gopress.NewServer(gopress.ServerOptions{
		Host: Conf.Server.DomainWeb,
		Port: Conf.Server.Addr,
	})

	// init and register services
	 s.RegisterServices(
	 	services.NewDatabaseService(),
		services.NewCacheService(),
	 )

	// register middlewares
	s.RegisterGlobalMiddlewares(
		gopress.NewLoggingMiddleware("global", nil),
		middlewares.NewAuthMiddleware(),
	)

	// init and register controllers
	s.RegisterControllers(
		controllers.NewUsersController(),
		controllers.NewPostsController(),
		controllers.NewHomeController(),
	)

	s.App().Static("/static", "public")

	s.Start()
}
