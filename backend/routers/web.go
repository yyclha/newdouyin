package routers

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/global/variable"
	"douyin-backend/app/http/middleware/authorization"
	"douyin-backend/app/http/middleware/cors"
	validatorFactory "douyin-backend/app/http/validator/core/factory"
	"douyin-backend/app/utils/gin_release"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
)

// InitWebRouter 初始化 HTTP 路由、全局中间件和所有 Web 端接口。
func InitWebRouter() *gin.Engine {
	var router *gin.Engine
	// 非调试模式下将 Gin 访问日志写入文件。
	if variable.ConfigYml.GetBool("AppDebug") == false {
		gin.DisableConsoleColor()
		f, _ := os.Create(variable.BasePath + variable.ConfigYml.GetString("Logs.GinLogName"))
		gin.DefaultWriter = io.MultiWriter(f)
		router = gin_release.ReleaseRouter()
	} else {
		// 调试模式下启用 pprof，便于排查性能问题。
		router = gin.Default()
		pprof.Register(router)
	}

	// 根据配置设置可信代理列表。
	if variable.ConfigYml.GetInt("HTTPServer.TrustProxies.IsOpen") == 1 {
		if err := router.SetTrustedProxies(variable.ConfigYml.GetStringSlice("HttpServer.TrustProxies.ProxyServerList")); err != nil {
			variable.ZapLog.Error(consts.GinSetTrustProxyError, zap.Error(err))
		}
	} else {
		_ = router.SetTrustedProxies(nil)
	}

	// 根据配置启用跨域中间件。
	if variable.ConfigYml.GetBool("HttpServer.AllowCrossDomain") {
		router.Use(cors.Next())
	}

	router.GET("/", func(context *gin.Context) {
		context.String(http.StatusOK, "douyin-backend")
	})

	// GET /public/*path 提供后端生成的静态资源访问，例如图片、封面和上传文件。
	router.Static("/public", "./public")

	auth := router.Group("base/")
	{
		// POST /base/register 用户注册。
		auth.POST("register", validatorFactory.Create(consts.ValidatorPrefix+"Register"))
		// POST /base/login 用户登录并返回 Token。
		auth.POST("login", validatorFactory.Create(consts.ValidatorPrefix+"Login"))
	}

	// GET /message/ws 建立消息系统的 WebSocket 长连接。
	router.GET("message/ws", validatorFactory.Create(consts.ValidatorPrefix+"WebsocketConnect"))
	// GET /video/comments 获取指定视频的评论列表。
	router.GET("video/comments", validatorFactory.Create(consts.ValidatorPrefix+"GetComments"))

	router.Use(authorization.CheckTokenAuth())

	upload := router.Group("upload/")
	{
		// POST /upload/avatar 上传用户头像。
		upload.POST("avatar", validatorFactory.Create(consts.ValidatorPrefix+"Avatar"))
		// POST /upload/cover 上传视频封面。
		upload.POST("cover", validatorFactory.Create(consts.ValidatorPrefix+"Cover"))
		// POST /upload/video/init 初始化视频分片上传任务。
		upload.POST("video/init", validatorFactory.Create(consts.ValidatorPrefix+"VideoInit"))
		// POST /upload/video/chunk 上传单个视频分片。
		upload.POST("video/chunk", validatorFactory.Create(consts.ValidatorPrefix+"VideoChunk"))
		// POST /upload/video/complete 合并分片并完成视频上传。
		upload.POST("video/complete", validatorFactory.Create(consts.ValidatorPrefix+"VideoComplete"))
		// GET /upload/video/status 查询后台视频处理状态。
		upload.GET("video/status", validatorFactory.Create(consts.ValidatorPrefix+"VideoStatus"))
	}

	user := router.Group("user/")
	{
		// POST /user/update-info 更新当前用户资料。
		user.POST("update-info", validatorFactory.Create(consts.ValidatorPrefix+"UpdateInfo"))
		// GET /user/video-list 获取指定用户的视频列表。
		user.GET("video-list", validatorFactory.Create(consts.ValidatorPrefix+"GetUserVideoList"))
		// GET /user/panel 获取当前用户个人主页面板信息。
		user.GET("panel", validatorFactory.Create(consts.ValidatorPrefix+"GetPanel"))
		// GET /user/friends 获取当前用户好友列表。
		user.GET("friends", validatorFactory.Create(consts.ValidatorPrefix+"GetFriends"))
		// GET /user/follow 获取当前用户关注列表。
		user.GET("follow", validatorFactory.Create(consts.ValidatorPrefix+"GetFollow"))
		// GET /user/fans 获取当前用户粉丝列表。
		user.GET("fans", validatorFactory.Create(consts.ValidatorPrefix+"GetFans"))

		// POST /user/attention 关注或取消关注指定用户。
		user.POST("attention", validatorFactory.Create(consts.ValidatorPrefix+"Attention"))
		// GET /user/aweme-status 获取当前用户作品相关状态统计。
		user.GET("aweme-status", validatorFactory.Create(consts.ValidatorPrefix+"AwemeStatus"))
		// GET /user/my-video 获取当前用户发布的视频列表。
		user.GET("my-video", validatorFactory.Create(consts.ValidatorPrefix+"GetMyVideo"))
		// GET /user/my-private 获取当前用户私密视频列表。
		user.GET("my-private", validatorFactory.Create(consts.ValidatorPrefix+"GetMyPrivateVideo"))
		// GET /user/my-like-video 获取当前用户点赞过的视频列表。
		user.GET("my-like-video", validatorFactory.Create(consts.ValidatorPrefix+"GetMyLikeVideo"))
		// GET /user/my-collect-video 获取当前用户收藏的视频列表。
		user.GET("my-collect-video", validatorFactory.Create(consts.ValidatorPrefix+"GetMyCollectVideo"))
		// GET /user/my-history-video 获取当前用户观看历史中的视频列表。
		user.GET("my-history-video", validatorFactory.Create(consts.ValidatorPrefix+"GetMyHistoryVideo"))
		// GET /user/my-history-other 获取当前用户观看历史中的其他内容。
		user.GET("my-history-other", validatorFactory.Create(consts.ValidatorPrefix+"GetMyHistoryOther"))
		// POST /user/delete-video 删除当前用户自己发布的视频。
		user.POST("delete-video", validatorFactory.Create(consts.ValidatorPrefix+"DeleteMyVideo"))
	}

	post := router.Group("post/")
	{
		// GET /post/recommended 获取推荐图文动态列表。
		post.GET("recommended", validatorFactory.Create(consts.ValidatorPrefix+"GetPostRecommended"))
	}

	shop := router.Group("shop/")
	{
		// GET /shop/recommended 获取推荐商品列表。
		shop.GET("recommended", validatorFactory.Create(consts.ValidatorPrefix+"GetShopRecommended"))
	}

	video := router.Group("video/")
	{
		// POST /video/digg 对视频点赞或取消点赞。
		video.POST("digg", validatorFactory.Create(consts.ValidatorPrefix+"VideoDigg"))
		// POST /video/comment 发表评论。
		video.POST("comment", validatorFactory.Create(consts.ValidatorPrefix+"VideoComment"))
		// POST /video/comment-digg 对评论点赞或取消点赞。
		video.POST("comment-digg", validatorFactory.Create(consts.ValidatorPrefix+"CommentDigg"))
		// POST /video/delete-comment 删除评论。
		video.POST("delete-comment", validatorFactory.Create(consts.ValidatorPrefix+"DeleteComment"))
		// POST /video/collect 收藏或取消收藏视频。
		video.POST("collect", validatorFactory.Create(consts.ValidatorPrefix+"VideoCollect"))
		// POST /video/share 将视频分享给其他用户。
		video.POST("share", validatorFactory.Create(consts.ValidatorPrefix+"VideoShare"))
		// GET /video/recommended 获取推荐短视频流。
		video.GET("recommended", validatorFactory.Create(consts.ValidatorPrefix+"GetVideoRecommended"))
		// GET /video/long-recommended 获取推荐长视频列表。
		video.GET("long-recommended", validatorFactory.Create(consts.ValidatorPrefix+"GetLongVideoRecommended"))
	}

	msg := router.Group("message")
	{
		// GET /message/all-msg 获取当前用户消息会话列表。
		msg.GET("all-msg", validatorFactory.Create(consts.ValidatorPrefix+"AllMsg"))
	}
	return router
}
