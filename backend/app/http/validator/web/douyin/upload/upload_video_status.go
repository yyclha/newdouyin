package upload

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/http/controller/web"
	"douyin-backend/app/http/validator/core/data_transfer"
	"douyin-backend/app/utils/response"

	"github.com/gin-gonic/gin"
)

// VideoStatus 定义视频后台处理任务状态查询参数。
type VideoStatus struct {
	TaskID
}

// CheckParams 校验视频后台处理任务状态查询参数并分发到控制器。
func (v VideoStatus) CheckParams(ctx *gin.Context) {
	if err := ctx.ShouldBind(&v); err != nil {
		response.ValidatorError(ctx, err)
		return
	}

	extraAddBindDataContext := data_transfer.DataAddContext(v, consts.ValidatorPrefix, ctx)
	if extraAddBindDataContext == nil {
		response.ErrorSystem(ctx, "video status validator failed", "")
		return
	}
	(&web.UploadController{}).VideoStatus(extraAddBindDataContext)
}
