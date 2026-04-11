package upload

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/global/variable"
	"douyin-backend/app/http/controller/web"
	"douyin-backend/app/http/validator/core/data_transfer"
	"douyin-backend/app/utils/response"
	"github.com/gin-gonic/gin"
	"strconv"
)

type VideoInit struct {
	UploadID
	FileName
	FileSize
	ChunkSize
	TotalChunks
	ContentType
	Description
	Tags
	PrivateStatus
}

func (v VideoInit) CheckParams(ctx *gin.Context) {
	if err := ctx.ShouldBind(&v); err != nil {
		response.ValidatorError(ctx, err)
		return
	}

	sizeLimit := variable.ConfigYml.GetInt64("FileUploadSetting.Size")
	fileSize := int64(*v.FileSize.FileSize)
	chunkSize := int64(*v.ChunkSize.ChunkSize)
	totalChunks := int(*v.TotalChunks.TotalChunks)
	if fileSize <= 0 || chunkSize <= 0 || totalChunks <= 0 {
		response.Fail(ctx, consts.ValidatorParamsCheckFailCode, "file_size, chunk_size and total_chunks must be greater than 0", "")
		return
	}
	if fileSize > sizeLimit<<20 {
		response.Fail(ctx, consts.FilesUploadMoreThanMaxSizeCode, consts.FilesUploadMoreThanMaxSizeMsg+strconv.FormatInt(sizeLimit, 10)+"M", "")
		return
	}
	if chunkSize > sizeLimit<<20 {
		response.Fail(ctx, consts.FilesUploadMoreThanMaxSizeCode, "chunk size exceeds limit", "")
		return
	}

	extraAddBindDataContext := data_transfer.DataAddContext(v, consts.ValidatorPrefix, ctx)
	if extraAddBindDataContext == nil {
		response.ErrorSystem(ctx, "video init validator failed", "")
		return
	}
	(&web.UploadController{}).VideoInit(extraAddBindDataContext)
}
