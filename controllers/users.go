package controllers

import (
	"net/http"

	"go_bbs/helpers"

	"go_bbs/models"

	"go_bbs/services"

	"github.com/fpay/gopress"
	"github.com/garyburd/redigo/redis"
	"github.com/leebenson/conform"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
	"strconv"
	"time"
)

// UsersController
type UsersController struct {
	// Uncomment this line if you want to use services in the app
	app       *gopress.App
	db        *services.DatabaseService
	cache     *services.CacheService
	helper    *helpers.Common
	validator *CustomValidator
	user      *models.User
}

type (
	userRegisterData struct {
		Name         string `json:"name" form:"name" validate:"required,lt=32" conform:"trim"`
		Email        string `json:"email" form:"email" validate:"required,email" conform:"trim"`
		Password     string `json:"password" form:"password" validate:"required,gt=3,lt=32" conform:"trim"`
		SurePassword string `json:"surePassword" form:"surePassword" validate:"required" conform:"trim"`
	}

	userLoginData struct {
		Name     string `json:"name" form:"name" validate:"required,lt=32" conform:"trim"`
		Password string `json:"password" form:"password" validate:"required,gt=3,lt=32" conform:"trim"`
	}

	CustomValidator struct {
		validator *validator.Validate
	}
)

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// NewUsersController returns users controller instance.
func NewUsersController() *UsersController {
	return new(UsersController)
}

// RegisterRoutes registes routes to app
// It is used to implements gopress.Controller.
func (c *UsersController) RegisterRoutes(app *gopress.App) {
	// Uncomment this line if you want to use services in the app
	c.app = app
	c.db = app.Services.Get(services.DatabaseServiceName).(*services.DatabaseService)
	c.cache = app.Services.Get(services.CacheServiceName).(*services.CacheService)
	c.validator = &CustomValidator{validator: validator.New()}
	c.helper = new(helpers.Common)

	app.GET("/users", c.IndexGetAction)
	app.GET("/users/login", c.LoginGetAction)
	app.POST("/users/doLogin", c.DoLoginPostAction)
	app.GET("/users/register", c.RegisterGetAction)
	app.POST("/users/doRegister", c.DoRegisterPostAction)
	app.GET("/users/posts", c.PostsGetAction)
	app.GET("/users/getReplies", c.GetRepliesGetAction)
	app.GET("/users/sendReplies", c.SendRepliesGetAction)

}

func (c *UsersController) IndexGetAction(ctx gopress.Context) error {
	//	data := map[string]interface{}{}
	user := c.getUser(ctx)
	if user == nil {
		return ctx.Redirect(http.StatusFound, "/users/login")
	}
	return ctx.Render(http.StatusOK, "users/index", user)
}

func (c *UsersController) getUser(ctx gopress.Context) *models.User {
	cookie, err := ctx.Cookie("uid")
	if err != nil {
		return nil
	}
	user := &models.User{}
	uid, _ := strconv.Atoi(cookie.Value)
	if c.db.DB.First(user, uid).RowsAffected == 0 {
		return nil
	}
	return user
}

func (c *UsersController) LoginGetAction(ctx gopress.Context) error {
	data := map[string]interface{}{
		"loginUrl":    "/users/doLogin",
		"registerUrl": "/users/register",
	}
	return ctx.Render(http.StatusOK, "users/login", data)
}

func (c *UsersController) DoLoginPostAction(ctx gopress.Context) error {
	data := map[string]interface{}{}

	if c.loginLimit(ctx) == false {
		data["code"] = 4000
		data["message"] = "登录太频繁,5分钟再试"
		return ctx.JSON(http.StatusOK, data)
	}

	userData := new(userLoginData)
	var err error
	if err = ctx.Bind(userData); err != nil {
		data["code"] = 4001
		data["message"] = "数据不完整"
		return ctx.JSON(http.StatusOK, data)
	}

	conform.Strings(userData)

	if err = c.validator.Validate(userData); err != nil {
		data["code"] = 4002
		data["message"] = "数据不合法"
		return ctx.JSON(http.StatusOK, data)
	}

	user := models.User{}

	if c.db.DB.Where("name = ?", userData.Name).First(&user).RowsAffected == 0 {
		data["code"] = 4003
		data["message"] = "用户不存在"
		return ctx.JSON(http.StatusOK, data)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userData.Password)); err != nil {
		ctx.Logger().Error(err)
		data["code"] = 4004
		data["message"] = "密码错误"
		return ctx.JSON(http.StatusOK, data)
	}

	c.cacheUserInfo(user, ctx)
	data["code"] = 0
	data["message"] = "登录成功"
	data["redirectUrl"] = "/users"
	return ctx.JSON(http.StatusOK, data)
}

// 登录限制 5分钟3次
func (c *UsersController) loginLimit(ctx gopress.Context) bool {
	ip := ctx.RealIP()
	key := "loginLimit:" + ip
	c.cache.Redis.Do("SET", key, 0, "NX")
	loginTime, _ := redis.Int(c.cache.Redis.Do("INCR", key))
	c.cache.Redis.Do("EXPIRE", key, 300)

	data := map[string]interface{}{"ip": ip, "loginTime": loginTime}
	ctx.Logger().Info("loginLimit:", data)

	if loginTime <= 3 {
		return true
	}
	return false
}

func (c *UsersController) RegisterGetAction(ctx gopress.Context) error {
	data := map[string]interface{}{
		"loginUrl":    "/users/doLogin",
		"registerUrl": "/users/doRegister",
	}
	return ctx.Render(http.StatusOK, "users/register", data)
}

func (c *UsersController) DoRegisterPostAction(ctx gopress.Context) error {
	data := map[string]interface{}{}
	userData := new(userRegisterData)
	var err error
	if err = ctx.Bind(userData); err != nil {
		data["code"] = 4001
		data["message"] = "数据不完整"
		data["error"] = err.Error()
		return ctx.JSON(http.StatusOK, data)
	}

	conform.Strings(userData)

	if err = c.validator.Validate(userData); err != nil {
		data["code"] = 4002
		data["message"] = "数据不合法"
		data["error"] = err.Error()
		return ctx.JSON(http.StatusOK, data)
	}

	user := models.User{
		Name:  userData.Name,
		Email: userData.Email,
	}
	if has := c.db.DB.Where("name = ?", user.Name).First(&user).RowsAffected; has != 0 {
		data["code"] = 4003
		data["message"] = "用户名已存在"
		return ctx.JSON(http.StatusOK, data)
	}

	if has := c.db.DB.Where("email = ?", user.Email).First(&user).RowsAffected; has != 0 {
		data["code"] = 4004
		data["message"] = "邮箱已注册"
		return ctx.JSON(http.StatusOK, data)
	}

	// make password hash
	hash, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		data["code"] = 5000
		data["message"] = "网络繁忙,稍后再试"
		ctx.Logger().Info(err.Error())
		return ctx.JSON(http.StatusOK, data)
	}

	user.Password = string(hash)
	//c.db.DB.LogMode(true)

	res := c.db.DB.Create(&user)
	if res.Error != nil {
		data["code"] = 5000
		data["message"] = "网络繁忙,稍后再试"
		ctx.Logger().Info(res.GetErrors())
		return ctx.JSON(http.StatusOK, data)
	}

	c.cacheUserInfo(user, ctx)

	data["code"] = 0
	data["message"] = "注册成功"
	data["redirectUrl"] = "/users"
	return ctx.JSON(http.StatusOK, data)
}

func (c *UsersController) cacheUserInfo(user models.User, ctx gopress.Context) bool {

	cookie := new(http.Cookie)
	cookie.Name = "uid"
	cookie.Value = strconv.Itoa(int(user.ID))
	cookie.Path = "/"
	cookie.Expires = time.Now().Add(24 * time.Minute)
	ctx.SetCookie(cookie)

	return true
}

func (c *UsersController) PostsGetAction(ctx gopress.Context) error {
	user := c.getUser(ctx)

	c.db.DB.Model(user).Related(&user.Posts)
	data := map[string]interface{}{
		"posts":      user.Posts,
		"addPostUrl": "/posts/add",
	}
	return ctx.Render(http.StatusOK, "users/posts", data)
}

func (c *UsersController) GetRepliesGetAction(ctx gopress.Context) error {
	data := map[string]interface{}{}

	return ctx.Render(http.StatusOK, "users/getReplies", data)
}

func (c *UsersController) SendRepliesGetAction(ctx gopress.Context) error {

	data := map[string]interface{}{}

	user := c.getUser(ctx)
	replies := []*models.Reply{}
	c.db.DB.Model(user).Related(&replies)

	data["replies"] = replies

	return ctx.Render(http.StatusOK, "users/sendReplies", data)
}
