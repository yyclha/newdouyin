package upload

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/global/variable"
	"douyin-backend/app/http/controller/web"
	"douyin-backend/app/http/validator/core/data_transfer"
	"douyin-backend/app/utils/response"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

// VideoChunk 定义视频分片上传请求参数。
type VideoChunk struct {
	UploadID
	ChunkIndex
	TotalChunks
	ChunkHash
}

// CheckParams 校验视频分片上传参数并分发到控制器。
func (v VideoChunk) CheckParams(ctx *gin.Context) {
	tmpFile, err := ctx.FormFile(variable.ConfigYml.GetString("FileUploadSetting.UploadFileField"))
	if err != nil {
		response.Fail(ctx, consts.FilesUploadFailCode, consts.FilesUploadFailMsg, err.Error())
		return
	}

	sizeLimit := variable.ConfigYml.GetInt64("FileUploadSetting.Size")
	if tmpFile.Size == 0 {
		response.Fail(ctx, consts.FilesUploadMoreThanMaxSizeCode, consts.FilesUploadIsEmpty, "")
		return
	}
	if tmpFile.Size > sizeLimit<<20 {
		response.Fail(ctx, consts.FilesUploadMoreThanMaxSizeCode, consts.FilesUploadMoreThanMaxSizeMsg+strconv.FormatInt(sizeLimit, 10)+"M", "")
		return
	}

	if err = ctx.ShouldBind(&v); err != nil {
		response.ValidatorError(ctx, err)
		return
	}
	if int(*v.TotalChunks.TotalChunks) <= 0 || int(*v.ChunkIndex.ChunkIndex) < 0 {
		response.Fail(ctx, consts.ValidatorParamsCheckFailCode, "chunk_index and total_chunks are invalid", "")
		return
	}
	if strings.TrimSpace(*v.ChunkHash.ChunkHash) == "" {
		response.Fail(ctx, consts.ValidatorParamsCheckFailCode, "chunk_hash is required", "")
		return
	}

	extraAddBindDataContext := data_transfer.DataAddContext(v, consts.ValidatorPrefix, ctx)
	if extraAddBindDataContext == nil {
		response.ErrorSystem(ctx, "video chunk validator failed", "")
		return
	}
	(&web.UploadController{}).VideoChunk(extraAddBindDataContext)
}
