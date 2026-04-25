package snowflake_interf

// InterfaceSnowFlake 定义业务数据结构。
type InterfaceSnowFlake interface {
	GetId() int64
}
