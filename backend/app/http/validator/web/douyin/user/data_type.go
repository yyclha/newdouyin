package user

// Nickname 定义用户昵称参数。
type Nickname struct {
	Nickname *string `form:"nickname" json:"nickname" binding:"required"`
}

// Phone 定义手机号参数。
type Phone struct {
	Phone *string `form:"phone" json:"phone" binding:"required,len=11"`
}

// Password 定义用户密码参数。
type Password struct {
	Password *string `form:"password" json:"password" binding:"required,min=6,max=20"`
}

// Uid 定义用户 ID 参数。
type Uid struct {
	Uid *float64 `form:"uid" json:"uid" binding:"required,gt=0"`
}

// PageNo 定义分页页码参数。
type PageNo struct {
	PageNo *float64 `form:"pageNo" json:"pageNo" binding:"required,min=0"`
}

// PageSize 定义分页条数参数。
type PageSize struct {
	PageSize *float64 `form:"pageSize" json:"pageSize" binding:"required,min=0"`
}

// AwemeID 定义视频作品 ID 参数。
type AwemeID struct {
	AwemeID *string `form:"aweme_id" json:"aweme_id" binding:"required,numeric"`
}

// Action 定义布尔操作参数。
type Action struct {
	Action *bool `form:"action" json:"action" binding:"required"`
}

// FollowingId 定义被关注用户 ID 参数。
type FollowingId struct {
	FollowingId *string `form:" following_id" json:"following_id" binding:"required"`
}

// OperationType 定义操作类型参数。
type OperationType struct {
	OperationType *float64 `form:"operation_type" json:"operation_type" binding:"required"`
}

// Data 定义通用数据载荷参数。
type Data struct {
	Data *string `form:" data" json:"data" binding:"required"`
}
