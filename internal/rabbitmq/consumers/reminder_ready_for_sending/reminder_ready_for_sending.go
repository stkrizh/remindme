package reminderreadyforsending

import (
	"context"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/services"
	sendreminder "remindme/internal/core/services/send_reminder"
	"remindme/internal/rabbitmq"
	"remindme/internal/rabbitmq/schema"

	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	log     logging.Logger
	channel *rabbitmq.Channel
	queue   string
	service services.Service[sendreminder.Input, sendreminder.Result]
}

func New(
	log logging.Logger,
	channel *rabbitmq.Channel,
	queue string,
	service services.Service[sendreminder.Input, sendreminder.Result],
) *Consumer {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if channel == nil {
		panic(e.NewNilArgumentError("channel"))
	}
	if queue == "" {
		panic("queue name must not be empty")
	}
	if service == nil {
		panic(e.NewNilArgumentError("service"))
	}

	return &Consumer{log: log, channel: channel, queue: queue, service: service}
}

func (c *Consumer) Consume() error {
	deliveries, err := c.channel.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		c.log.Error(context.Background(), "Could not start cosuming.", logging.Entry("err", err))
		return err
	}

	go func() {
		for delivery := range deliveries {
			rem := &schema.Reminder{}
			if err := rem.Unmarshal(delivery.Body); err != nil {
				c.log.Error(
					context.Background(),
					"Could not unmarshal reminder.",
					logging.Entry("err", err),
					logging.Entry("delivery", delivery),
				)
				c.Ack(delivery)
				continue
			}

			c.log.Info(
				context.Background(),
				"Got ready for sending reminder.",
				logging.Entry("reminder", rem),
			)
			_, err := c.service.Run(
				context.Background(),
				sendreminder.Input{ReminderID: reminder.ID(rem.ID), At: rem.At},
			)
			if err != nil {
				c.log.Error(
					context.Background(),
					"Could not send reminder, service returned an error.",
					logging.Entry("reminder", rem),
					logging.Entry("err", err),
				)
			}
			c.Ack(delivery)
		}
	}()
	return nil
}

func (c *Consumer) Ack(delivery amqp091.Delivery) {
	if err := delivery.Ack(true); err != nil {
		c.log.Error(context.Background(), "Could not ACK AMQP message.", logging.Entry("err", err))
	}
}
