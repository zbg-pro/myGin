package model

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type User struct {
	ID         uint      `gorm:"primaryKey;column:id" column:"id" json:"id"`
	Name       string    `gorm:"column:name" column:"name" json:"name"`
	Age        int       `gorm:"column:age" column:"age" json:"age"`
	Password   string    `gorm:"column:password" column:"password" json:"password"`
	CreateTime time.Time `gorm:"column:create_time" column:"create_time" json:"createTime"`
}

func (u User) TableName() string {
	return "users"
}

type UserReq struct {
	*User
	NameLike        string   `json:"nameLike,omitempty"`
	AgeStart        int      `json:"ageStart,omitempty"`
	AgeEnd          int      `json:"ageEnd,omitempty"`
	AgeMin          int      `json:"ageMin,omitempty"`
	AgeMax          int      `json:"ageMax,omitempty"`
	NameList        []string `json:"nameList,omitempty"`
	AgeList         []string `json:"ageList,omitempty"`
	AgeNqList       []string `json:"ageNqList,omitempty"`
	CreateTimeStart string   `json:"createTimeStart,omitempty"`
	CreateTimeEnd   string   `json:"createTimeEnd,omitempty"`
	CreateTimeMin   string   `json:"createTimeMin,omitempty"`
	CreateTimeMax   string   `json:"createTimeMax,omitempty"`
}

var jwtKey = []byte("your-secret-key")

// ParseToken 解析和验证 JWT
func ParseToken(tokenString string) (*jwt.StandardClaims, error) {
	// 解析 JWT
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	// 验证 JWT
	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("Invalid token")
	}

	return claims, nil
}

// GenerateToken 生成 JWT
func (u *User) GenerateToken() (string, error) {
	// 设置 JWT 的有效期为一小时
	expirationTime := time.Now().Add(time.Hour)

	// 创建一个 JWT 的声明
	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		IssuedAt:  time.Now().Unix(),
		Subject:   fmt.Sprintf("%d", u.ID),
	}

	// 创建 JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名 JWT
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
