package web

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/global/variable"
	"douyin-backend/app/service/upload_file"
	"douyin-backend/app/utils/response"
	"github.com/gin-gonic/gin"
)

// UploadController 处理头像、封面和视频上传相关接口。
type UploadController struct {
}

// Avatar 上传用户头像文件。
func (u *UploadController) Avatar(ctx *gin.Context) {
	savePath := variable.ConfigYml.GetString("FileUploadSetting.UploadRootPath") + variable.ConfigYml.GetString("FileUploadSetting.AvatarSmallUploadFileSavePath")
	if r, finnalSavePath := upload_file.UploadAvatar(ctx, savePath); r == true {
		response.Success(ctx, consts.CurdStatusOkMsg, finnalSavePath)
	} else {
		response.Fail(ctx, consts.FilesUploadFailCode, consts.FilesUploadFailMsg, "")
	}

}

// Cover 上传视频封面文件。
func (u *UploadController) Cover(ctx *gin.Context) {
	savePath := variable.ConfigYml.GetString("FileUploadSetting.UploadRootPath") + variable.ConfigYml.GetString("FileUploadSetting.CoverUploadFileSavePath")
	if r, finnalSavePath := upload_file.UploadCover(ctx, savePath); r == true {
		response.Success(ctx, consts.CurdStatusOkMsg, finnalSavePath)
	} else {
		response.Fail(ctx, consts.FilesUploadFailCode, consts.FilesUploadFailMsg, "")
	}

}

// VideoInit 初始化视频分片上传任务。
func (u *UploadController) VideoInit(ctx *gin.Context) {
	if r, finalSavePath, message := upload_file.InitVideoChunkUpload(ctx); r == true {
		response.Success(ctx, consts.CurdStatusOkMsg, finalSavePath)
	} else {
		if message == "" {
			message = consts.FilesUploadFailMsg
		}
		response.Fail(ctx, consts.FilesUploadFailCode, message, finalSavePath)
	}
}

// VideoChunk 保存单个视频分片。
func (u *UploadController) VideoChunk(ctx *gin.Context) {
	if r, finalSavePath, message := upload_file.SaveVideoChunk(ctx); r == true {
		response.Success(ctx, consts.CurdStatusOkMsg, finalSavePath)
	} else {
		if message == "" {
			message = consts.FilesUploadFailMsg
		}
		response.Fail(ctx, consts.FilesUploadFailCode, message, finalSavePath)
	}
}

// VideoComplete 合并分片并完成视频上传。
func (u *UploadController) VideoComplete(ctx *gin.Context) {
	savePath := variable.ConfigYml.GetString("FileUploadSetting.UploadRootPath") + variable.ConfigYml.GetString("FileUploadSetting.VideoUploadFileSavePath")
	if r, finalSavePath, message := upload_file.CompleteVideoChunkUpload(ctx, savePath); r == true {
		response.Success(ctx, consts.CurdStatusOkMsg, finalSavePath)
	} else {
		if message == "" {
			message = consts.FilesUploadFailMsg
		}
		response.Fail(ctx, consts.FilesUploadFailCode, message, finalSavePath)
	}
}

// VideoStatus 查询后台视频处理任务状态。
func (u *UploadController) VideoStatus(ctx *gin.Context) {
	if r, payload, message := upload_file.GetVideoUploadTaskStatus(ctx); r == true {
		response.Success(ctx, consts.CurdStatusOkMsg, payload)
	} else {
		if message == "" {
			message = consts.CurdSelectFailMsg
		}
		response.Fail(ctx, consts.CurdSelectFailCode, message, payload)
	}
}
