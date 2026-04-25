package web

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/global/variable"
	"douyin-backend/app/model/user"
	"douyin-backend/app/model/video"
	userstoken "douyin-backend/app/service/users/token"
	"douyin-backend/app/utils/auth"
	"douyin-backend/app/utils/md5_encrypt"
	"douyin-backend/app/utils/response"
	"github.com/gin-gonic/gin"
	"strconv"
)

// UserController 处理用户相关的 HTTP 接口。
type UserController struct {
}

// Register 处理用户注册请求。
func (u *UserController) Register(ctx *gin.Context) {
	var phone = ctx.GetString(consts.ValidatorPrefix + "phone")
	var password = ctx.GetString(consts.ValidatorPrefix + "password")
	var userIp = ctx.ClientIP()
	if user.CreateUserFactory("").Register(phone, md5_encrypt.Base64Md5(password), userIp) {
		response.Success(ctx, consts.CurdStatusOkMsg, consts.CurdRegisterOkMsg)
	} else {
		response.Fail(ctx, consts.CurdRegisterFailCode, consts.CurdRegisterFailMsg, "")
	}
}

// Login 处理用户登录并签发 Token。
func (u *UserController) Login(ctx *gin.Context) {
	var phone = ctx.GetString(consts.ValidatorPrefix + "phone")
	var password = ctx.GetString(consts.ValidatorPrefix + "password")
	userModel, ok := user.CreateUserFactory("").Login(phone, password)
	if ok {
		userTokenFactory := userstoken.CreateUserFactory()
		if userToken, err := userTokenFactory.GenerateToken(userModel.UID, userModel.Nickname, userModel.Phone, variable.ConfigYml.GetInt64("Token.JwtTokenCreatedExpireAt")); err == nil {
			if userTokenFactory.RecordLoginToken(userToken, ctx.ClientIP()) {
				response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
					"isExist": true,
					"uid":     strconv.FormatInt(userModel.UID, 10),
					"token":   userToken,
				})
			} else {
				response.Fail(ctx, consts.CurdLoginFailCode, "Token 记录失败，请检查数据表 tb_auth_access_tokens 是否存在", gin.H{
					"isExist": true,
					"uid":     strconv.FormatInt(userModel.UID, 10),
					"token":   "",
				})
			}

		} else {
			variable.ZapLog.Error("生成 token 出错")
		}
	} else {
		response.Fail(ctx, consts.CurdLoginFailCode, consts.CurdLoginFailMsg, gin.H{
			"isExist": false,
			"uid":     strconv.FormatInt(userModel.UID, 10),
			"token":   "",
		})
	}
}

// UpdateInfo 更新当前登录用户的资料字段。
func (u *UserController) UpdateInfo(ctx *gin.Context) {
	uid := auth.GetUidFromToken(ctx)
	var operationType = ctx.GetFloat64(consts.ValidatorPrefix + "operation_type")
	var data = ctx.GetString(consts.ValidatorPrefix + "data")
	updateState := user.CreateUserFactory("").UpdateInfo(uid, int(operationType), data)
	if updateState {
		response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
			"data": updateState,
			"msg":  "修改成功",
		})
	} else {
		response.Fail(ctx, consts.CurdUpdateFailCode, consts.CurdUpdateFailMsg, gin.H{
			"data": updateState,
			"msg":  "修改失败",
		})
	}
}

// Attention 处理关注或取消关注用户的操作。
func (u *UserController) Attention(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	var followingId = ctx.GetString(consts.ValidatorPrefix + "following_id")
	var action = ctx.GetBool(consts.ValidatorPrefix + "action")
	var followingIdInt64, _ = strconv.ParseInt(followingId, 10, 64)
	actionStatus := user.CreateUserFactory("").Attention(uid, followingIdInt64, action)
	if actionStatus {
		if action {
			response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
				"data": actionStatus,
				"msg":  "关注成功",
			})
		} else {
			response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
				"data": actionStatus,
				"msg":  "取消关注成功",
			})
		}
	} else {
		if action {
			response.Fail(ctx, consts.CurdInsertFailCode, consts.CurdInsertFailMsg, gin.H{
				"data": actionStatus,
				"msg":  "关注失败",
			})
		} else {
			response.Fail(ctx, consts.CurdDeleteFailCode, consts.CurdDeleteFailMsg, gin.H{
				"data": actionStatus,
				"msg":  "取消关注失败",
			})
		}
	}
}

// AwemeStatus 获取当前用户作品相关状态统计。
func (u *UserController) AwemeStatus(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	awemeStatus, success := user.CreateUserFactory("").AwemeStatus(uid)
	if success {
		response.Success(ctx, consts.CurdStatusOkMsg, awemeStatus)
	} else {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, awemeStatus)
	}
}

// GetUserVideoList 获取指定用户的视频列表。
func (u *UserController) GetUserVideoList(ctx *gin.Context) {
	uid, _ := strconv.Atoi(ctx.Query("uid"))
	videoList, ok := video.CreateVideoFactory("").GetUserVideoList(int64(uid))
	if ok {
		response.Success(ctx, consts.CurdStatusOkMsg, videoList)
	} else {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// GetPanel 获取当前用户个人主页面板数据。
func (u *UserController) GetPanel(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	userinfo, ok := user.CreateUserFactory("").GetPanel(uid)
	if ok {
		response.Success(ctx, consts.CurdStatusOkMsg, userinfo)
	} else {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// GetFriends 获取当前用户好友列表。
func (u *UserController) GetFriends(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	friends, ok := user.CreateUserFactory("").GetFriends(uid)
	if ok {
		response.Success(ctx, consts.CurdStatusOkMsg, friends)
	} else {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// GetFollow 获取当前用户关注列表。
func (u *UserController) GetFollow(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	follow, ok := user.CreateUserFactory("").GetFollow(uid)
	if ok {
		response.Success(ctx, consts.CurdStatusOkMsg, follow)
	} else {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// GetFans 获取当前用户粉丝列表。
func (u *UserController) GetFans(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	fans, ok := user.CreateUserFactory("").GetFans(uid)
	if ok {
		response.Success(ctx, consts.CurdStatusOkMsg, fans)
	} else {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// GetMyVideo 获取当前用户已发布的视频列表。
func (u *UserController) GetMyVideo(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	var PageNo = ctx.GetFloat64(consts.ValidatorPrefix + "pageNo")
	var PageSize = ctx.GetFloat64(consts.ValidatorPrefix + "pageSize")
	list, total, ok := video.CreateVideoFactory("").GetMyVideo(uid, int64(PageNo), int64(PageSize))
	if ok {
		response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
			"pageNo": PageNo,
			"total":  total,
			"list":   list,
		})
	} else {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// GetMyPrivateVideo 获取当前用户私密视频列表。
func (u *UserController) GetMyPrivateVideo(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	var PageNo = ctx.GetFloat64(consts.ValidatorPrefix + "pageNo")
	var PageSize = ctx.GetFloat64(consts.ValidatorPrefix + "pageSize")
	list, total, ok := video.CreateVideoFactory("").GetMyPrivateVideo(uid, int64(PageNo), int64(PageSize))
	if ok {
		response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
			"pageNo": PageNo,
			"total":  total,
			"list":   list,
		})
	} else {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// DeleteMyVideo 删除当前用户自己发布的视频。
func (u *UserController) DeleteMyVideo(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	var awemeIdStr = ctx.GetString(consts.ValidatorPrefix + "aweme_id")
	awemeId, err := strconv.ParseInt(awemeIdStr, 10, 64)
	if err != nil {
		response.Fail(ctx, consts.CurdDeleteFailCode, "Invalid aweme_id", "")
		return
	}

	if video.CreateVideoFactory("").DeleteMyVideo(uid, awemeId) {
		response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
			"aweme_id": strconv.FormatInt(awemeId, 10),
		})
	} else {
		response.Fail(ctx, consts.CurdDeleteFailCode, consts.CurdDeleteFailMsg, "")
	}
}

// GetMyLikeVideo 获取当前用户点赞过的视频列表。
func (u *UserController) GetMyLikeVideo(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	var PageNo = ctx.GetFloat64(consts.ValidatorPrefix + "pageNo")
	var PageSize = ctx.GetFloat64(consts.ValidatorPrefix + "pageSize")
	list, total, ok := video.CreateVideoFactory("").GetMyLikeVideo(uid, int64(PageNo), int64(PageSize))
	if ok {
		response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
			"pageNo": PageNo,
			"total":  total,
			"list":   list,
		})
	} else {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// GetMyCollectVideo 获取当前用户收藏的视频列表。
func (u *UserController) GetMyCollectVideo(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	var PageNo = ctx.GetFloat64(consts.ValidatorPrefix + "pageNo")
	var PageSize = ctx.GetFloat64(consts.ValidatorPrefix + "pageSize")
	list, total, ok := video.CreateVideoFactory("").GetMyCollectVideo(uid, int64(PageNo), int64(PageSize))

	if ok {
		response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
			"video": gin.H{
				"pageNo": PageNo,
				"total":  total,
				"list":   list,
			},
			"music": gin.H{
				"list":  []interface{}{},
				"total": 0,
			},
		})
	} else {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// GetMyHistoryVideo 获取当前用户的视频观看历史。
func (u *UserController) GetMyHistoryVideo(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	var PageNo = ctx.GetFloat64(consts.ValidatorPrefix + "pageNo")
	var PageSize = ctx.GetFloat64(consts.ValidatorPrefix + "pageSize")

	list, total, ok := video.CreateVideoFactory("").GetMyHistoryVideo(uid, int64(PageNo), int64(PageSize))
	if ok {
		response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
			"pageNo": PageNo,
			"total":  total,
			"list":   list,
		})
	} else {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "")
	}
}

// GetMyHistoryOther 获取当前用户的其他历史记录。
func (u *UserController) GetMyHistoryOther(ctx *gin.Context) {
	response.Success(ctx, consts.CurdStatusOkMsg, "GetMyHistoryOther-ok")
}
