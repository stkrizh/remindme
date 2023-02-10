package reminderscheduler

import (
	"context"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/rabbitmq"
	"remindme/internal/rabbitmq/schema"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	log        logging.Logger
	channel    *rabbitmq.Channel
	exchange   string
	routingKey string
	now        func() time.Time
}

func NewRabbitMQ(
	log logging.Logger,
	channel *rabbitmq.Channel,
	exchange string,
	routingKey string,
	now func() time.Time,
) *RabbitMQ {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if channel == nil {
		panic(e.NewNilArgumentError("channel"))
	}
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	return &RabbitMQ{log: log, channel: channel, exchange: exchange, routingKey: routingKey, now: now}
}

func (s *RabbitMQ) ScheduleReminder(ctx context.Context, r reminder.Reminder) error {
	now := s.now()
	delay := r.At.Sub(now).Milliseconds()

	reminder := &schema.Reminder{
		ID: int64(r.ID),
		At: r.At,
	}
	data, err := reminder.Marshal()
	if err != nil {
		s.log.Error(ctx, "Could not marshal reminder.", logging.Entry("err", err))
		return err
	}

	err = s.channel.PublishWithContext(ctx, s.exchange, s.routingKey, false, false, amqp091.Publishing{
		Headers:     amqp091.Table{"x-delay": delay},
		ContentType: "application/json",
		Body:        data,
	})
	if err != nil {
		logging.Error(ctx, s.log, err)
		return err
	}
	s.log.Info(
		ctx,
		"Reminder has been successfully scheduled to RabbitMQ exchange.",
		logging.Entry("exchange", s.exchange),
		logging.Entry("RK", s.routingKey),
		logging.Entry("reminderID", reminder.ID),
		logging.Entry("reminderAt", reminder.At),
	)
	return nil
}
