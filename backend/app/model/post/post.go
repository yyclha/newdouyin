package post

import (
	"douyin-backend/app/global/variable"
	"douyin-backend/app/model"
	"github.com/goccy/go-json"
	"gorm.io/gorm"
)

// PostModel 封装图文动态相关的查询操作。
type PostModel struct {
	*gorm.DB  `gorm:"-" json:"-"`
	ID        string          `json:"id"`         // varchar(100)
	ModelType string          `json:"model_type"` // varchar(100)
	NoteCard  json.RawMessage `json:"note_card"`  // json
}

// CreatePostFactory 创建带数据库连接的图文动态模型实例。
func CreatePostFactory(sqlType string) *PostModel {
	return &PostModel{DB: model.UseDbConn(sqlType)}
}

// GetPostRecommended 分页查询推荐图文动态列表。
func (u *PostModel) GetPostRecommended(uid, pageNo, pageSize int64) (slice []Post, total int64, ok bool) {
	sql1 := `
		SELECT *
		from tb_posts as tu
		LIMIT ? OFFSET ?;`
	sql2 := `
		SELECT COUNT(*)
		FROM tb_posts as a;
		`

	offset := pageNo * pageSize
	result1 := u.Raw(sql2).Count(&total)
	result2 := u.Raw(sql1, pageSize, offset).Find(&slice)

	if result1.Error != nil || result2.Error != nil {
		variable.ZapLog.Error("GetPostRecommended SQL代码执行出错!")
		ok = false
		return
	}
	ok = true
	return
}
