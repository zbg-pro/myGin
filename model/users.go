package model

type User struct {
	ID         uint   `gorm:"primaryKey;column:id" column:"id" json:"id"`
	Name       string `gorm:"column:name" column:"name" json:"name"`
	Age        int    `gorm:"column:age" column:"age" json:"age"`
	Password   string `gorm:"column:password" column:"password" json:"password"`
	CreateTime string `gorm:"column:create_time" column:"create_time" json:"createTime"`
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
