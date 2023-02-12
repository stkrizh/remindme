package consumers

import (
	"context"
	"remindme/internal/app/deps"
	"remindme/internal/app/services"
	dl "remindme/internal/core/domain/logging"
	reminderreadyforsending "remindme/internal/rabbitmq/consumers/reminder_ready_for_sending"
)

func initReminderReadyForSendingConsumer(deps *deps.Deps, services *services.Services) func() {
	rabbitmqChannel, err := deps.Rabbitmq.Channel()
	if err != nil {
		deps.Logger.Error(context.Background(), "Could not create RabbitMQ channel.", dl.Entry("err", err))
		panic(err)
	}

	queue := deps.Config.RabbitmqReminderReadyQueue
	reminderReadyForSendingConsumer := reminderreadyforsending.New(
		deps.Logger,
		rabbitmqChannel,
		queue,
		services.SendReminder,
	)
	if err = reminderReadyForSendingConsumer.Consume(); err != nil {
		deps.Logger.Error(
			context.Background(),
			"Could not start RabbitMQ consuming.",
			dl.Entry("err", err),
			dl.Entry("queue", queue),
		)
		panic(err)
	}

	deps.Logger.Info(context.Background(), "Consumer has started.", dl.Entry("queue", queue))
	return func() { rabbitmqChannel.Close() }
}

func InitConsumers(deps *deps.Deps, services *services.Services) func() {
	shutdownReminderReadyForSendingConsumer := initReminderReadyForSendingConsumer(deps, services)

	return func() {
		shutdownReminderReadyForSendingConsumer()
	}
}
