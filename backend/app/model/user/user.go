package user

import (
	"douyin-backend/app/global/variable"
	"douyin-backend/app/model"
	"douyin-backend/app/model/video"
	"douyin-backend/app/utils/md5_encrypt"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

// UserModel 封装用户相关的数据库读写操作。
type UserModel struct {
	*gorm.DB                `gorm:"-" json:"-"`
	UID                     int64           `json:"uid"`                       // bigint
	ShortID                 int             `json:"short_id"`                  // int
	UniqueID                string          `json:"unique_id"`                 // varchar(255)
	Gender                  string          `json:"gender"`                    // char(1)
	UserAge                 int             `json:"user_age"`                  // int
	Nickname                string          `json:"nickname"`                  // varchar(100)
	Country                 string          `json:"country"`                   // varchar(100)
	Province                string          `json:"province"`                  // varchar(100)
	District                string          `json:"district"`                  // varchar(255)
	City                    string          `json:"city"`                      // varchar(255)
	Signature               string          `json:"signature"`                 // text
	IPLocation              string          `json:"ip_location"`               // varchar(100)
	BirthdayHideLevel       int             `json:"birthday_hide_level"`       // int
	CanShowGroupCard        int             `json:"can_show_group_card"`       // int
	AwemeCount              int64           `json:"aweme_count"`               // bigint
	TotalFavorited          int64           `json:"total_favorited"`           // bigint
	FavoritingCount         int             `json:"favoriting_count"`          // int
	FollowerCount           int64           `json:"follower_count"`            // bigint
	FollowingCount          int64           `json:"following_count"`           // bigint
	ForwardCount            int             `json:"forward_count"`             // int
	PublicCollectsCount     int             `json:"public_collects_count"`     // int
	MPlatformFollowersCount int64           `json:"mplatform_followers_count"` // bigint
	MaxFollowerCount        int64           `json:"max_follower_count"`        // bigint
	FollowStatus            int             `json:"follow_status"`             // int
	FollowerStatus          int             `json:"follower_status"`           // int
	FollowerRequestStatus   int             `json:"follower_request_status"`   // int
	CoverColour             string          `json:"cover_colour"`              // varchar(100)
	CoverURL                json.RawMessage `json:"cover_url"`                 // json
	WhiteCoverURL           json.RawMessage `json:"white_cover_url"`           // json
	ShareInfo               json.RawMessage `json:"share_info"`                // json
	CommerceInfo            json.RawMessage `json:"commerce_info"`             // json
	CommerceUserInfo        json.RawMessage `json:"commerce_user_info"`        // json
	CommerceUserLevel       int             `json:"commerce_user_level"`       // int
	CardEntries             json.RawMessage `json:"card_entries"`              // json
	Avatar168x168           json.RawMessage `json:"avatar_168x168"`            // json
	Avatar300x300           json.RawMessage `json:"avatar_300x300"`            // json
}

// CreateUserFactory 创建带数据库连接的用户模型实例。
func CreateUserFactory(sqlType string) *UserModel {
	return &UserModel{DB: model.UseDbConn(sqlType)}
}

// Register 创建账号记录，并初始化对应的用户资料骨架数据。
func (u *UserModel) Register(phone, password, userIp string) bool {
	var createTime = time.Now().Unix()
	sql1 := `INSERT INTO tb_accounts(phone, password, last_login_ip, create_time) SELECT ?, ?, ?, ? FROM DUAL WHERE NOT EXISTS(SELECT 1 FROM tb_accounts WHERE phone=?)`
	result := u.Exec(sql1, phone, password, userIp, createTime, phone)
	if result.Error != nil {
		variable.ZapLog.Error("Register SQL代码执行出错!", zap.Error(result.Error))
		return false
	}
	// 获取 UID
	var uid int64
	sql2 := `SELECT uid FROM tb_accounts WHERE phone = ?;`
	result = u.Raw(sql2, phone).Find(&uid)
	if result.Error != nil {
		variable.ZapLog.Error("Register-SELECT-uid SQL代码执行出错!", zap.Error(result.Error))
		return false
	}
	sql3 := `INSERT INTO tb_users(uid, card_entries) VALUES (?, ?);`
	result = u.Exec(sql3, uid, json.RawMessage{})
	// sql3 := `INSERT INTO tb_users(uid, nickname, unique_id, short_id, card_entries, avatar_168x168, avatar_300x300, avatar_small, avatar_large, signature, following_count, follower_count, total_favorited, aweme_count, favoriting_count) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	// nickname := fmt.Sprintf("User_%d", uid)
	// uniqueId := fmt.Sprintf("%d", uid)
	// defaultAvatar := `{"url_list":["https://p16-dy.byteimg.com/img/tos-cn-i-0813/c66072a3920942858548332150393433~c5_300x300.jpeg?from=4010531038"],"uri":"tos-cn-i-0813/c66072a3920942858548332150393433"}`

	// // 更新 tb_accounts 的 nickname
	// updateAccountSql := `UPDATE tb_accounts SET nickname=? WHERE uid=?;`
	// u.Exec(updateAccountSql, nickname, uid)
	// result = u.Exec(sql3, uid, nickname, uniqueId, 0, json.RawMessage("{}"), json.RawMessage(defaultAvatar), json.RawMessage(defaultAvatar), json.RawMessage(defaultAvatar), json.RawMessage(defaultAvatar), "这家伙很懒，什么都没留下", 0, 0, 0, 0, 0)
	if result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Login 校验账号密码，并返回匹配到的账号信息。
func (u *UserModel) Login(phone, password string) (account Account, ok bool) {
	sql := `
		SELECT ta.uid, ta.nickname, ta.phone, ta.password
		from tb_accounts as ta
		where phone=?
		limit 1;`
	result := u.Raw(sql, phone).Find(&account)
	if result.Error != nil {
		variable.ZapLog.Error("Login SQL代码执行出错!", zap.Error(result.Error))
		ok = false
		return
	}
	if account.Password == md5_encrypt.Base64Md5(password) {
		ok = true
	} else {
		ok = false
	}
	return
}

// Attention 执行关注或取消关注用户的操作。
func (u *UserModel) Attention(uid, followingId int64, action bool) bool {
	currentTime := time.Now().Unix()
	attentionSql := `INSERT INTO tb_relations (follower_id, following_id, create_time) VALUES (?, ?, ?);`
	unattentionSql := `DELETE FROM tb_relations WHERE follower_id=? and following_id=?;`
	var result *gorm.DB
	if action {
		result = u.Exec(attentionSql, uid, followingId, currentTime)
	} else {
		result = u.Exec(unattentionSql, uid, followingId)
	}
	if result.Error != nil {
		variable.ZapLog.Error("Attention SQL执行出错!", zap.Error(result.Error))
	}
	if result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// UpdateInfo 根据操作类型更新一个可编辑的用户资料字段。
func (u *UserModel) UpdateInfo(uid int64, operationType int, data string) bool {
	// 目前支持三种修改类型(nickname/unique_id/signature)
	var sql string
	switch operationType {
	case 1:
		sql = `UPDATE tb_users SET nickname=? WHERE uid=?;`
	case 2:
		sql = `UPDATE tb_users SET unique_id=? WHERE uid=?;`
	case 3:
		sql = `UPDATE tb_users SET signature=? WHERE uid=?;`
	case 4:
		sql = `UPDATE tb_users SET gender=? WHERE uid=?;`
	case 5:
		sql = `UPDATE tb_users SET birthday=? WHERE uid=?;`

	}
	result := u.Exec(sql, data, uid)
	if result.Error == nil {
		return true
	} else {
		variable.ZapLog.Error("Update UserInfo failed!")
		return false
	}
}

// AwemeStatus 查询当前用户的关注、点赞和收藏状态数据。
func (u *UserModel) AwemeStatus(uid int64) (awemeStatus AwemeStatusModel, success bool) {
	attentionSql := `SELECT following_id FROM tb_relations WHERE follower_id=?`
	diggSql := `SELECT aweme_id FROM tb_diggs WHERE uid=?`
	collectSql := `SELECT aweme_id FROM tb_collects WHERE uid=?`
	u.Raw(attentionSql, uid).Find(&awemeStatus.Attentions)
	u.Raw(diggSql, uid).Find(&awemeStatus.Likes)
	u.Raw(collectSql, uid).Find(&awemeStatus.Collects)
	return awemeStatus, true
}

// GetPanel 查询目标用户的个人主页面板数据。
func (u *UserModel) GetPanel(uid int64) (userinfo model.User, ok bool) {
	sql := `
		SELECT *
		from tb_users as tu
		where uid=?
		limit 1;`
	result := u.Raw(sql, uid).Find(&userinfo)
	if result.Error != nil {
		variable.ZapLog.Error("GetPanel SQL执行出错!", zap.Error(result.Error))
		ok = false
		return
	}
	likeCache := video.NewUserLikeStatusCache()
	if totalFavorited, cacheOK := likeCache.GetUserTotalFavorited(uid); cacheOK {
		userinfo.TotalFavorited = totalFavorited
	} else if totalFavorited, loadOK := likeCache.LoadUserTotalFavorited(uid); loadOK {
		userinfo.TotalFavorited = totalFavorited
	}
	ok = true
	return
}

// GetFriends 查询当前用户的互相关注好友列表。
func (u *UserModel) GetFriends(uid int64) (userinfo []model.User, ok bool) {
	sql := `
		SELECT *
		from tb_users as tu
		where uid IN (
		    SELECT tr.following_id
		    FROM tb_relations as tr
		    WHERE  tr.follower_id=?) AND
		    uid IN (
		    SELECT tr.follower_id
		    FROM tb_relations as tr
		    WHERE  tr.following_id=?
		    );`
	result := u.Raw(sql, uid, uid).Find(&userinfo)
	if result.Error != nil {
		variable.ZapLog.Error("GetFriends SQL执行出错!", zap.Error(result.Error))
		ok = false
		return
	}
	ok = true
	return
}

// GetFollow 查询当前用户的关注列表。
func (u *UserModel) GetFollow(uid int64) (userinfo []model.User, ok bool) {
	sql := `
		SELECT *
		from tb_users as tu
		where uid IN (
		    SELECT tr.following_id
		    FROM tb_relations as tr
		    WHERE  tr.follower_id=?);`
	result := u.Raw(sql, uid).Find(&userinfo)
	if result.Error != nil {
		variable.ZapLog.Error("GetFollow SQL执行出错!", zap.Error(result.Error))
		ok = false
		return
	}
	ok = true
	return
}

// GetFans 查询当前用户的粉丝列表。
func (u *UserModel) GetFans(uid int64) (userinfo []model.User, ok bool) {
	sql := `
		SELECT *
		from tb_users as tu
		where uid IN (
		    SELECT tr.follower_id
		    FROM tb_relations as tr
		    WHERE  tr.following_id=?);`
	result := u.Raw(sql, uid).Find(&userinfo)
	if result.Error != nil {
		variable.ZapLog.Error("GetFans SQL执行出错!", zap.Error(result.Error))
		ok = false
		return
	}
	ok = true
	return
}

// OauthCheckTokenIsOk 检查指定 Token 是否仍属于当前用户的有效 Token 集合。
func (u *UserModel) OauthCheckTokenIsOk(uid int64, token string) bool {
	sql := `SELECT token 
			FROM tb_auth_access_tokens 
			WHERE uid=? AND revoked=0 AND expires_at>UNIX_TIMESTAMP(NOW()) ORDER BY expires_at DESC, updated_at DESC LIMIT ?;`
	maxOnlineUsers := variable.ConfigYml.GetInt("Token.JwtTokenOnlineUsers")
	rows, err := u.Raw(sql, uid, maxOnlineUsers).Rows()
	defer func() {
		// 释放记录集
		_ = rows.Close()
	}()
	if err == nil && rows != nil {
		for rows.Next() {
			var tempToken string
			err := rows.Scan(&tempToken)
			fmt.Println(err)
			if err == nil {
				if tempToken == token {
					return true
				}
			}
		}
	}
	return false
}

// OauthLoginToken 记录当前用户新签发的登录 Token。
func (u *UserModel) OauthLoginToken(uid int64, token string, expiresAt int64, clientIp string) bool {
	sql := `INSERT INTO tb_auth_access_tokens(uid, action_name, token, created_at, expires_at, client_ip)
			SELECT ?, 'login', ?, ?, ?, ? FROM DUAL WHERE NOT EXISTS(SELECT 1 FROM tb_auth_access_tokens a WHERE a.uid=? AND a.action_name='login' AND a.token=?)
			`
	createdAt := time.Now().Unix()
	if u.Exec(sql, uid, token, createdAt, expiresAt, clientIp, uid, token).Error == nil {
		// 异步缓存用户有效的token到redis
		if variable.ConfigYml.GetInt("Token.IsCacheToRedis") == 1 {
			go u.ValidTokenCacheToRedis(uid)
		}
		return true
	}
	return false
}
