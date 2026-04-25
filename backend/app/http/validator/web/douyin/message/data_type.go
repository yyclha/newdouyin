package message

// TxUid 定义消息发送方用户 ID 参数。
type TxUid struct {
	TxUid *string `form:"tx_uid" json:"tx_uid" binding:"required,numeric"`
}

// RxUid 定义消息接收方用户 ID 参数。
type RxUid struct {
	RxUid *string `form:"rx_uid" json:"rx_uid" binding:"required,numeric"`
}

// MsgType 定义消息类型参数。
type MsgType struct {
	MsgType *float64 `form:"msg_type" json:"msg_type" binding:"required"`
}

// MsgData 定义消息内容参数。
type MsgData struct {
	MsgData *string `form:"msg_data" json:"msg_data" binding:"required"`
}

// ReadState 定义消息已读状态参数。
type ReadState struct {
	ReadState *float64 `form:"read_state" json:"read_state" binding:"required"`
}

// CreateTime 定义消息创建时间参数。
type CreateTime struct {
	CreateTime *float64 `form:"create_time" json:"create_time" binding:"required"`
}
