package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"myGin/config"
	"myGin/model"
	"net/http"
	"sync"
)

var dsn string // 全局变量 dsn

var dbPool *sync.Pool

func init() {
	cfg := config.LoadConfigByFile()
	// 构建 MySQL 连接字符串
	dsn = cfg.Mysql.Username + ":" + cfg.Mysql.Password + "@tcp(" + cfg.Mysql.Addr + ")/" + cfg.Mysql.Dbname + "?charset=utf8mb4&parseTime=True&loc=Local"

	dbPool = &sync.Pool{
		New: func() interface{} {
			// 连接 MySQL 数据库
			db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
			if err != nil {
				log.Println("Failed to connect to database:", err)
			} else {
				db.Logger.LogMode(logger.Info)
			}
			// 获取底层的 *sql.DB 对象
			sqlDB, err := db.DB()
			if err != nil {
				panic(err.Error())
			}
			// 设置连接池参数
			sqlDB.SetMaxOpenConns(100)
			sqlDB.SetMaxIdleConns(10)

			log.Println("Database connection successful")

			return db
		},
	}

	dbConn := dbPool.Get().(*gorm.DB)
	dbConn.AutoMigrate(&model.CoinKey{})

}

func main() {

	fmt.Println("start...")
	//Init()
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

	r.GET("/coinKeys", func(c *gin.Context) {
		rs, err := QueryCoinKeys()
		if err == nil {
			c.JSON(http.StatusOK, rs)
		}
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

	r.POST("/bindUserAddress", func(c *gin.Context) {
		var requestBody model.CoinKey
		if err := c.BindJSON(&requestBody); err == nil {
			count := bindUserAddress(requestBody)
			fmt.Println(count)
			c.JSON(http.StatusOK, gin.H{
				"updateCount": count,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"errorParam": c.Request.Body,
			})
		}
	})

	r.Run(":8084")
}

func QueryCoinKeys() ([]model.CoinKey, error) {
	dbConn := dbPool.Get().(*gorm.DB)
	var result []model.CoinKey
	//db.Raw("select * from coinKey").Scan(&result)

	if err := dbConn.Find(&result).Error; err != nil {
		log.Println("Failed to query CoinKeys:", err)
		dbPool.Put(dbConn)
		return nil, err
	}
	dbPool.Put(dbConn)
	return result, nil
}

func bindUserAddress(data model.CoinKey) int64 {
	dbConn := dbPool.Get().(*gorm.DB)
	// 定义需要更新的字段
	updateFields := map[string]interface{}{
		"userName": data.UserName,
		"userId":   data.UserId,
		// 添加其他需要更新的字段...
	}
	rs := dbConn.Model(&model.CoinKey{}).Where("id=?", data.Id).Updates(updateFields)
	count := rs.RowsAffected
	return count
}
