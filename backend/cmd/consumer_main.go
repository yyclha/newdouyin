package main

import (
	"douyin-backend/app/model/video"
	videodiggmq "douyin-backend/app/service/video_digg_mq"
	_ "douyin-backend/bootstrap"
)

// main 启动视频点赞异步事件消费者进程。
func main() {
	_ = videodiggmq.RunVideoDiggConsumer(func(event videodiggmq.VideoDiggEvent) error {
		return video.CreateDiggFactory("").HandleAsyncDiggEvent(event)
	})
}
