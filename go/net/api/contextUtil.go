package api

import (
	sqlite "GoForum/go/db"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

// PaginationQueryParse 解析分页查询
func PaginationQueryParse(context *gin.Context) sqlite.Pagination {
	pagination := sqlite.Pagination{
		Id:-10086,
		PerPage:       10,
		Offset:        0,
		OrderByColumn: "time",
		DESC:          false,
	}
	page, err1 := strconv.Atoi(context.PostForm("page"))
	perPage, err2 := strconv.Atoi(context.PostForm("perPage"))
	if err1 == nil && err2 == nil {
		pagination.PerPage = perPage
		pagination.Offset = (page - 1) * perPage
	}
	id, err := strconv.Atoi(context.PostForm("id"))
	if err==nil {
		pagination.Id = id
	}
	order := context.PostForm("order")
	if order != "" && !strings.Contains(order, " ") {
		pagination.OrderByColumn = order
	}
	orderRule := context.PostForm("orderRule")
	if orderRule == "DESC" {
		pagination.DESC = true
	}

	return pagination
}
