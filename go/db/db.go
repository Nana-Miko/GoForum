package sqlite

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var sqlitePath = "db/database"

type Pagination struct {
	Id            int
	PerPage       int
	Offset        int
	OrderByColumn string
	DESC          bool
}

type User struct {
	Email  string `json:"email"`
	Name   string `json:"name"`
	PswMd5 string `json:"pswMd5"`
	Sex    int    `json:"sex"`
}
type Posts struct {
	PostID    int     `json:"postId"`
	Content   string  `json:"content"`
	UserEmail string  `json:"userEmail"`
	Time      string  `json:"time"`
	Hot       float32 `json:"hot"`
	Liked     int     `json:"liked"`
	Viewed    int     `json:"viewed"`
	Comment   int     `json:"comment"`
	Title     string  `json:"title"`
}
type Comments struct {
	PostID    int    `json:"postId"`
	UserEmail string `json:"userEmail"`
	Content   string `json:"content"`
	Time      string `json:"time"`
	UserName  string `json:"userName"`
	Liked     int    `json:"liked"`
	CommentID int    `json:"commentId"`
}

var database *sql.DB

func openDB() (*sql.DB, error) {
	if database == nil || database.Ping() != nil {
		db, err := sql.Open("sqlite3", sqlitePath)
		if err != nil {
			return nil, err
		}
		database = db
		return db, nil
	}
	return database, nil

}

func OpenDB() (*sql.DB, error) {
	return openDB()
}

// AddCommentCount 增加帖子的评论计数
func AddCommentCount(postId int) {
	db, err := openDB()
	if err != nil {
		return
	}
	defer db.Close()
	db.Exec("UPDATE posts SET comment=comment+1 WHERE post_id=?", postId)
}

// AddLikedCount 增加帖子的点赞计数
func AddLikedCount(postId int) {
	db, err := openDB()
	if err != nil {
		return
	}
	defer db.Close()
	db.Exec("UPDATE posts SET liked=liked+1 WHERE post_id=?", postId)
}

// AddViewedCount 增加帖子的浏览计数
func AddViewedCount(postId int) {
	db, err := openDB()
	if err != nil {
		return
	}
	defer db.Close()
	db.Exec("UPDATE posts SET viewed=viewed+1 WHERE post_id=?", postId)
}
