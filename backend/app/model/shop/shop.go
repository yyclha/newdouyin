package shop

import (
	"douyin-backend/app/global/variable"
	"douyin-backend/app/model"
	"gorm.io/gorm"
)

// GoodsModel 封装商品相关的查询操作。
type GoodsModel struct {
	*gorm.DB   `gorm:"-" json:"-"`
	ID         int64   `json:"id"`         // bigint
	Name       string  `json:"name"`       // varchar(255)
	Cover      string  `json:"cover"`      // varchar(255)
	Imgs       string  `json:"imgs"`       // text
	IsLowPrice bool    `json:"isLowPrice"` // tinyint(1)
	Discount   string  `json:"discount"`   // varchar(100)
	Sold       float64 `json:"sold"`       // float
	Price      float64 `json:"price"`      // float
	RealPrice  float64 `json:"real_price"` // float
}

// CreateShopFactory 创建带数据库连接的商品模型实例。
func CreateShopFactory(sqlType string) *GoodsModel {
	return &GoodsModel{DB: model.UseDbConn(sqlType)}
}

// GetShopRecommended 分页查询推荐商品列表。
func (u *GoodsModel) GetShopRecommended(uid, pageNo, pageSize int64) (slice []Goods, total int64, ok bool) {
	sql1 := `
		SELECT *
		from tb_goods as tu
		LIMIT ? OFFSET ?;`
	sql2 := `
		SELECT COUNT(*)
		FROM tb_goods as a;
		`

	offset := pageNo * pageSize
	result1 := u.Raw(sql2).Count(&total)
	result2 := u.Raw(sql1, pageSize, offset).Find(&slice)

	if result1.Error != nil || result2.Error != nil {
		variable.ZapLog.Error("GetShopRecommended SQL代码执行出错!")
		ok = false
		return
	}
	ok = true
	return
}
