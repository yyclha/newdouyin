package user

import (
	"gorm.io/gorm"
)

// Account 定义登录账号表对应的数据结构。
type Account struct {
	*gorm.DB `gorm:"-" json:"-"`
	UID      int64  `json:"uid"`      // bigint
	Nickname string `json:"nickname"` // varchar(100)
	Phone    string `json:"phone"`    // varchar(11)
	Password string `json:"password"` // varchar(128)
}

// AwemeStatusModel 汇总当前用户的关注、点赞和收藏状态数据。
type AwemeStatusModel struct {
	Attentions []int64
	Likes      []string
	Collects   []string
}
