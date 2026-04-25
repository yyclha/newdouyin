package websocket

import (
	serviceWs "douyin-backend/app/service/websocket"
	"github.com/gin-gonic/gin"
)

/*
WebSocket 相关实现可参考 gorilla/websocket 官方示例：
https://github.com/gorilla/websocket/tree/master/examples
*/

// Ws 处理 WebSocket 握手和消息分发入口。
type Ws struct {
}

// OnOpen 处理 WebSocket 握手并完成协议升级。
func (w *Ws) OnOpen(context *gin.Context) (*serviceWs.Ws, bool) {
	return (&serviceWs.Ws{}).OnOpen(context)
}

// OnMessage 处理 WebSocket 业务消息。
func (w *Ws) OnMessage(serviceWs *serviceWs.Ws, context *gin.Context) {
	serviceWs.OnMessage(context)
}
