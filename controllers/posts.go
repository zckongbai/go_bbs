package controllers

import (
	"net/http"

	"github.com/fpay/gopress"
	"github.com/garyburd/redigo/redis"
	"github.com/leebenson/conform"
	"go_bbs/helpers"
	"go_bbs/models"
	"go_bbs/services"
	"gopkg.in/go-playground/validator.v9"
	"strconv"
	"time"
)

// PostsController
type PostsController struct {
	// Uncomment this line if you want to use services in the app
	app       *gopress.App
	db        *services.DatabaseService
	cache     *services.CacheService
	helper    *helpers.Common
	validator *validator.Validate
}

type (
	addPostData struct {
		Title   string `json:"title" form:"title" validate:"required,lt=64" conform:"trim"`
		Content string `json:"content" form:"content" validate:"required" conform:"trim"`
	}
	updatePostData struct {
		ID      string `json:"id" form:"id" validate:"required" conform:"trim"`
		Title   string `json:"title" form:"title" validate:"required,lt=64" conform:"trim"`
		Content string `json:"content" form:"content" validate:"required" conform:"trim"`
	}

	addReplyData struct {
		PostId  int    `json:"post_id" form:"post_id" validate:"required" conform:"trim"`
		Content string `json:"content" form:"content" validate:"required" conform:"trim"`
	}
)

// NewPostsController returns posts controller instance.
func NewPostsController() *PostsController {
	return new(PostsController)
}

// RegisterRoutes registes routes to app
// It is used to implements gopress.Controller.
func (c *PostsController) RegisterRoutes(app *gopress.App) {
	// Uncomment this line if you want to use services in the app
	c.app = app
	c.db = app.Services.Get(services.DatabaseServiceName).(*services.DatabaseService)
	c.cache = app.Services.Get(services.CacheServiceName).(*services.CacheService)
	c.helper = new(helpers.Common)
	c.validator = validator.New()

	app.GET("/posts", c.IndexGetAction)
	app.GET("/posts/add", c.AddGetAction)
	app.GET("/posts/:id", c.ViewGetAction)
	app.POST("/posts/doAdd", c.DoAddPostAction)
	app.GET("/posts/update/:id", c.UpdateGetAction)
	app.POST("/posts/doUpdate", c.DoUpdatePostAction)
	app.POST("/posts/addReply", c.AddReplyPostAction)
}

func (c *PostsController) IndexGetAction(ctx gopress.Context) error {
	// Or you can get app from request context
	// app := gopress.AppFromContext(ctx)

	//pindex := int(ctx.QueryParam("p"))
	//if pindex == 0 {
	pindex := 1
	//}
	psize := 10
	data := map[string]interface{}{}
	posts := []models.Post{}
	c.db.DB.Limit(psize).Offset((pindex - 1) * psize).Find(&posts).Rows()
	//ctx.Logger().Info("res:", res)
	ctx.Logger().Info("posts:", posts)
	data["posts"] = posts
	data["addUrl"] = "/posts/add"
	return ctx.Render(http.StatusOK, "posts/index", data)
}

func (c *PostsController) AddGetAction(ctx gopress.Context) error {

	data := map[string]interface{}{
		"title":  "添加文章",
		"addUrl": "/posts/doAdd",
	}
	return ctx.Render(http.StatusOK, "posts/add", data)
}

func (c *PostsController) DoAddPostAction(ctx gopress.Context) error {
	data := map[string]interface{}{}

	if c.addLimit(ctx) == false {
		data["code"] = 4003
		data["message"] = "重复提交"
		return ctx.JSON(http.StatusOK, data)
	}

	addData := new(addPostData)
	var err error
	if err = ctx.Bind(addData); err != nil {
		data["code"] = 4001
		data["message"] = "数据不完整"
		return ctx.JSON(http.StatusOK, data)
	}

	conform.Strings(addData)

	if err = c.validator.Struct(addData); err != nil {
		ctx.Logger().Error(err)
		data["code"] = 4002
		data["message"] = "数据不合法"
		return ctx.JSON(http.StatusOK, data)
	}
	cookie, err := ctx.Cookie("uid")
	if err != nil {
		data["code"] = 4006
		data["message"] = "请先登录"
		return ctx.JSON(http.StatusOK, data)
	}
	uid, _ := strconv.Atoi(cookie.Value)
	post := models.Post{
		UserId:  uid,
		Title:   addData.Title,
		Content: addData.Content,
	}

	if c.db.DB.Create(&post).Error != nil {
		data["code"] = 5000
		data["message"] = "服务器繁忙, 请稍后再试"
		return ctx.JSON(http.StatusOK, data)
	}

	data["code"] = 0
	data["message"] = "添加成功"
	data["redirectUrl"] = "/posts/" + strconv.Itoa(int(post.ID))
	return ctx.JSON(http.StatusOK, data)
}

func (c *PostsController) addLimit(ctx gopress.Context) bool {
	return true
}

func (c *PostsController) ViewGetAction(ctx gopress.Context) error {
	id := ctx.Param("id")
	data := map[string]interface{}{}

	post := models.Post{}
	c.db.DB.First(&post, id)

	// 更新点击数
	c.updatePostClickNumber(ctx, &post)

	c.db.DB.Select("name").Model(&post).Related(&post.User)

	c.db.DB.Model(&post).Related(&post.Replis)
	for index, _ := range post.Replis {
		c.db.DB.Select("name").Model(&post.Replis[index]).Related(&post.Replis[index].User)
	}

	data["post"] = post
	data["replyUrl"] = "/posts/addReply"

	return ctx.Render(http.StatusOK, "posts/view", data)
}

func (c *PostsController) updatePostClickNumber(ctx gopress.Context, post *models.Post) bool {
	// 不增加自己的文章
	var key string
	uidCookie, err := ctx.Cookie("uid")
	if err == nil {
		uid, _ := strconv.Atoi(uidCookie.Value)
		if uid == post.UserId {
			return false
		}
		key = "view:" + strconv.Itoa(int(post.ID)) + "|" + uidCookie.Value
	} else {
		ip := ctx.RealIP()
		key = "view:" + strconv.Itoa(int(post.ID)) + "|" + ip
	}
	if has, _ := redis.Bool(c.cache.Redis.Do("EXISTS", key)); has == true {
		return false
	}
	// 3分钟间隔
	c.cache.Redis.Do("SET", key, 1, "EX", 150)

	post.ClickNumber++
	c.db.DB.Save(post)
	return true
}

func (c *PostsController) AddReplyPostAction(ctx gopress.Context) error {
	data := map[string]interface{}{}

	addData := new(addReplyData)
	var err error
	if err = ctx.Bind(addData); err != nil {
		data["code"] = 4001
		data["message"] = "数据不完整"
		return ctx.JSON(http.StatusOK, data)
	}
	conform.Strings(addData)
	if err = c.validator.Struct(addData); err != nil {
		ctx.Logger().Error(err)
		data["code"] = 4002
		data["message"] = "数据不合法"
		return ctx.JSON(http.StatusOK, data)
	}
	cookie, err := ctx.Cookie("uid")
	if err != nil {
		data["code"] = 4006
		data["message"] = "请先登录"
		return ctx.JSON(http.StatusOK, data)
	}

	uid, _ := strconv.Atoi(cookie.Value)
	post := &models.Post{}

	c.db.DB.First(post, addData.PostId)

	// 事务
	tran := c.db.DB.Begin()
	post.ReplyNumber++
	//post.LastReplyAt = c.helper.DateTime()
	post.LastReplyAt = time.Now()

	reply := models.Reply{
		PostId:    addData.PostId,
		PostTitle: post.Title,
		UserId:    uid,
		Floor:     post.ReplyNumber,
		Content:   addData.Content,
	}

	err1 := tran.Create(&reply).GetErrors()
	err2 := tran.Save(post).GetErrors()
	if len(err1) == 0 && len(err2) == 0 {
		tran.Commit()
		data["code"] = 0
		data["message"] = "添加评论成功"
		data["redirectUrl"] = "/posts/" + strconv.Itoa(addData.PostId)
		return ctx.JSON(http.StatusOK, data)
	}
	ctx.Logger().Info("sqlerr1:", err1)
	ctx.Logger().Info("sqlerr2:", err2)
	tran.Rollback()
	data["code"] = 5000
	data["message"] = "服务器繁忙, 稍后再试"
	return ctx.JSON(http.StatusOK, data)
}

func (c *PostsController) UpdateGetAction(ctx gopress.Context) error {
	data := map[string]interface{}{}

	id := ctx.Param("id")

	post := models.Post{}
	if c.db.DB.First(&post, id).Error != nil {
		data["message"] = "文章不存在"
		data["redirectUrl"] = "/users"
		data["redirectName"] = "用户中心"
		return ctx.Render(http.StatusOK, "common/error", data)
	}
	cookie, err := ctx.Cookie("uid")
	if err != nil {
		data["message"] = "请先登录"
		data["redirectUrl"] = "/users/login"
		data["redirectName"] = "用户登录"
		return ctx.Render(http.StatusOK, "common/error", data)
	}

	uid, _ := strconv.Atoi(cookie.Value)

	if uid != post.UserId {
		data["message"] = "非法访问"
		data["redirectUrl"] = "/users"
		data["redirectName"] = "用户中心"
		return ctx.Render(http.StatusOK, "common/error", data)
	}

	data["post"] = post
	data["updateUrl"] = "/posts/doUpdate"
	return ctx.Render(http.StatusOK, "/posts/update", data)
}

func (c *PostsController) DoUpdatePostAction(ctx gopress.Context) error {
	id := ctx.FormValue("id")
	data := map[string]interface{}{
		"redirectUrl":  "/posts/" + id,
		"redirectName": "查看文章",
	}

	if c.addLimit(ctx) == false {
		data["message"] = "重复修改"
		return ctx.Render(http.StatusOK, "common/error", data)
	}

	updateData := new(updatePostData)
	var err error
	if err = ctx.Bind(updateData); err != nil {
		data["message"] = "数据不完整"
		return ctx.Render(http.StatusOK, "common/error", data)
	}
	conform.Strings(updateData)
	if err = c.validator.Struct(updateData); err != nil {
		ctx.Logger().Error(err)
		data["message"] = "数据不合法"
		data["redirectName"] = "查看文章"
		return ctx.JSON(http.StatusOK, data)
	}

	post := models.Post{}
	if c.db.DB.First(&post, id).Error != nil {
		data["message"] = "文章不存在"
		data["redirectUrl"] = "/users"
		data["redirectName"] = "用户中心"
		return ctx.Render(http.StatusOK, "common/error", data)
	}
	cookie, _ := ctx.Cookie("uid")
	uid, _ := strconv.Atoi(cookie.Value)

	if uid != post.UserId {
		data["message"] = "非法访问"
		data["redirectUrl"] = "/users"
		data["redirectName"] = "用户中心"
		return ctx.Render(http.StatusOK, "common/error", data)
	}

	post.Title = updateData.Title
	post.Content = updateData.Content
	c.db.DB.Save(post)

	data["message"] = "修改成功"

	return ctx.Render(http.StatusOK, "common/success", data)
}
