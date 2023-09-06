package commentAPI

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

func init() {
	// 3600000毫秒=1小时
	// 1秒=1000毫秒
	likedThrottler = throttle.NewThrottler(1000 * 10)
}

// CommentCreate 发布评论
func CommentCreate(context *gin.Context) {
	response := net.GetDefaultResponse()
	defer net.ResponseDefer(&response, context)

	user := crypto.GetUserFromContext(context)

	postID, err := strconv.Atoi(context.PostForm("postId"))
	if err != nil {
		response.Error("postId格式错误")
		return
	}
	content := context.PostForm("content")

	db, err := sqlite.OpenDB()
	if err != nil {
		response.UnknownError(err)
		return
	}

	comment := sqlite.Comments{
		PostID:    postID,
		UserEmail: user.Email,
		Content:   content,
		Time:      util.CurrentTimeStampStrMilli(),
		UserName:  user.Name,
		Liked:     0,
	}

	res, err := db.Exec("INSERT INTO comments(post_id,user_email,content,time,user_name) VALUES (?,?,?,?,?)", comment.PostID, comment.UserEmail, comment.Content, comment.Time, comment.UserName)
	if err != nil {
		response.UnknownError(err)
		return
	}

	lastID, _ := res.LastInsertId()
	comment.CommentID = int(lastID)

	response.AppendResponseDefer(func() {
		sqlite.AddCommentCount(postID)
	})

	response.Successful(comment)

}

// CommentSearch 查询评论
func CommentSearch(context *gin.Context) {
	response := net.GetDefaultResponse()
	defer net.ResponseDefer(&response, context)

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
	pagination := api.PaginationQueryParse(context)

	query := fmt.Sprintf("SELECT * FROM comments WHERE post_id=? ORDER BY %s LIMIT ? OFFSET ?", pagination.OrderByColumn)
	if pagination.DESC {
		query = fmt.Sprintf("SELECT * FROM comments WHERE post_id=? ORDER BY %s DESC LIMIT ? OFFSET ?", pagination.OrderByColumn)
	}

	fmt.Println(pagination)

	stmt, err := db.Prepare(query)
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query(postId, pagination.PerPage, pagination.Offset)
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer rows.Close()
	comments := make([]sqlite.Comments, 0)
	for rows.Next() {
		var comment sqlite.Comments
		err := rows.Scan(&comment.PostID, &comment.UserEmail, &comment.Content, &comment.Time, &comment.UserName, &comment.Liked, &comment.CommentID)
		if err != nil {
			response.UnknownError(err)
			return
		}
		comments = append(comments, comment)
	}
	response.Successful(comments)
}

// CommentLiked 点赞评论
func CommentLiked(context *gin.Context) {

	response := net.GetDefaultResponse()
	defer net.ResponseDefer(&response, context)

	commentId, err := strconv.Atoi(context.PostForm("commentId"))
	if err != nil {
		response.Error("commentId格式错误")
		return
	}
	if !likedThrottler.ThrottleVerify(context.MustGet("user").(sqlite.User), commentId) {
		response.Error("操作限流")
		return
	}

	db, err := sqlite.OpenDB()
	if err != nil {
		response.UnknownError(err)
		return
	}
	defer db.Close()
	_, err = db.Exec("UPDATE comments set liked=liked+1 WHERE comment_id=?", commentId)
	if err != nil {
		response.UnknownError(err)
		return
	}

}
