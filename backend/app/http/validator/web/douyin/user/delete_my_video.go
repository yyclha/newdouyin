package user

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/http/controller/web"
	"douyin-backend/app/http/validator/core/data_transfer"
	"douyin-backend/app/utils/response"
	"github.com/gin-gonic/gin"
)

// DeleteMyVideo 定义删除个人视频请求参数。
type DeleteMyVideo struct {
	AwemeID
}

// CheckParams 校验删除个人视频参数并分发到控制器。
func (d DeleteMyVideo) CheckParams(context *gin.Context) {
	if err := context.ShouldBind(&d); err != nil {
		response.ValidatorError(context, err)
		return
	}
	extraAddBindDataContext := data_transfer.DataAddContext(d, consts.ValidatorPrefix, context)
	if extraAddBindDataContext == nil {
		response.ErrorSystem(context, "delete_my_video 表单验证器json化失败", "")
	} else {
		(&web.UserController{}).DeleteMyVideo(extraAddBindDataContext)
	}
}
