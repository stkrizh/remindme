package reminderreadyforsending

import (
	"context"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/rabbitmq"
)

type Consumer struct {
	log     logging.Logger
	channel *rabbitmq.Channel
	queue   string
}

func New(log logging.Logger, channel *rabbitmq.Channel, queue string) *Consumer {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if channel == nil {
		panic(e.NewNilArgumentError("channel"))
	}
	if queue == "" {
		panic("queue name must not be empty")
	}

	return &Consumer{log: log, channel: channel, queue: queue}
}

func (c *Consumer) Consume() error {
	deliveries, err := c.channel.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		c.log.Error(context.Background(), "Could not start cosuming.", logging.Entry("err", err))
		return err
	}

	go func() {
		for delivery := range deliveries {
			body := delivery.Body
			if err := delivery.Ack(true); err != nil {
				c.log.Error(context.Background(), "Could not ACK AMQP message.", logging.Entry("err", err))
			}
			c.log.Info(context.Background(), "Got AMQP message.", logging.Entry("msg", body))
		}
	}()
	return nil
}
