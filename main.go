package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"myGin/config"
	"myGin/model"
	"myGin/utils"
	"net/http"
	"strings"
	"sync"
	"time"
)

var dataSource map[string]*sync.Pool
var clientsMutex sync.Mutex

type Client struct {
	Conn      *websocket.Conn
	Token     string
	UserId    string
	Topics    *utils.Set
	pingMutex sync.Mutex
	lastPing  time.Time
}

func (c *Client) updateLastPing() {
	c.pingMutex.Lock()
	defer c.pingMutex.Unlock()
	c.lastPing = time.Now()
}

var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

var clients = make(map[string]*Client)

func init() {
	dataSource = make(map[string]*sync.Pool)

	cfg := config.LoadConfigByFile()

	mysqlConfig := cfg.Mysql
	for i := 0; i < len(mysqlConfig); i++ {
		uqName := mysqlConfig[i].UqName
		if uqName == "" {
			uqName = mysqlConfig[i].Dbname
		}

		// 构建 MySQL 连接字符串
		dsn := mysqlConfig[i].Username + ":" + mysqlConfig[i].Password + "@tcp(" + mysqlConfig[i].Addr + ")/" + mysqlConfig[i].Dbname + "?charset=utf8mb4&parseTime=True&loc=Local"

		dbPool := &sync.Pool{
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

		dataSource[uqName] = dbPool
	}

	//dbConn := dataSource["zb_trx"].Get().(*gorm.DB)
	//dbConn.AutoMigrate(&model.CoinKey{})

}

func main() {

	fmt.Println("start...")
	//Init()
	r := gin.Default()
	// 启用 CORS 中间件
	r.Use(cors.Default())
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:1111"} // 允许的来源
	r.Use(cors.New(config))

	r.GET("/", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	r.POST("/login", func(context *gin.Context) {
		var requestBody = make(map[string]string)
		if err := context.BindJSON(&requestBody); err == nil {
			log.Println("requestBody", requestBody)
		}
		context.JSON(http.StatusOK, gin.H{
			"token": "Hello, World! I am TOKEN",
		})
	})

	r.GET("/ws", func(context *gin.Context) {
		handleWebSocket(context.Writer, context.Request)
	})
	// 启动心跳检测协程
	go startHeartbeat()

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

func startHeartbeat() {
	for {
		time.Sleep(10 * time.Second)

		for _, client := range clients {
			client.pingMutex.Lock()
			if time.Since(client.lastPing) > 60*time.Second {
				// 超过一定时间没有收到心跳，认为连接不可用，执行清理操作
				delete(clients, client.Token)
				client.Conn.Close()
				fmt.Printf("Client disconnected due to heartbeat timeout: %s\n", client.Token)
			}
			client.pingMutex.Unlock()
		}
	}
}

func handleWebSocket(writer gin.ResponseWriter, request *http.Request) {
	token := getTokenFromRequest(request)
	client, exist := clients[token]
	if !exist {
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			fmt.Println("Error upgrading to WebSocket:", err)
			return
		}

		client = &Client{Conn: conn, Token: token}
		clientsMutex.Lock()
		clients[token] = client
		clientsMutex.Unlock()
		fmt.Printf("Client connected: %s\n", client.Token)
	}
	// 处理 WebSocket 连接的读写操作
	go handleWebSocketConnection(client)
}

func handleWebSocketConnection(client *Client) {
	/*defer func() {
		fmt.Printf("Client disconnected: %s\n", client.Token)
		clientsMutex.Lock()
		delete(clients, client.Token)
		clientsMutex.Unlock()

		client.Conn.Close()
	}()*/

	loopFlag := true
	for loopFlag {
		messageType, p, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				fmt.Println("Unexpected close error:", err)
			}
			// 客户端断开连接，执行清理操作
			return
		}

		fmt.Printf("%s Received messageType:%d, message: %s\n", time.Now().String(), messageType, p)

		// 更新最后一次收到消息的时间，用于心跳检测
		client.updateLastPing()

		// 处理接收到的消息
		handleMessage(client, messageType, string(p))

	}
}

func handleMessage(client *Client, messageType int, receiveMsg string) {
	// 在这里处理接收到的消息
	// 你可以根据需要实现不同的业务逻辑

	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	var sendMsg = ""
	if receiveMsg == "ping" {
		sendMsg = "pong"
	} else {
		sendMsg = "消息收到了：" + receiveMsg
	}
	err := client.Conn.WriteMessage(messageType, []byte(sendMsg))
	if err != nil {
		fmt.Println("Error writing message:", err)
	}

}

func getTokenFromRequest(r *http.Request) string {
	// 这里根据实际需求获取 token
	// 你可以从请求中提取 cookie 或其他信息来生成唯一的 token
	// 这里简单返回 IP 地址作为示例
	ip := strings.Split(r.RemoteAddr, ":")[0]
	return ip
}

func QueryCoinKeys() ([]model.CoinKey, error) {
	dbConn := dataSource["zb_trx"].Get().(*gorm.DB)
	defer closeDs(dbConn, "zb_trx")
	var result []model.CoinKey
	//db.Raw("select * from coinKey").Scan(&result)

	if err := dbConn.Find(&result).Error; err != nil {
		log.Println("Failed to query CoinKeys:", err)
		return nil, err
	}
	return result, nil
}

func closeDs(db *gorm.DB, databaseName string) {
	log.Println("enter ...")
	if db != nil {
		dataSource[databaseName].Put(db)
	}
}

func bindUserAddress(data model.CoinKey) int64 {
	dbConn := dataSource["zb_trx"].Get().(*gorm.DB)
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
