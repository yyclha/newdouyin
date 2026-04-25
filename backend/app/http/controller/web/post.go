package web

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/model/post"
	"douyin-backend/app/utils/auth"
	"douyin-backend/app/utils/response"
	"github.com/gin-gonic/gin"
)

// PostController 处理图文动态相关的 HTTP 接口。
type PostController struct {
}

// GetPostRecommended 获取推荐图文动态列表。
func (u *PostController) GetPostRecommended(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	var pageNo = ctx.GetFloat64(consts.ValidatorPrefix + "pageNo")
	var pageSize = ctx.GetFloat64(consts.ValidatorPrefix + "pageSize")
	list, total, ok := post.CreatePostFactory("").GetPostRecommended(uid, int64(pageNo), int64(pageSize))
	if !ok {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "获取推荐动态失败")
	} else {
		response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
			"pageNo": pageNo,
			"total":  total,
			"list":   list,
		})
	}
}
