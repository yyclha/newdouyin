package register_validator

import (
	"douyin-backend/app/core/container"
	"douyin-backend/app/global/consts"
	"douyin-backend/app/http/validator/common/websocket"
	"douyin-backend/app/http/validator/web/douyin/message"
	"douyin-backend/app/http/validator/web/douyin/post"
	"douyin-backend/app/http/validator/web/douyin/shop"
	"douyin-backend/app/http/validator/web/douyin/upload"
	"douyin-backend/app/http/validator/web/douyin/user"
	"douyin-backend/app/http/validator/web/douyin/video"
)

// WebRegisterValidator 执行业务处理。
func WebRegisterValidator() {
	containers := container.CreateContainersFactory()

	var key string
	// base
	{
		key = consts.ValidatorPrefix + "Login"
		containers.Set(key, user.Login{})

		key = consts.ValidatorPrefix + "Register"
		containers.Set(key, user.Register{})
	}
	// upload
	{
		key = consts.ValidatorPrefix + "Avatar"
		containers.Set(key, upload.Avatar{})

		key = consts.ValidatorPrefix + "Cover"
		containers.Set(key, upload.Cover{})

		key = consts.ValidatorPrefix + "VideoInit"
		containers.Set(key, upload.VideoInit{})

		key = consts.ValidatorPrefix + "VideoChunk"
		containers.Set(key, upload.VideoChunk{})

		key = consts.ValidatorPrefix + "VideoComplete"
		containers.Set(key, upload.VideoComplete{})

		key = consts.ValidatorPrefix + "VideoStatus"
		containers.Set(key, upload.VideoStatus{})
	}

	// user
	{
		key = consts.ValidatorPrefix + "UpdateInfo"
		containers.Set(key, user.UpdateInfo{})

		key = consts.ValidatorPrefix + "GetUserVideoList"
		containers.Set(key, user.GetUserVideoList{})

		key = consts.ValidatorPrefix + "GetPanel"
		containers.Set(key, user.GetPanel{})

		key = consts.ValidatorPrefix + "Attention"
		containers.Set(key, user.Attention{})

		key = consts.ValidatorPrefix + "AwemeStatus"
		containers.Set(key, user.AwemeStatus{})

		key = consts.ValidatorPrefix + "GetFriends"
		containers.Set(key, user.GetFriends{})

		key = consts.ValidatorPrefix + "GetFollow"
		containers.Set(key, user.GetFollow{})

		key = consts.ValidatorPrefix + "GetFans"
		containers.Set(key, user.GetFans{})

		key = consts.ValidatorPrefix + "GetMyVideo"
		containers.Set(key, user.GetMyVideo{})

		key = consts.ValidatorPrefix + "GetMyPrivateVideo"
		containers.Set(key, user.GetMyPrivateVideo{})

		key = consts.ValidatorPrefix + "GetMyLikeVideo"
		containers.Set(key, user.GetMyLikeVideo{})

		key = consts.ValidatorPrefix + "GetMyCollectVideo"
		containers.Set(key, user.GetMyCollectVideo{})

		key = consts.ValidatorPrefix + "GetMyHistoryVideo"
		containers.Set(key, user.GetMyHistoryVideo{})

		key = consts.ValidatorPrefix + "GetMyHistoryOther"
		containers.Set(key, user.GetMyHistoryOther{})

		key = consts.ValidatorPrefix + "DeleteMyVideo"
		containers.Set(key, user.DeleteMyVideo{})

	}
	// video
	{
		key = consts.ValidatorPrefix + "GetVideoRecommended"
		containers.Set(key, video.GetVideoRecommended{})

		key = consts.ValidatorPrefix + "GetLongVideoRecommended"
		containers.Set(key, video.GetLongVideoRecommended{})

		key = consts.ValidatorPrefix + "GetComments"
		containers.Set(key, video.GetComments{})

		key = consts.ValidatorPrefix + "VideoDigg"
		containers.Set(key, video.VideoDigg{})

		key = consts.ValidatorPrefix + "VideoComment"
		containers.Set(key, video.VideoComment{})

		key = consts.ValidatorPrefix + "CommentDigg"
		containers.Set(key, video.CommentDigg{})

		key = consts.ValidatorPrefix + "DeleteComment"
		containers.Set(key, video.DeleteComment{})

		key = consts.ValidatorPrefix + "VideoCollect"
		containers.Set(key, video.VideoCollect{})

		key = consts.ValidatorPrefix + "VideoShare"
		containers.Set(key, video.VideoShare{})
	}
	// shop
	{
		key = consts.ValidatorPrefix + "GetShopRecommended"
		containers.Set(key, shop.GetShopRecommended{})
	}

	// post
	{
		key = consts.ValidatorPrefix + "GetPostRecommended"
		containers.Set(key, post.GetPostRecommended{})
	}
	// msg
	{
		key = consts.ValidatorPrefix + "WebsocketConnect"
		containers.Set(key, websocket.Connect{})

		key = consts.ValidatorPrefix + "AllMsg"
		containers.Set(key, message.AllMsg{})
	}

}
