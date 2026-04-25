package web

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/model/message"
	"douyin-backend/app/utils/auth"
	"douyin-backend/app/utils/response"
	"github.com/gin-gonic/gin"
)

// MessageController 处理消息相关的 HTTP 接口。
type MessageController struct {
}

// GetAllMsg 获取当前用户的全部消息会话。
func (m *MessageController) GetAllMsg(ctx *gin.Context) {
	uid := auth.GetUidFromToken(ctx)
	allMsg, ok := message.CreateMsgFactory("").GetAllMsg(uid)
	if !ok {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
	if len(allMsg) > 0 {
		response.Success(ctx, consts.CurdStatusOkMsg, allMsg)
	} else {
		response.Success(ctx, consts.CurdStatusOkMsg, []interface{}{})
	}
}
