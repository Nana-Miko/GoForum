package route

import (
	"GoForum/go/crypto"
	sqlite "GoForum/go/db"
	"GoForum/go/net"
	"GoForum/go/util"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
)

// AuthVerifyMiddleware 账户Token验证中间件
func AuthVerifyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		response := net.GetDefaultResponse()
		defer net.ResponseDefer(&response, c)

		// 从请求头中获取令牌
		tokenString, err := crypto.GetTokenStringFromContext(c)
		if err != nil {
			response.Error(err.Error())
			response.Code = http.StatusUnauthorized
			response.Abort()
			return
		}

		tm := crypto.GetTokenManager()
		if tm.IsExpire(tokenString) {
			response.Error("token失效")
			response.Code = http.StatusUnauthorized
			response.Abort()
			return
		}

		// 解析令牌
		token, err := crypto.TokenParse(tokenString)

		if err != nil || !token.Valid {
			response.Error("token无效 " + err.Error())
			response.Code = http.StatusUnauthorized
			response.Abort()
			return
		}

		// 将令牌中的声明信息放入请求上下文中，以便后续处理中使用
		claims := token.Claims.(jwt.MapClaims)

		userJson, ok1 := claims["user"].(map[string]any)
		email, ok2 := userJson["email"].(string)
		name, ok3 := userJson["name"].(string)
		pswMD5, ok4 := userJson["pswMd5"].(string)
		sex, ok5 := userJson["sex"].(float64)

		if !(ok1 && ok2 && ok3 && ok4 && ok5) {
			response.Error("token无效，User断言失败")
			response.Code = http.StatusUnauthorized
			response.Abort()
			return
		}

		user := sqlite.User{
			Email:  email,
			Name:   name,
			PswMd5: pswMD5,
			Sex:    int(sex),
		}

		response.AbortResponse()
		c.Set("user", user) // 在请求上下文中设置User
		c.Next()

	}
}

// TextFilteringMiddleware 文本敏感词检测中间件
func TextFilteringMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		response := net.GetDefaultResponse()
		defer net.ResponseDefer(&response, c)
		text := c.PostForm("title") + " " + c.PostForm("content")
		valRes, err := util.DoubleValidation(text)
		if err != nil {
			response.UnknownError(err)
			response.Abort()
			return
		}
		if !valRes.Pass {
			response.Data = valRes.Words
			response.Error(valRes.Tips)
			response.Abort()
			return
		}
		response.AbortResponse()
		c.Next()

	}
}
