package main

import (
	"douyin-backend/app/global/variable" // 项目编译之前加载全局变量
	"douyin-backend/app/model/video"
	videodiggmq "douyin-backend/app/service/video_digg_mq"
	_ "douyin-backend/bootstrap" // 项目初始化
	"douyin-backend/routers"
)

// 后端路由启动入口
func main() {
	go func() {
		if err := videodiggmq.RunVideoDiggConsumer(func(event videodiggmq.VideoDiggEvent) error {
			return video.CreateDiggFactory("").HandleAsyncDiggEvent(event)
		}); err != nil {
			variable.ZapLog.Error("视频点赞消费者退出:" + err.Error())
		}
	}()

	router := routers.InitWebRouter()
	//fmt.Println(router.RouterGroup.Handlers)
	_ = router.Run(variable.ConfigYml.GetString("HttpServer.Web.Port"))
}
