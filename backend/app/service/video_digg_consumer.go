package service

import (
	"douyin-backend/app/global/variable"
	"github.com/goccy/go-json"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"time"
)

// RunVideoDiggConsumer 启动视频点赞事件消费者，并在异常退出时自动重连。
func RunVideoDiggConsumer(handler func(event VideoDiggEvent) error) error {
	reconnectInterval := variable.ConfigYml.GetDuration("RabbitMq.VideoDigg.ReconnectIntervalSec")
	if reconnectInterval <= 0 {
		reconnectInterval = 3
	}

	for {
		if err := consumeVideoDiggEvents(handler); err != nil {
			variable.ZapLog.Error("video digg consumer stopped", zap.Error(err))
			time.Sleep(reconnectInterval * time.Second)
			continue
		}
		return nil
	}
}

// consumeVideoDiggEvents 持续消费 RabbitMQ 中的视频点赞事件消息。
func consumeVideoDiggEvents(handler func(event VideoDiggEvent) error) error {
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
	prefetchCount := variable.ConfigYml.GetInt("RabbitMq.VideoDigg.PrefetchCount")
	if prefetchCount <= 0 {
		prefetchCount = 1
	}

	if _, err = ch.QueueDeclare(queueName, durable, !durable, false, false, nil); err != nil {
		return err
	}
	if err = ch.Qos(prefetchCount, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for msg := range msgs {
		var event VideoDiggEvent
		if err = json.Unmarshal(msg.Body, &event); err != nil {
			variable.ZapLog.Error("video digg consumer failed to decode message", zap.Error(err))
			_ = msg.Ack(false)
			continue
		}

		if err = handler(event); err != nil {
			variable.ZapLog.Error("video digg consumer failed to handle message", zap.Error(err), zap.Int64("uid", event.UID), zap.Int64("aweme_id", event.AwemeID))
			_ = msg.Nack(false, true)
			continue
		}

		_ = msg.Ack(false)
	}

	return amqp.ErrClosed
}
