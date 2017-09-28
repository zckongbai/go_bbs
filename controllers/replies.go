package controllers

import (
	"net/http"

	"github.com/fpay/gopress"
)

// RepliesController
type RepliesController struct {
	// Uncomment this line if you want to use services in the app
	// app *gopress.App
}

// NewRepliesController returns replies controller instance.
func NewRepliesController() *RepliesController {
	return new(RepliesController)
}

// RegisterRoutes registes routes to app
// It is used to implements gopress.Controller.
func (c *RepliesController) RegisterRoutes(app *gopress.App) {
	// Uncomment this line if you want to use services in the app
	// c.app = app

	app.GET("/replies/sample", c.SampleGetAction)
	// app.POST("/replies/sample", c.SamplePostAction)
	// app.PUT("/replies/sample", c.SamplePutAction)
	// app.DELETE("/replies/sample", c.SampleDeleteAction)
}

// SampleGetAction Action
// Parameter gopress.Context is just alias of echo.Context
func (c *RepliesController) SampleGetAction(ctx gopress.Context) error {
	// Or you can get app from request context
	// app := gopress.AppFromContext(ctx)
	data := map[string]interface{}{}
	return ctx.Render(http.StatusOK, "replies/sample", data)
}
