package log

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"os"
)

func InitLogger() {
	// 配置日志输出到控制台
	gin.DefaultWriter = io.MultiWriter(os.Stdout)

	// 配置日志输出到文件
	file, err := os.OpenFile("../gin.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
	} else {
		gin.DefaultWriter = io.MultiWriter(os.Stdout, file)
		log.SetOutput(io.MultiWriter(os.Stdout, file))
	}
}
