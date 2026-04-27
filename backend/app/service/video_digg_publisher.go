package service

import (
	"douyin-backend/app/global/variable"
	"fmt"
	"github.com/goccy/go-json"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
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
	if err = ch.Confirm(false); err != nil {
		return err
	}
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	queueName := variable.ConfigYml.GetString("RabbitMq.VideoDigg.QueueName")
	if queueName == "" {
		queueName = defaultVideoDiggQueueName
	}
	durable := variable.ConfigYml.GetBool("RabbitMq.VideoDigg.Durable")

	deadLetterQueueName := videoDiggDeadLetterQueueName()
	if _, err = ch.QueueDeclare(deadLetterQueueName, durable, !durable, false, false, nil); err != nil {
		return err
	}
	if _, err = ch.QueueDeclare(queueName, durable, !durable, false, false, amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": deadLetterQueueName,
	}); err != nil {
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

	if err = ch.Publish("", queueName, false, false, amqp.Publishing{
		DeliveryMode: deliveryMode,
		ContentType:  "application/json",
		Body:         body,
	}); err != nil {
		return err
	}

	confirmTimeout := variable.ConfigYml.GetDuration("RabbitMq.VideoDigg.ConfirmTimeoutSec")
	if confirmTimeout <= 0 {
		confirmTimeout = 5
	}
	select {
	case confirm := <-confirms:
		if confirm.Ack {
			return nil
		}
		return fmt.Errorf("rabbitmq publish not acknowledged")
	case <-time.After(confirmTimeout * time.Second):
		return fmt.Errorf("rabbitmq publish confirm timeout")
	}
}
