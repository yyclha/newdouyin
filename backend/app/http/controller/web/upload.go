package web

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/global/variable"
	"douyin-backend/app/service/upload_file"
	"douyin-backend/app/utils/response"
	"github.com/gin-gonic/gin"
)

type UploadController struct {
}

func (u *UploadController) Avatar(ctx *gin.Context) {
	savePath := variable.ConfigYml.GetString("FileUploadSetting.UploadRootPath") + variable.ConfigYml.GetString("FileUploadSetting.AvatarSmallUploadFileSavePath")
	if r, finnalSavePath := upload_file.UploadAvatar(ctx, savePath); r == true {
		response.Success(ctx, consts.CurdStatusOkMsg, finnalSavePath)
	} else {
		response.Fail(ctx, consts.FilesUploadFailCode, consts.FilesUploadFailMsg, "")
	}

}

func (u *UploadController) Cover(ctx *gin.Context) {
	savePath := variable.ConfigYml.GetString("FileUploadSetting.UploadRootPath") + variable.ConfigYml.GetString("FileUploadSetting.CoverUploadFileSavePath")
	if r, finnalSavePath := upload_file.UploadCover(ctx, savePath); r == true {
		response.Success(ctx, consts.CurdStatusOkMsg, finnalSavePath)
	} else {
		response.Fail(ctx, consts.FilesUploadFailCode, consts.FilesUploadFailMsg, "")
	}

}

func (u *UploadController) Video(ctx *gin.Context) {
	savePath := variable.ConfigYml.GetString("FileUploadSetting.UploadRootPath") + variable.ConfigYml.GetString("FileUploadSetting.VideoUploadFileSavePath")
	if r, finnalSavePath, message := upload_file.UploadVideo(ctx, savePath); r == true {
		response.Success(ctx, consts.CurdStatusOkMsg, finnalSavePath)
	} else {
		if message == "" {
			message = consts.FilesUploadFailMsg
		}
		response.Fail(ctx, consts.FilesUploadFailCode, message, "")
	}
}

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
