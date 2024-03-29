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

	c.log.Info(context.Background(), "Reminder sending consumer has started.", logging.Entry("queue", c.queue))
	go func() {
		for delivery := range deliveries {
			c.processDelivery(delivery)
		}
		c.log.Info(context.Background(), "Reminder sending consumer has stopped.", logging.Entry("queue", c.queue))
	}()

	return nil
}

func (c *Consumer) processDelivery(delivery amqp091.Delivery) {
	// defer c.nack(delivery)

	rem := &schema.Reminder{}
	if err := rem.Unmarshal(delivery.Body); err != nil {
		c.log.Error(
			context.Background(),
			"Could not unmarshal reminder.",
			logging.Entry("err", err),
			logging.Entry("delivery", delivery),
		)
		c.ack(delivery)
		return
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
	c.ack(delivery)
}

func (c *Consumer) ack(delivery amqp091.Delivery) {
	if err := delivery.Ack(false); err != nil {
		c.log.Error(context.Background(), "Could not Ack AMQP message.", logging.Entry("err", err))
	}
}

func (c *Consumer) nack(delivery amqp091.Delivery) {
	if err := delivery.Nack(false, true); err != nil {
		c.log.Error(
			context.Background(),
			"Could not Nack AMQP delivery.",
			logging.Entry("err", err),
		)
	}
}
