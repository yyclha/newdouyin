package video

// IpLocation 定义 IP 归属地参数。
type IpLocation struct {
	IpLocation *string `form:"ip_location" json:"ip_location"`
}

// ShareUidList 定义分享用户 ID 列表参数。
type ShareUidList struct {
	ShareUidList *string `form:"share_uid_list" json:"share_uid_list"`
}

// Message 定义消息文本参数。
type Message struct {
	Message *string `form:"message" json:"message"`
}

// Uid 定义用户 ID 参数。
type Uid struct {
	Uid *string `form:"uid" json:"uid" binding:"required,numeric"`
}

// Start 定义分页起始偏移参数。
type Start struct {
	Start *float64 `form:"start" json:"start" binding:"required,min=0"`
}

// PageNo 定义分页页码参数。
type PageNo struct {
	PageNo *float64 `form:"pageNo" json:"pageNo" binding:"required,min=0"` // 注意：gin数字的存储形式以 float64 接受
}

// PageSize 定义分页条数参数。
type PageSize struct {
	PageSize *float64 `form:"pageSize" json:"pageSize" binding:"required,min=0"` // 注意：gin数字的存储形式以 float64 接受
}

// AwemeID 定义视频作品 ID 参数。
type AwemeID struct {
	AwemeID *string `form:"aweme_id" json:"aweme_id" binding:"required,numeric"`
}

// CommentID 定义评论 ID 参数。
type CommentID struct {
	CommentID *string `form:"comment_id" json:"comment_id" binding:"required,numeric"`
}

// Action 定义点赞或取消点赞参数。
type Action struct {
	Action *bool `form:"action" json:"action" binding:"required"`
}

// Content 定义评论内容参数。
type Content struct {
	Content *string `form:"content" json:"content" binding:"required"`
}

// ShortID 定义用户短号参数。
type ShortID struct {
	ShortID *string `form:"short_id" json:"short_id"`
}

// UniqueID 定义用户唯一标识参数。
type UniqueID struct {
	UniqueID *string `form:"unique_id" json:"unique_id"`
}

// Signature 定义个性签名参数。
type Signature struct {
	Signature *string `form:"signature" json:"signature"`
}

// Nickname 定义用户昵称参数。
type Nickname struct {
	Nickname *string `form:"nickname" json:"nickname"`
}

// Avatar 定义头像地址参数。
type Avatar struct {
	Avatar *string `form:"avatar" json:"avatar"`
}
