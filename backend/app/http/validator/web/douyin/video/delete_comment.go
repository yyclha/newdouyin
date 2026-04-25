package video

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/http/controller/web"
	"douyin-backend/app/http/validator/core/data_transfer"
	"douyin-backend/app/utils/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DeleteComment 定义删除评论请求参数。
type DeleteComment struct {
	CommentID
}

// CheckParams 校验删除评论参数并分发到控制器。
func (v DeleteComment) CheckParams(context *gin.Context) {
	commentID := context.PostForm("comment_id")
	if commentID == "" {
		commentID = context.Query("comment_id")
	}
	if commentID == "" {
		var body struct {
			CommentID interface{} `json:"comment_id"`
		}
		if err := context.ShouldBindJSON(&body); err == nil && body.CommentID != nil {
			switch value := body.CommentID.(type) {
			case string:
				commentID = value
			case float64:
				commentID = strconv.FormatInt(int64(value), 10)
			}
		}
	}

	if commentID == "" {
		response.ErrorParam(context, gin.H{"comment_id": "comment_id 为必填项"})
		return
	}
	if _, err := strconv.ParseInt(commentID, 10, 64); err != nil {
		response.ValidatorError(context, err)
		return
	}

	v.CommentID.CommentID = &commentID
	extraAddBindDataContext := data_transfer.DataAddContext(v, consts.ValidatorPrefix, context)
	if extraAddBindDataContext == nil {
		response.ErrorSystem(context, "delete_comment data add context error", "")
	} else {
		(&web.VideoController{}).DeleteComment(extraAddBindDataContext)
	}
}
