package test

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	jsonData := `{"nameLike": "zl", "ageList":[24,25,26], "Age":"26"}`

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
