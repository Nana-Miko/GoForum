package crypto

import (
	sqlite "GoForum/go/db"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

// 令牌私钥
var secretKey = []byte("My-1121917292")

// CalculateMD5 字符串转MD5
func CalculateMD5(str string) string {
	md5Hash := md5.Sum([]byte(str))
	md5String := hex.EncodeToString(md5Hash[:])
	return md5String
}

// CreateUserToken 生成用户Token
func CreateUserToken(user sqlite.User, expirationTime time.Duration) string {

	// 创建一个令牌的声明
	claims := jwt.MapClaims{
		"user": user,                                  // User信息
		"exp":  time.Now().Add(expirationTime).Unix(), // 令牌过期时间
	}
	// 使用声明和密钥创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 使用密钥签名令牌
	signedToken, _ := token.SignedString(secretKey)
	return signedToken
}

// TokenParse Token解析
func TokenParse(tokenString string) (*jwt.Token, error) {
	//tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 使用密钥来验证令牌的签名
		return secretKey, nil
	})

	return token, err
}

// GetUserFromContext 在上下文中获取User
func GetUserFromContext(context *gin.Context) sqlite.User {
	return context.MustGet("user").(sqlite.User)
}

// GetTokenStringFromContext 在上下文中获取正确的Token字符串
func GetTokenStringFromContext(context *gin.Context) (string, error) {
	tokenString := context.GetHeader("Authorization")
	if tokenString == "" {
		return "", errors.New("token丢失")
	}
	return strings.TrimPrefix(tokenString, "Bearer "), nil

}
