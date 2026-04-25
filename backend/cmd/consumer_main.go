package main

import (
	"douyin-backend/app/model/video"
	"douyin-backend/app/service"
	_ "douyin-backend/bootstrap"
)

// main 启动视频点赞异步事件消费者进程。
func main() {
	_ = service.RunVideoDiggConsumer(func(event service.VideoDiggEvent) error {
		return video.CreateDiggFactory("").HandleAsyncDiggEvent(event)
	})
}
