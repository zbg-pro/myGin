package test

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"myGin/config"
	"testing"
)

func TestConfig(t *testing.T) {
	fmt.Println("aa")
	//cfg := config.GetConfig()

	//fmt.Println(cfg.Mysql.Username)

	cfg := config.LoadConfigByFile()
	fmt.Println(cfg.Mysql.Username)

}

func TestDB(t *testing.T) {
	db, err := sql.Open("mysql", "root:iPYDU0o3MRQOreEW@tcp(172.16.100.130:3306)/zb_op")
	if err != nil {
		panic(err.Error())
	}
	// 确保连接被关闭
	defer db.Close()

	// 查询数据库
	rows, err := db.Query("SELECT * FROM sys_user")

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
