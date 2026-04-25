package service

import (
	"douyin-backend/app/global/variable"
	"github.com/goccy/go-json"
	amqp "github.com/rabbitmq/amqp091-go"
)

// defaultVideoDiggQueueName 定义视频点赞事件的默认消息队列名。
const defaultVideoDiggQueueName = "video_digg_queue"

// PublishVideoDiggEvent 将视频点赞事件发布到 RabbitMQ 队列。
func PublishVideoDiggEvent(event VideoDiggEvent) error {
	conn, err := amqp.Dial(variable.ConfigYml.GetString("RabbitMq.VideoDigg.Addr"))
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	queueName := variable.ConfigYml.GetString("RabbitMq.VideoDigg.QueueName")
	if queueName == "" {
		queueName = defaultVideoDiggQueueName
	}
	durable := variable.ConfigYml.GetBool("RabbitMq.VideoDigg.Durable")

	if _, err = ch.QueueDeclare(queueName, durable, !durable, false, false, nil); err != nil {
		return err
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	deliveryMode := amqp.Transient
	if durable {
		deliveryMode = amqp.Persistent
	}

	return ch.Publish("", queueName, false, false, amqp.Publishing{
		DeliveryMode: deliveryMode,
		ContentType:  "application/json",
		Body:         body,
	})
}
