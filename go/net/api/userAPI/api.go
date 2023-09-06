package userAPI

import (
	"GoForum/go/crypto"
	sqlite "GoForum/go/db"
	"GoForum/go/net"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"time"
)

// UserRegister 用户注册
func UserRegister(context *gin.Context) {
	response := net.GetDefaultResponse()
	defer net.ResponseDefer(&response, context)

	email := context.PostForm("email")
	pswMd5 := crypto.CalculateMD5(context.PostForm("psw"))
	sex, err := strconv.Atoi(context.PostForm("sex"))
	if err != nil {
		response.Error("性别字段只能为数字，0男性1女性")
		return
	}
	name := context.PostForm("name")

	db, err := sqlite.OpenDB()
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer db.Close()
	_, err = db.Exec("INSERT INTO user(email,name,psw_md5,sex) VALUES (?,?,?,?)", email, name, pswMd5, sex)
	if err != nil {
		response.UnknownError(err)
		return
	}
	response.Successful(sqlite.User{
		Email: email,
		Name:  name,
		Sex:   sex,
	})
}

// UserVerify 用户登录验证
func UserVerify(context *gin.Context) {
	response := net.GetDefaultResponse()
	defer net.ResponseDefer(&response, context)

	email := context.PostForm("email")
	pswMd5 := crypto.CalculateMD5(context.PostForm("psw"))
	db, err := sqlite.OpenDB()
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer db.Close()
	rows, err := db.Query("SELECT * FROM user WHERE email=?", email)
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer rows.Close()
	var user sqlite.User
	for rows.Next() {
		err := rows.Scan(&user.Email, &user.Name, &user.PswMd5, &user.Sex)
		if err != nil {
			response.UnknownError(err)
			return
		}
	}
	if user.Email == "" {
		response.Error("用户名或密码错误")
		return
	}
	if pswMd5 != user.PswMd5 {
		response.Error("用户名或密码错误")
		return
	}

	validTimeStr := context.PostForm("validTime")
	if validTimeStr == "" {
		response.Successful(map[string]string{"token": crypto.CreateUserToken(user, 1440*time.Minute)})
	} else {
		validTime, err := strconv.Atoi(validTimeStr)
		if err != nil {
			response.Error("validTime格式错误")
			return
		}
		response.Successful(gin.H{
			"user":  user,
			"token": crypto.CreateUserToken(user, time.Duration(validTime)*time.Minute),
		})

	}

}

// UserTokenParse 用户Token解析
func UserTokenParse(context *gin.Context) {
	response := net.GetDefaultResponse()
	defer net.ResponseDefer(&response, context)

	user := crypto.GetUserFromContext(context)

	user.PswMd5 = ""
	response.Successful(user)
}

// UserUpdateInfo 用户信息更新
func UserUpdateInfo(context *gin.Context) {
	// TODO 用户信息更新
	response := net.GetDefaultResponse()
	defer net.ResponseDefer(&response, context)

	user := crypto.GetUserFromContext(context)
	email := user.Email
	pswMd5 := crypto.CalculateMD5(context.PostForm("psw"))
	if pswMd5 == user.PswMd5 {
		response.Error("密码无法与之前相同")
		return
	}
	sex, err := strconv.Atoi(context.PostForm("sex"))
	if err != nil {
		response.Error("性别字段只能为数字，0男性1女性")
		return
	}
	name := context.PostForm("name")

	db, err := sqlite.OpenDB()
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer db.Close()

	_, err = db.Exec("UPDATE user SET psw_md5=?,sex=?,name=? WHERE email=?", pswMd5, sex, name, email)
	if err != nil {
		response.UnknownError(err)
		return
	}

	tokenString, _ := crypto.GetTokenStringFromContext(context)
	err = crypto.GetTokenManager().Invalidate(tokenString)
	if err != nil {
		response.UnknownError(err)
		return
	}
	response.Msg = "修改成功，原Token已失效，请重新登录"

}

// UserSearch 用户信息查询
func UserSearch(context *gin.Context) {
	response := net.GetDefaultResponse()
	defer net.ResponseDefer(&response, context)

	usersEmail := context.PostFormArray("userEmail")

	if len(usersEmail) <= 0 {
		response.Error("userEmail不能为空")
		return
	}

	db, err := sqlite.OpenDB()
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer db.Close()

	// 构建IN子句参数
	params := make([]interface{}, len(usersEmail))
	for i, id := range usersEmail {
		params[i] = id
	}

	// 创建IN子句占位符字符串
	placeholders := strings.Repeat("?,", len(usersEmail))
	placeholders = placeholders[:len(placeholders)-1] // 移除末尾的逗号

	// 构建SQL语句
	sqlStmt := fmt.Sprintf("SELECT * FROM user WHERE email IN (%s)", placeholders)

	// 执行SQL查询操作
	rows, err := db.Query(sqlStmt, params...)
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer rows.Close()
	var usersInfo = make(map[string]sqlite.User)
	for rows.Next() {
		userInfo := sqlite.User{}
		err := rows.Scan(&userInfo.Email, &userInfo.Name, &userInfo.PswMd5, &userInfo.Sex)
		if err != nil {
			response.UnknownError(err)
			return
		}
		userInfo.PswMd5 = ""
		usersInfo[userInfo.Email] = userInfo

	}

	if len(usersInfo) <= 0 {
		response.Error("没有找到的对象")
		return
	}

	response.Successful(usersInfo)

}
