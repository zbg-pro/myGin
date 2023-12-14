package model

type User struct {
	ID   uint   `gorm:"primaryKey;column:id" column:"id" json:"id"`
	Name string `gorm:"column:name" column:"name" json:"name"`
	Age  int    `gorm:"column:age" column:"age" json:"age"`
}
