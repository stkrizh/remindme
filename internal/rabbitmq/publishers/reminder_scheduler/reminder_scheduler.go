package reminderscheduler

import (
	"context"
	"fmt"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/rabbitmq"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	log        logging.Logger
	channel    *rabbitmq.Channel
	exchange   string
	routingKey string
}

func NewRabbitMQ(log logging.Logger, channel *rabbitmq.Channel, exchange string, routingKey string) *RabbitMQ {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if channel == nil {
		panic(e.NewNilArgumentError("channel"))
	}
	return &RabbitMQ{log: log, channel: channel, exchange: exchange, routingKey: routingKey}
}

func (s *RabbitMQ) ScheduleReminder(ctx context.Context, r reminder.Reminder) error {
	err := s.channel.PublishWithContext(ctx, s.exchange, s.routingKey, false, false, amqp091.Publishing{
		Headers:     amqp091.Table{"x-delay": 10_000},
		ContentType: "application/json",
		Body:        []byte(fmt.Sprintf("%d", r.ID)),
	})
	if err != nil {
		logging.Error(ctx, s.log, err)
		return err
	}
	s.log.Info(
		ctx,
		"AMQP message has been successfully published.",
		logging.Entry("exchange", s.exchange),
		logging.Entry("RK", s.routingKey),
		logging.Entry("reminderID", r.ID),
	)
	return nil
}
