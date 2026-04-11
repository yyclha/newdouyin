package web

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/global/variable"
	"douyin-backend/app/http/middleware/my_jwt"
	"douyin-backend/app/model/video"
	userTokenService "douyin-backend/app/service/users/token"
	"douyin-backend/app/utils/auth"
	"douyin-backend/app/utils/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/rand"
)

type VideoController struct {
}

func (v *VideoController) VideoDigg(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	var awemeId = ctx.GetString(consts.ValidatorPrefix + "aweme_id")
	var action = ctx.GetBool(consts.ValidatorPrefix + "action")
	var awemeIDInt64, _ = strconv.ParseInt(awemeId, 10, 64)
	actionStatus := video.CreateDiggFactory("").VideoDigg(uid, awemeIDInt64, action)
	if actionStatus {
		if action {
			response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
				"data": actionStatus,
				"msg":  "点赞成功",
			})
		} else {
			response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
				"data": actionStatus,
				"msg":  "取消点赞成功",
			})
		}
	} else {
		if action {
			response.Fail(ctx, consts.CurdInsertFailCode, consts.CurdInsertFailMsg, gin.H{
				"data": actionStatus,
				"msg":  "点赞失败",
			})
		} else {
			response.Fail(ctx, consts.CurdInsertFailCode, consts.CurdInsertFailMsg, gin.H{
				"data": actionStatus,
				"msg":  "取消点赞失败",
			})
		}
	}
}

func (v *VideoController) VideoComment(ctx *gin.Context) {
	var ipLocation = ctx.GetString(consts.ValidatorPrefix + "ip_location")
	var awemeId = ctx.GetString(consts.ValidatorPrefix + "aweme_id")
	var content = ctx.GetString(consts.ValidatorPrefix + "content")
	var uid = auth.GetUidFromToken(ctx)
	var shortId = ctx.GetString(consts.ValidatorPrefix + "short_id")
	var uniqueId = ctx.GetString(consts.ValidatorPrefix + "unique_id")
	var signature = ctx.GetString(consts.ValidatorPrefix + "signature")
	var nickname = ctx.GetString(consts.ValidatorPrefix + "nickname")
	var avatar = ctx.GetString(consts.ValidatorPrefix + "avatar")
	var awemeIDInt64, _ = strconv.ParseInt(awemeId, 10, 64)

	// 如果前端传递了 short_id，则使用它；否则尝试从 token 对应的用户信息中获取（如果需要更严谨，可以在这里查询用户信息）
	// 目前保持原有逻辑，直接透传参数

	commentID, commentDone := video.CreateCommentFactory("").VideoComment(uid, awemeIDInt64, ipLocation, content, shortId, uniqueId, signature, nickname, avatar)
	if commentDone {
		response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
			"comment_id": strconv.FormatInt(commentID, 10),
			"msg":        "评论成功",
		})
	} else {
		response.Fail(ctx, consts.CurdInsertFailCode, consts.CurdInsertFailMsg, gin.H{
			"data": false,
			"msg":  "评论失败",
		})
	}

}

func (v *VideoController) CommentDigg(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	var commentIDStr = ctx.GetString(consts.ValidatorPrefix + "comment_id")
	var action = ctx.GetBool(consts.ValidatorPrefix + "action")
	var commentID, _ = strconv.ParseInt(commentIDStr, 10, 64)
	actionStatus := video.CreateCommentDiggFactory("").CommentDigg(uid, commentID, action)
	if actionStatus {
		if action {
			response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
				"data": actionStatus,
				"msg":  "评论点赞成功",
			})
		} else {
			response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
				"data": actionStatus,
				"msg":  "取消评论点赞成功",
			})
		}
	} else {
		if action {
			response.Fail(ctx, consts.CurdInsertFailCode, consts.CurdInsertFailMsg, gin.H{
				"data": actionStatus,
				"msg":  "评论点赞失败",
			})
		} else {
			response.Fail(ctx, consts.CurdInsertFailCode, consts.CurdInsertFailMsg, gin.H{
				"data": actionStatus,
				"msg":  "取消评论点赞失败",
			})
		}
	}
}

func (v *VideoController) DeleteComment(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	var commentIDStr = ctx.GetString(consts.ValidatorPrefix + "comment_id")
	var commentID, _ = strconv.ParseInt(commentIDStr, 10, 64)

	deleteDone := video.CreateCommentFactory("").DeleteComment(uid, commentID)
	if deleteDone {
		response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
			"data": true,
			"msg":  "删除评论成功",
		})
	} else {
		response.Fail(ctx, consts.CurdDeleteFailCode, consts.CurdDeleteFailMsg, gin.H{
			"data": false,
			"msg":  "删除评论失败",
		})
	}
}

func (v *VideoController) VideoCollect(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	var awemeId = ctx.GetString(consts.ValidatorPrefix + "aweme_id")
	var action = ctx.GetBool(consts.ValidatorPrefix + "action")
	var awemeIDInt64, _ = strconv.ParseInt(awemeId, 10, 64)
	actionStatus := video.CreateCollectFactory("").VideoCollect(uid, awemeIDInt64, action)
	if actionStatus {
		if action {
			response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
				"data": actionStatus,
				"msg":  "收藏成功",
			})
		} else {
			response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
				"data": actionStatus,
				"msg":  "取消收藏成功",
			})
		}
	} else {
		if action {
			response.Fail(ctx, consts.CurdInsertFailCode, consts.CurdInsertFailMsg, gin.H{
				"data": actionStatus,
				"msg":  "收藏失败",
			})
		} else {
			response.Fail(ctx, consts.CurdInsertFailCode, consts.CurdInsertFailMsg, gin.H{
				"data": actionStatus,
				"msg":  "取消收藏失败",
			})
		}
	}
}

func (v *VideoController) VideoShare(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	var awemeId = ctx.GetString(consts.ValidatorPrefix + "aweme_id")
	var message = ctx.GetString(consts.ValidatorPrefix + "message")
	var shareUidList = ctx.GetString(consts.ValidatorPrefix + "share_uid_list")
	var awemeIDInt64, _ = strconv.ParseInt(awemeId, 10, 64)
	shareDone := video.CreateShareFactory("").VideoShare(uid, awemeIDInt64, message, shareUidList)
	if shareDone {
		response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
			"data": shareDone,
			"msg":  "分享成功",
		})
	} else {
		response.Fail(ctx, consts.CurdInsertFailCode, consts.CurdInsertFailMsg, gin.H{
			"data": shareDone,
			"msg":  "分享失败",
		})
	}
}

func (v *VideoController) GetComments(ctx *gin.Context) {
	awemeIdStr := ctx.GetString(consts.ValidatorPrefix + "aweme_id")
	awemeId, err := strconv.ParseInt(awemeIdStr, 10, 64)
	if err != nil {
		response.Fail(ctx, consts.CurdSelectFailCode, "Invalid aweme_id", "")
		return
	}
	currentUID := tryGetUIDFromHeaderToken(ctx)
	pageNo := getInt64FromContext(ctx, consts.ValidatorPrefix+"pageNo", 0)
	pageSize := getInt64FromContext(ctx, consts.ValidatorPrefix+"pageSize", 20)
	comments, total, hasMore, ok := video.CreateCommentFactory("").GetComments(awemeId, currentUID, pageNo, pageSize)
	if ok {
		response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
			"list":     comments,
			"total":    total,
			"pageNo":   pageNo,
			"pageSize": pageSize,
			"hasMore":  hasMore,
		})
	} else {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

func getInt64FromContext(ctx *gin.Context, key string, defaultValue int64) int64 {
	value, exists := ctx.Get(key)
	if !exists || value == nil {
		return defaultValue
	}

	switch typed := value.(type) {
	case float64:
		if int64(typed) <= 0 && defaultValue > 0 {
			return defaultValue
		}
		return int64(typed)
	case float32:
		if int64(typed) <= 0 && defaultValue > 0 {
			return defaultValue
		}
		return int64(typed)
	case int:
		if int64(typed) <= 0 && defaultValue > 0 {
			return defaultValue
		}
		return int64(typed)
	case int64:
		if typed <= 0 && defaultValue > 0 {
			return defaultValue
		}
		return typed
	case string:
		parsed, err := strconv.ParseInt(typed, 10, 64)
		if err != nil {
			return defaultValue
		}
		if parsed <= 0 && defaultValue > 0 {
			return defaultValue
		}
		return parsed
	default:
		return defaultValue
	}
}

func tryGetUIDFromHeaderToken(ctx *gin.Context) int64 {
	token := ctx.GetHeader("Token")
	if token == "" {
		return 0
	}

	customClaims, err := userTokenService.CreateUserFactory().ParseToken(token)
	if err != nil {
		return 0
	}

	key := variable.ConfigYml.GetString("Token.BindContextKeyName")
	ctx.Set(key, my_jwt.CustomClaims(customClaims))
	return customClaims.UID
}

func (v *VideoController) GetHistoryOther(context *gin.Context) {
	// TODO 具体业务逻辑实现
}

func (v *VideoController) GetLongVideoRecommended(ctx *gin.Context) {
	// TODO 具体业务逻辑实现
	var uid = auth.GetUidFromToken(ctx)
	var PageNo = ctx.GetFloat64(consts.ValidatorPrefix + "pageNo")
	var PageSize = ctx.GetFloat64(consts.ValidatorPrefix + "pageSize")
	list, total, ok := video.CreateVideoFactory("").GetLongVideoRecommended(uid, int64(PageNo), int64(PageSize))
	if ok {
		response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
			"total": total,
			"list":  list,
		})
	} else {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

func (v *VideoController) GetVideoRecommended(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	var Start = ctx.GetFloat64(consts.ValidatorPrefix + "start")
	var PageSize = ctx.GetFloat64(consts.ValidatorPrefix + "pageSize")
	list, total, ok := video.CreateVideoFactory("").GetVideoRecommended(uid, int64(Start), int64(PageSize))
	if ok && len(list) > 0 {
		rand.Seed(uint64(time.Now().UnixNano()))
		// 打乱切片
		rand.Shuffle(len(list), func(i, j int) {
			list[i], list[j] = list[j], list[i]
		})
		response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
			"total": total,
			"list":  list,
		})
	} else {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

func (v *VideoController) GetHistory(context *gin.Context) {
	// TODO 具体业务逻辑实现

	//var id = context.GetFloat64(consts.ValidatorPrefix + "id")
	//video := sv_home.CreateShortVideoFactory("").GetVideoById(int(id))
	//if video.Id != 0 {
	//	response.Success(context, consts.CurdStatusOkMsg, video)
	//} else {
	//	response.Fail(context, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	//}
}
