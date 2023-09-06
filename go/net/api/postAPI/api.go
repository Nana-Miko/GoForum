package postAPI

import (
	"GoForum/go/crypto"
	sqlite "GoForum/go/db"
	"GoForum/go/net"
	"GoForum/go/net/api"
	"GoForum/go/net/throttle"
	"GoForum/go/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
)

var likedThrottler throttle.Throttler
var viewedThrottler throttle.Throttler

func init() {
	// 3600000毫秒=1小时
	// 1秒=1000毫秒
	likedThrottler = throttle.NewThrottler(1000 * 10)
	viewedThrottler = throttle.NewThrottler(1000 * 10)
}

// PostCreate 发布帖子
func PostCreate(context *gin.Context) {
	response := net.GetDefaultResponse()
	defer net.ResponseDefer(&response, context)

	user := crypto.GetUserFromContext(context)

	content := context.PostForm("content")
	title := context.PostForm("title")
	db, err := sqlite.OpenDB()
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer db.Close()
	res, err := db.Exec("INSERT INTO posts(content,user_email,time,title) values (?,?,?,?)", content, user.Email, util.CurrentTimeStampStrMilli(), title)
	if err != nil {
		response.UnknownError(err)
		return
	}
	lastInsertID, err := res.LastInsertId()
	if err != nil {
		response.UnknownError(err)
		return
	}

	response.Successful(sqlite.Posts{
		PostID: int(lastInsertID),
	})

}

// PostSearch 查询帖子
func PostSearch(context *gin.Context) {
	response := net.GetDefaultResponse()
	defer net.ResponseDefer(&response, context)

	db, err := sqlite.OpenDB()
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer db.Close()
	pagination := api.PaginationQueryParse(context)

	// 执行指定ID查询
	if pagination.Id != -10086 {
		rows, err := db.Query("SELECT * FROM posts WHERE post_id=?", pagination.Id)
		if err != nil {
			response.UnknownError(err)
			return
		}
		defer rows.Close()
		var post sqlite.Posts
		for rows.Next() {
			err := rows.Scan(&post.PostID, &post.Content, &post.UserEmail, &post.Time, &post.Hot, &post.Liked, &post.Viewed, &post.Comment, &post.Title)
			if err != nil {
				response.UnknownError(err)
				return
			}
		}
		response.Successful(post)
		return
	}

	query := fmt.Sprintf("SELECT * FROM posts ORDER BY %s LIMIT ? OFFSET ?", pagination.OrderByColumn)
	if pagination.DESC {
		query = fmt.Sprintf("SELECT * FROM posts ORDER BY %s DESC LIMIT ? OFFSET ?", pagination.OrderByColumn)
	}

	stmt, err := db.Prepare(query)
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query(pagination.PerPage, pagination.Offset)
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer rows.Close()
	posts := make([]sqlite.Posts, 0)
	for rows.Next() {
		var post sqlite.Posts
		err := rows.Scan(&post.PostID, &post.Content, &post.UserEmail, &post.Time, &post.Hot, &post.Liked, &post.Viewed, &post.Comment, &post.Title)
		if err != nil {
			response.UnknownError(err)
			return
		}
		posts = append(posts, post)
	}
	response.Successful(posts)
}

// PostUpdate 更新帖子
func PostUpdate(context *gin.Context) {
	response := net.GetDefaultResponse()
	defer net.ResponseDefer(&response, context)

	user := crypto.GetUserFromContext(context)

	content := context.PostForm("content")
	title := context.PostForm("title")
	postId, err := strconv.Atoi(context.PostForm("postId"))
	if err != nil {
		response.Error("postId格式不正确")
		return
	}
	db, err := sqlite.OpenDB()
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer db.Close()
	_, err = db.Exec("UPDATE posts SET content=?,title=?,time=? WHERE post_id=? AND user_email=?", content, title, util.CurrentTimeStampStrMilli(), postId, user.Email)
	if err != nil {
		response.UnknownError(err)
		return
	}

	response.Successful(sqlite.Posts{
		PostID: postId,
	})
}

// PostDelete 删除帖子
func PostDelete(context *gin.Context) {
	response := net.GetDefaultResponse()
	defer net.ResponseDefer(&response, context)

	user := crypto.GetUserFromContext(context)

	postId, err := strconv.Atoi(context.PostForm("postId"))
	if err != nil {
		response.Error("postId格式不正确")
		return
	}
	db, err := sqlite.OpenDB()
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer db.Close()
	_, err = db.Exec("DELETE FROM posts WHERE post_id=? AND user_email=?", postId, user.Email)
	if err != nil {
		response.UnknownError(err)
		return
	}

}

// PostLiked 点赞帖子
func PostLiked(context *gin.Context) {

	response := net.GetDefaultResponse()
	defer net.ResponseDefer(&response, context)

	postId, err := strconv.Atoi(context.PostForm("postId"))
	if err != nil {
		response.Error("postId格式错误")
		return
	}
	if !likedThrottler.ThrottleVerify(context.MustGet("user").(sqlite.User), postId) {
		response.Error("操作限流")
		return
	}

	db, err := sqlite.OpenDB()
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer db.Close()
	_, err = db.Exec("UPDATE posts set liked=liked+1 WHERE post_id=?", postId)
	if err != nil {
		response.UnknownError(err)
		return
	}

}

// PostViewed 增加帖子浏览数
func PostViewed(context *gin.Context) {

	response := net.GetDefaultResponse()
	defer net.ResponseDefer(&response, context)

	postId, err := strconv.Atoi(context.PostForm("postId"))
	if err != nil {
		response.Error("postId格式错误")
		return
	}
	if !viewedThrottler.ThrottleVerify(crypto.GetUserFromContext(context), postId) {
		response.Error("操作限流")
		return
	}

	db, err := sqlite.OpenDB()
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer db.Close()
	_, err = db.Exec("UPDATE posts set viewed=viewed+1 WHERE post_id=?", postId)
	if err != nil {
		response.UnknownError(err)
		return
	}

}
