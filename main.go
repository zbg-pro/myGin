package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"myGin/config"
	"net/http"
)

func main() {

	fmt.Println("start...")

	r := gin.Default()

	r.GET("/", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	// 添加一个 GET 请求路由
	r.GET("/AA", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "AA, Hello, World!",
		}) //{"message":"AA, Hello, World!"}%
	})

	r.GET("someJSON", func(context *gin.Context) {

		data := map[string]interface{}{
			"lang": "GO语言",
			"tag":  "<br/>",
		}

		context.AsciiJSON(http.StatusOK, data) //{"lang":"GO\u8bed\u8a00","tag":"\u003cbr/\u003e"}
	})

	r.GET("JSONP", func(context *gin.Context) {
		data := map[string]interface{}{
			"foo": "bar",
		}
		context.JSONP(http.StatusOK, data)
	})

	cfg := config.LoadConfigByFile()
	fmt.Println(cfg.Mysql.Username)

	r.Run(":8084")
}
