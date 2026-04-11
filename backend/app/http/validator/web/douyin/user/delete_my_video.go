package user

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/http/controller/web"
	"douyin-backend/app/http/validator/core/data_transfer"
	"douyin-backend/app/utils/response"
	"github.com/gin-gonic/gin"
)

type DeleteMyVideo struct {
	AwemeID
}

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
