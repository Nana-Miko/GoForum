package main

import (
	"GoForum/go/net/route"
	"github.com/gin-gonic/gin"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		runHttpServer("0.0.0.0:8081")
	}()
	wg.Wait()
}

func runHttpServer(host string) {

	ginServer := gin.Default()

	// 路由注册
	route.RouterGroup(ginServer)

	ginServer.Run(host)

}
