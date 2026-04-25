package upload

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/http/controller/web"
	"douyin-backend/app/http/validator/core/data_transfer"
	"douyin-backend/app/utils/response"
	"github.com/gin-gonic/gin"
)

// VideoComplete 定义视频上传完成请求参数。
type VideoComplete struct {
	UploadID
}

// CheckParams 校验视频上传完成参数并分发到控制器。
func (v VideoComplete) CheckParams(ctx *gin.Context) {
	if err := ctx.ShouldBind(&v); err != nil {
		response.ValidatorError(ctx, err)
		return
	}

	extraAddBindDataContext := data_transfer.DataAddContext(v, consts.ValidatorPrefix, ctx)
	if extraAddBindDataContext == nil {
		response.ErrorSystem(ctx, "video complete validator failed", "")
		return
	}
	(&web.UploadController{}).VideoComplete(extraAddBindDataContext)
}
