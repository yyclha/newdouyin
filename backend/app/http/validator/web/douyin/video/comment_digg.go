package video

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/http/controller/web"
	"douyin-backend/app/http/validator/core/data_transfer"
	"douyin-backend/app/utils/response"
	"github.com/gin-gonic/gin"
)

// CommentDigg 定义评论点赞请求参数。
type CommentDigg struct {
	CommentID
	Action
}

// CheckParams 校验评论点赞参数并分发到控制器。
func (v CommentDigg) CheckParams(context *gin.Context) {
	if err := context.ShouldBind(&v); err != nil {
		response.ValidatorError(context, err)
		return
	}
	extraAddBindDataContext := data_transfer.DataAddContext(v, consts.ValidatorPrefix, context)
	if extraAddBindDataContext == nil {
		response.ErrorSystem(context, "comment_digg 表单验证器json化失败", "")
	} else {
		(&web.VideoController{}).CommentDigg(extraAddBindDataContext)
	}
}
