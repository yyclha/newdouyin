package websocket

// Request 定义业务数据结构。
type Request struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

// Response 定义业务数据结构。
type Response struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}
