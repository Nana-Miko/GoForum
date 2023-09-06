package route

import (
	"GoForum/go/net"
	"GoForum/go/net/api/commentAPI"
	"GoForum/go/net/api/postAPI"
	"GoForum/go/net/api/userAPI"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RouterGroup(ginServer *gin.Engine) {

	// 404
	ginServer.NoRoute(notFound)

	// 根路由组
	root := ginServer.Group("/")
	root.GET("/", nil)

	// API路由组
	api := ginServer.Group("api")
	api.GET("/", nil)
	// API/USER路由组
	user := api.Group("user")
	user.POST("/register", userAPI.UserRegister)
	user.POST("/verify", userAPI.UserVerify)
	user.POST("/tokenParse", AuthVerifyMiddleware(), userAPI.UserTokenParse)
	user.POST("/update", AuthVerifyMiddleware(), userAPI.UserUpdateInfo)
	user.POST("/search", userAPI.UserSearch)

	// API/POST路由组
	post := api.Group("post")

	post.POST("/create", AuthVerifyMiddleware(), TextFilteringMiddleware(), postAPI.PostCreate)
	post.POST("/search", postAPI.PostSearch)
	post.POST("/viewed", AuthVerifyMiddleware(), postAPI.PostViewed)
	post.POST("/liked", AuthVerifyMiddleware(), postAPI.PostLiked)
	post.POST("/update", AuthVerifyMiddleware(), TextFilteringMiddleware(), postAPI.PostUpdate)
	post.POST("/delete", AuthVerifyMiddleware(), postAPI.PostDelete)

	// API/Comment路由组
	comment := api.Group("comment")
	comment.POST("/create", AuthVerifyMiddleware(), TextFilteringMiddleware(), commentAPI.CommentCreate)
	comment.POST("/search", commentAPI.CommentSearch)
	comment.POST("/liked", AuthVerifyMiddleware(), commentAPI.CommentLiked)
}

func notFound(context *gin.Context) {
	response := net.GetDefaultResponse()
	defer net.ResponseDefer(&response, context)

	response.Code = http.StatusNotFound
	response.Msg = "你好像来到了一个不存在的路由,聖園未花提醒你请检查一下URL噢~Kira"
}
