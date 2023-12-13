package model

import "time"

type CoinKey struct {
	Id              int64     `gorm:"column:id;primaryKey" json:"id"`
	KeyPre          string    `gorm:"column:keyPre" json:"keyPre"`
	UserId          int64     `gorm:"column:userId" json:"userId"`
	UserName        string    `gorm:"column:userName" json:"userName"`
	Wallet          string    `gorm:"column:wallet" json:"wallet"`
	Coins           int64     `gorm:"column:coins" json:"coins"`
	CreateTime      time.Time `gorm:"column:createTime" json:"createTime"`
	UsedTimes       uint      `gorm:"column:usedTimes" json:"usedTimes"`
	Tag             uint      `gorm:"column:tag" json:"tag"`
	MerchantOrderNo string    `gorm:"column:merchantOrderNo" json:"merchantOrderNo"`
	VId             uint      `gorm:"column:vId" json:"VId"`
	VerifyTime      int64     `gorm:"column:verifyTime" json:"verifyTime"`
}

func (CoinKey) TableName() string {
	return "coinKey"
}
