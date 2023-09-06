package crypto

import (
	sqlite "GoForum/go/db"
	"fmt"
	"strings"
	"time"
)

var manager *TokenManager

func init() {
	tm := GetTokenManager()
	initTokenManager(tm)
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				_, err := tm.cleanupInvalidTokens()
				if err != nil {
					fmt.Println("在尝试清理失效Token的时候发生了一个错误：" + err.Error())
				}
			}
		}
	}()
}

func GetTokenManager() *TokenManager {
	if manager == nil {
		manager = &TokenManager{ExpireTokenIdMap: map[string]int64{}}
	}
	return manager
}

type TokenManager struct {
	ExpireTokenIdMap map[string]int64
}

// IsExpire 判断Token是否失效
func (tm *TokenManager) IsExpire(tokenString string) bool {
	_, exists := tm.ExpireTokenIdMap[tokenString]
	return exists
}

// Invalidate 使Token失效
func (tm *TokenManager) Invalidate(tokenString string) error {
	if _, exists := tm.ExpireTokenIdMap[tokenString]; exists {
		return nil
	}
	id, err := appendInvalidateTokenToDB(tokenString)
	if err != nil {
		return err
	}
	tm.ExpireTokenIdMap[tokenString] = id
	return nil
}

// 将失效Token保存至DB
func appendInvalidateTokenToDB(tokenString string) (int64, error) {
	db, err := sqlite.OpenDB()
	if err != nil {
		return -1, err
	}
	defer db.Close()
	res, err := db.Exec("INSERT INTO expire_token (token) VALUES (?)", tokenString)
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return id, nil
}

// 批量清除DB中的Token
func removeInvalidateTokenToDB(tokensId []int64) (int64, error) {
	db, err := sqlite.OpenDB()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// 构建IN子句参数
	params := make([]interface{}, len(tokensId))
	for i, id := range tokensId {
		params[i] = id
	}

	// 创建IN子句占位符字符串
	placeholders := strings.Repeat("?,", len(tokensId))
	placeholders = placeholders[:len(placeholders)-1] // 移除末尾的逗号

	// 构建SQL语句
	sqlStmt := fmt.Sprintf("DELETE FROM expire_token WHERE id IN (%s)", placeholders)

	// 执行SQL删除操作
	res, err := db.Exec(sqlStmt, params...)
	if err != nil {
		return 0, err
	}

	// 获取删除的行数
	numRowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return numRowsAffected, nil
}

// 加载DB中的Token
func loadTokenFromDB(tm *TokenManager) error {
	db, err := sqlite.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := db.Query("SELECT * FROM expire_token")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		tokenGet := struct {
			Token string
			Id    int64
		}{}
		err := rows.Scan(&tokenGet.Id, &tokenGet.Token)
		if err != nil {
			return err
		}
		tm.ExpireTokenIdMap[tokenGet.Token] = tokenGet.Id
	}
	return nil
}

// 清理过期的失效Token
func (tm *TokenManager) cleanupInvalidTokens() (int64, error) {
	fmt.Println("执行失效过期Token清理...")
	var tokensToDelete []string
	for tokenString, _ := range tm.ExpireTokenIdMap {
		token, err := TokenParse(tokenString)
		if err != nil {
			tokensToDelete = append(tokensToDelete, tokenString)
			continue
		}
		if !token.Valid {
			tokensToDelete = append(tokensToDelete, tokenString)
		}
	}

	if len(tokensToDelete) <= 0 {
		return 0, nil
	}

	tokensId := make([]int64, len(tokensToDelete))

	for i, tokenString := range tokensToDelete {
		tokensId[i] = tm.ExpireTokenIdMap[tokenString]
		delete(tm.ExpireTokenIdMap, tokenString)
	}

	dbRemoveNum, err := removeInvalidateTokenToDB(tokensId)

	if err != nil {
		return 0, err
	}

	fmt.Println(fmt.Sprintf("执行失效Token过期清理完成，缓存中已失效%d个，从DB删除%d个", len(tokensToDelete), dbRemoveNum))
	return dbRemoveNum, nil
}

// 初始化TokenManager
func initTokenManager(tm *TokenManager) {
	err := loadTokenFromDB(tm)
	if err != nil {
		panic("TokenManager初始化失败：" + err.Error())
	}
	deleteNum, err := tm.cleanupInvalidTokens()
	if err != nil {
		panic("TokenManager初始化失败：" + err.Error())
	}
	fmt.Println(fmt.Sprintf("TokenManager初始化完成，已加载%d个Token，清理%d个失效过期Token", len(tm.ExpireTokenIdMap), deleteNum))
}
