package test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"myGin/config"
	"myGin/dao"
	"myGin/model"
	"myGin/utils"
	"reflect"
	"sync"
	"testing"
	"time"
)

var dsn string // 全局变量 dsn

func init() {
	cfg := config.LoadConfigByFile()
	// 构建 MySQL 连接字符串
	fmt.Println(cfg)
}

func TestQueryMysqlDb(t *testing.T) {
	// 连接 MySQL 数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("Failed to connect to database:", err)
	} else {
		db.Logger.LogMode(logger.Info)
	}

	log.Println("Database connection successful")

	db.AutoMigrate(&model.CoinKey{})

	var result []model.CoinKey
	db.Raw("select * from coinKey").Scan(&result)

	fmt.Println(result)

	// 关闭数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database connection:", err)
	}
	sqlDB.Close()
}

func TestConfig(t *testing.T) {
	fmt.Println("aa")
	//cfg := config.GetConfig()

	//fmt.Println(cfg.Mysql.Username)

	cfg := config.LoadConfigByFile()
	fmt.Println(cfg.Mysql[0].Username)

}

func TestDB(t *testing.T) {
	db, err := sql.Open("mysql", "root:iPYDU0o3MRQOreEW@tcp(172.16.100.150:3309)/zb_trx")
	if err != nil {
		panic(err.Error())
	}
	// 确保连接被关闭
	defer db.Close()

	// 查询数据库
	rows, err := db.Query("SELECT * FROM coinKey")

	if err != nil {
		panic(err.Error())
	}

	fmt.Println(rows.Columns())

	/*var a []string

	a = rows.Columns()*/

	// 遍历结果集
	/*for rows.Next() {
		var user_id int
		var username string
		var email string
		err = rows.Scan(&user_id, &username, &email)

		if err != nil {
			panic(err.Error())
		}

		fmt.Printf("ID: %d, Name: %s, Email: %s\n", user_id, username, email)
	}*/
}

type QueryModel interface {
	TableName() string
}
type User struct {
	ID   uint   `gorm:"primaryKey;column:id" column:"id" json:"id"`
	Name string `gorm:"column:name" column:"name" json:"name"`
	Age  int    `gorm:"column:age" column:"age" json:"age"`
}

type UserReq struct {
	User
	NameLike  string   `json:"nameLike,omitempty"`
	AgeStart  int      `json:"ageStart,omitempty"`
	AgeEnd    int      `json:"ageEnd,omitempty"`
	AgeMin    int      `json:"ageMin,omitempty"`
	AgeMax    int      `json:"ageMax,omitempty"`
	NameList  []string `json:"nameList,omitempty"`
	AgeList   []string `json:"ageList,omitempty"`
	AgeNqList []string `json:"ageNqList,omitempty"`
}

func (u User) TableName() string {
	//TODO implement me
	return "users"
	//panic("implement me")
}

type Book struct {
	ID         uint      `gorm:"primaryKey"`
	Name       string    `gorm:"column:name"`
	CreateTime time.Time `gorm:"column:create_time"`
}

func (Book) TableName() string {
	return "book"
}

func TestSqlLiteCreateTable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("../sqllite/sqlLite-database.db"), &gorm.Config{})
	if err != nil {
		log.Println("创建失败：", err)
	}

	db.AutoMigrate(&Book{})

}

func TestSqlLiteInsert(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("../sqllite/sqlLite-database.db"), &gorm.Config{})
	if err != nil {
		log.Println("创建失败：", err)
	}

	newUser := User{
		Age:  23,
		Name: "allen",
	}

	tx := db.Create(&newUser)
	if tx.Error != nil {
		log.Fatal("Failed to insert record:", tx.Error)
	}

	log.Printf("Record inserted successfully, user ID:%d, effectRow:%d", newUser.ID, tx.RowsAffected)
}

func TestSqlLiteInsert2(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("../sqllite/sqlLite-database.db"), &gorm.Config{})
	if err != nil {
		log.Println("创建失败：", err)
	}

	newUser := []User{
		{Age: 24, Name: "allen2"},
		{Age: 25, Name: "allen3"},
		{Age: 26, Name: "allen4"},
	}

	tx := db.CreateInBatches(&newUser, 3)
	if tx.Error != nil {
		log.Fatal("Failed to insert record:", tx.Error)
	}

	log.Printf("Record inserted successfully, user ID:%d, effectRow:%d", newUser[1].ID, tx.RowsAffected)

}

func TestSqlLiteDel(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("../sqllite/sqlLite-database.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("error:", err)
	}
	db.Delete(&User{}, 1)

	rs := db.Where("age < ?", 24).Delete(&User{})
	log.Println("rs", rs.RowsAffected)
}

func TestSqlLiteUpdate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("../sqllite/sqlLite-database.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("error:", err)
	}
	tx1 := db.Begin()
	tx1.Rollback()
	tx1.Commit()

	tx := db.Where("id = ?", 5).Updates(User{Name: "zl239"})
	if tx.Error != nil {
		fmt.Println("tx.Error:", tx.Error)
	}

	log.Println(tx.RowsAffected)
}

func TestSqlLiteQuery(t *testing.T) {

	//通用封装
	// JSON 字符串
	jsonData := `{"nameLike": "zl", "ageList":[24,25,26], "Age":"26", "AgeMax":"100", "AgeMin":0}`
	jsonData = `{ "nameNqList":["a","b"],"ageList":[24,25,26]}`
	// 解析 JSON 数据到结构体
	var paramMap map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &paramMap)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	orderMap := map[string]bool{
		"Age":    true,
		"Name":   false,
		"ID":     true,
		"121age": true,
	}
	fmt.Println(orderMap)
	var targetList []User
	dao.Query(nil, User{}, paramMap, &targetList, orderMap)
	fmt.Println("users2", targetList)
}

func TestToStr(t *testing.T) {
	//arr := []int{1, 2, 3, 4, 5}
	strarr := []string{"1", "2"}
	str := utils.SliceToInClause(strarr)
	fmt.Println(str)
}

/**

if containKey(paramMap, fieldName+"List") {
	if isEmptyCollection(paramMap[fieldName+"List"]) {
		continue
	}

	sql = sql + " and " + fieldName + "in " + sliceToInClause(paramMap[fieldName+"List"].([]interface{}))
}
*/

func TestTypeField(t *testing.T) {
	// 使用反射获取结构体信息
	// JSON 字符串
	jsonData := `{"name": "zl239", "Age": 26, "age"："11"}`

	// 解析 JSON 数据到结构体
	var reqUser map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &reqUser)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	userInstance := reflect.New(reflect.TypeOf(User{})).Elem().Interface().(User)

	structType := reflect.TypeOf(userInstance)

	// 遍历结构体的字段
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// 打印字段名和值
		fmt.Printf("Field: %s, Value: %v\n", field.Name, "val---")
	}
}

func TestGetDefaultVal(t *testing.T) {
	paramMap := map[string]interface{}{
		"1":  "2",
		"aa": "b",

		"pageIndex": "-1",
		"pageNum":   "-123",
		"pageNo":    "-345",

		"pageSize": "100000000",
	}
	pageIndex := dao.GetPageIndex(paramMap)
	pageSize := dao.GetPageSize(paramMap)

	fmt.Println("pageIndex:", pageIndex)
	fmt.Println("pageSize:", pageSize)

	orderMap := map[string]bool{
		"Age":    true,
		"Name":   false,
		"ID":     true,
		"121age": true,
	}
	s := dao.GetOrderBySql(User{}, orderMap, pageIndex, pageSize)
	fmt.Println("$$$$$$$ " + s)
}

func TestGetColNameByFieldName(t *testing.T) {
	a := dao.GetTableColumnNameByFieldName(User{}, "Name")
	fmt.Println(a)

	fmt.Println(dao.GetPrimaryKeyJsonName(User{}))

	var user User
	dao.SelectById(nil, User{}, 5, &user)
	fmt.Println(user)
}

func TestSelectList(t *testing.T) {
	model := User{
		Age: 26,
	}

	var users []User
	dao.SelectList(nil, model, &users)
	fmt.Println(users)
}

// 示例结构体
type MyStruct struct {
	StringField string
	IntField    int
	FloatField  float64
}

// 将对象转换为 map，去掉零值或空值的字段
func StructToMap(obj interface{}) map[string]interface{} {
	resultMap := make(map[string]interface{})

	// 获取对象的反射类型和值
	objType := reflect.TypeOf(obj)
	objValue := reflect.ValueOf(obj)

	// 遍历对象的字段
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		fieldValue := objValue.Field(i).Interface()

		// 添加条件判断，只将非零值或非空值的字段放入 map
		// 判断字段的值是否为零值并且没有设置值
		if isZeroValue(fieldValue) && isEmptyStringValue(fieldValue) {
			continue
		}
		// 将非零值或非空字符串的字段添加到 map
		resultMap[field.Name] = fieldValue
	}

	return resultMap
}

// 判断值是否为其类型的零值
func isZeroValue(value interface{}) bool {
	zeroValue := reflect.Zero(reflect.TypeOf(value)).Interface()
	return reflect.DeepEqual(value, zeroValue)
}

// 判断字符串是否为空字符串
func isEmptyStringValue(value interface{}) bool {
	if str, ok := value.(string); ok {
		return str == ""
	}
	return false
}

func TestConvertMap(t *testing.T) {
	// 示例结构体实例
	myStruct := MyStruct{
		StringField: "",
		FloatField:  0.0,
	}

	// 将结构体转换为 map
	resultMap := StructToMap(myStruct)

	// 输出结果
	fmt.Println(resultMap)
}

func TestSet(t *testing.T) {
	mySet := make(utils.Set)
	mySet.Add("ddd")
	mySet.Add("ccc")
	mySet.Add("dfsdf")
	mySet.Remove("ddd")
	exist := mySet.Contains("ddd")
	fmt.Println("exist", exist)
}

func TestWebSocketClient(t *testing.T) {
	serverURL := "ws://localhost:8084/ws"
	// 使用默认的gorilla/websocket配置
	dialer := websocket.DefaultDialer

	//连接WebSocket服务器
	conn, _, err := dialer.Dial(serverURL, nil)
	if err != nil {
		fmt.Println("Error connecting to WebSocket server:", err)
		return
	}
	defer conn.Close()

	go handlerServerMsg(conn)

	// 在主goroutine中发送消息给服务器
	i := 0
	flag := true
	for flag {
		var message string

		i = i + 1
		message = fmt.Sprintf("hello, 你好！[%d]", i)
		fmt.Printf("start send msg to Server, msg:%s\n", message)
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			fmt.Println("Error sending message:", err)
			flag = false
			break
		}

		// 等待一些时间以确保服务器有足够的时间处理关闭请求
		time.Sleep(2 * time.Second)

	}

}

func handlerServerMsg(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			return
		}
		fmt.Printf("Received message from server: %s\n", message)

	}
}

func TestWaitGroupWait(t *testing.T) {
	var wg sync.WaitGroup // 申明一个协程等待组
	for i := 1; i <= 5; i++ {
		wg.Add(1)         // 在启动 goroutine 前增加计数器
		go worker(i, &wg) // 启动 goroutine
	}

	wg.Wait() // 等待所有 goroutine 完成
	fmt.Println("all worker end ")

}

// 这里必须使用指针，因为在go中结构体属于值类型，如果使用值传递，就会拷贝一个副本给方法
// go语言中引用类型有：map，数组切片，通道，接口
// 值类型有：基本类型int float64 bool 复数类型complex64 complex128 字符类型byte run
// 字符串string  数组Array 结构体
func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("worker start ", id)
	time.Sleep(time.Second)
	fmt.Println("worker end", id)
}

func TestArr(t *testing.T) {
	var arr [3]int
	arr[0] = 1
	arr[1] = 2
	arr[2] = 3

	arr2 := [3]int{1, 2, 3}

	for i, v := range arr2 {
		fmt.Printf("i: %d v: %d", i, v)
	}
}

func TestSlice(t *testing.T) {
	var slice []int
	fmt.Println(slice)
	//slice[1] = 1
	//slice[100] = 100
	//slice[200] = 22
	slice = append(slice, 100)
	slice = append(slice, 100)
	slice = append(slice, 100)
	slice = append(slice, 100)
	slice = append(slice, 100)
	slice = append(slice, 100)

	slice[5] = 50
	fmt.Println(slice)
}

func TestChan1(t *testing.T) {
	var chan1 = make(chan int)
	go func() {
		defer close(chan1)
		for i := 0; i < 5; i++ {
			chan1 <- i
		}
	}()

	for val := range chan1 {
		fmt.Println(val)
	}
}

func TestSelectChan1(t *testing.T) {
	var chan1 = make(chan string)
	var chan2 = make(chan string)

	go func() {
		time.Sleep(time.Second * 2)
		chan1 <- "chan1 msg!!"

		chan2 <- "chan2 msg!!"

	}()

	for {
		fmt.Println("time:", time.Now().String())
		select {
		case val := <-chan2:
			fmt.Println(val)
			fmt.Println("chan2 done")

		case val := <-chan1:
			fmt.Println(val)

		case <-time.After(time.Second * 3):

		}
	}

}

func TestBlockChan(t *testing.T) {
	var chan1 = make(chan string)
	fmt.Println("start1")
	go func() { chan1 <- "aaa" }()
	fmt.Println("start2")
	val := <-chan1
	fmt.Println("val", val)
	fmt.Println("end")
}

func TestBlockChan2(t *testing.T) {

	// 创建一个无缓冲的通道
	ch := make(chan int)

	// 启动一个 goroutine 接收数据
	go func() {
		fmt.Println("Receiving data...")
		data := <-ch // 阻塞，直到有发送者发送数据
		fmt.Println("Data received:", data)
	}()

	// 模拟一些其他工作
	time.Sleep(2 * time.Second)

	// 主 goroutine 发送数据
	fmt.Println("Sending data...")
	ch <- 42 // 阻塞，直到有接收者接收数据
	fmt.Println("Data sent.")

	// 等待一些时间，以便接收 goroutine 完成
	time.Sleep(2 * time.Second)

}

func TestCacheChan1(t *testing.T) {
	var wg sync.WaitGroup
	ch := make(chan int, 3)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			ch <- i
		}
		close(ch)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// 通过 ok 判断通道是否关闭
			if val, ok := <-ch; ok {
				fmt.Println("Received:", val)
			} else {
				fmt.Println("Channel closed. Exiting.")
				return
			}
		}
	}()

	fmt.Println("end1")
	wg.Wait()
	fmt.Println("end2")
}

func TestStopFor(t *testing.T) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ctx.Done()
	ctx.Value("qqq")
	ch := make(chan int, 3)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			ch <- i
		}
		close(ch)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			fmt.Println("for start")
			val, ok := <-ch
			if !ok {
				fmt.Println("Channel closed. Exiting.")
				return
			}
			fmt.Println("Received:", val)
			/*select {
			case val, ok := <-ch:
				if !ok {
					fmt.Println("Channel closed. Exiting.")
					return
				}
				fmt.Println("Received:", val)
				/*case <-ctx.Done():
				fmt.Println("Context cancelled. Exiting.")
				cancel()
				return
			}*/
		}
	}()

	wg.Wait()
}

func TestBlockCacheChan(t *testing.T) {
	ch := make(chan int, 3)
	for i := 0; i < 2; i++ {
		ch <- i
	}
	fmt.Println("aaaaa")
}

func TestContextWithCancel(t *testing.T) {
	ctx2, cancel2 := context.WithTimeout(context.Background(), 4*time.Second)
	fmt.Println("111")
	defer cancel2()
	fmt.Println("222")
	val := <-ctx2.Done()
	fmt.Println("333", val)

	ctx, cancel := context.WithCancel(context.Background())

	ctx.Value("sss")
	go func() {
		time.Sleep(time.Second * 3)
		fmt.Println("start cancel", time.Now().Format("2006-04-02 15:04:05.000.Z07"))
		cancel()
	}()
	select {
	case <-ctx.Done():
		fmt.Println("Context canceled", time.Now().String())
	}
	fmt.Println("end")

	//这个例子中，当执行cancel时，会触发ctx.Done()通道关闭，不再阻塞
}
