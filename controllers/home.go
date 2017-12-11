package controllers

import (
	"net/http"

	"github.com/fpay/gopress"
	. "go_bbs/conf"
	"go_bbs/models"
	"go_bbs/services"
)

// HomeController
type HomeController struct {
	// Uncomment this line if you want to use services in the app
	app *gopress.App
	db  *services.DatabaseService
}

// NewHomeController returns home controller instance.
func NewHomeController() *HomeController {
	return new(HomeController)
}

// RegisterRoutes registes routes to app
// It is used to implements gopress.Controller.
func (c *HomeController) RegisterRoutes(app *gopress.App) {
	// Uncomment this line if you want to use services in the app
	c.app = app
	c.db = app.Services.Get(services.DatabaseServiceName).(*services.DatabaseService)

	app.GET("/home", c.IndexGetAction)
	app.GET("/", c.IndexGetAction)
}

func (c *HomeController) IndexGetAction(ctx gopress.Context) error {
	data := map[string]interface{}{}
	c.db.DB.LogMode(true)
	ctx.Logger().Info(Conf.Redis.Server)
	//ctx.Logger().Info(conf.Conf.Redis.Server)

	post := models.Post{}

	data["hotPosts"] = post.HotPosts(c.db, 10)
	return ctx.Render(http.StatusOK, "home/index", data)
}
