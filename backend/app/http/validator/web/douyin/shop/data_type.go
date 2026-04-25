package shop

// Uid 定义用户 ID 参数。
type Uid struct {
	Uid *float64 `form:"uid" json:"uid" binding:"required,min=0"`
}

// PageNo 定义分页页码参数。
type PageNo struct {
	PageNo *float64 `form:"pageNo" json:"pageNo" binding:"required,min=0"` // 注意：gin数字的存储形式以 float64 接受
}

// PageSize 定义分页条数参数。
type PageSize struct {
	PageSize *float64 `form:"pageSize" json:"pageSize" binding:"required,min=0"` // 注意：gin数字的存储形式以 float64 接受
}
