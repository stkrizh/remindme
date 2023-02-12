package main

import (
	"context"
	"os"
	"os/signal"
	"remindme/internal/app/deps"
	"remindme/internal/app/services"
	"remindme/internal/core/domain/logging"
	schedulereminders "remindme/internal/core/services/schedule_reminders"
	"syscall"
	"time"
)

func main() {
	deps, shutdownDeps := deps.InitDeps()
	log := deps.Logger
	defer shutdownDeps()

	services := services.InitServices(deps)

	ticker := time.NewTicker(deps.Config.RemindersSchedulingPeriod)
	defer ticker.Stop()

	stopCh, closeCh := createChannel()
	defer closeCh()

	log.Info(
		context.Background(),
		"Starting periodic reminder scheduler.",
		logging.Entry("periodMinutes", (deps.Config.RemindersSchedulingPeriod).Minutes()),
	)

loop:
	for {
		select {
		case <-stopCh:
			log.Info(context.Background(), "Stopping periodic reminde scheduler.")
			break loop
		case <-ticker.C:
			log.Info(context.Background(), "Launching reminders scheduling service.")
			_, err := services.ScheduleReminders.Run(context.Background(), schedulereminders.Input{})
			if err != nil {
				log.Error(context.Background(), "Scheduling service returned an error.", logging.Entry("err", err))
			}
		}
	}
}

func createChannel() (chan os.Signal, func()) {
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	return stopCh, func() {
		close(stopCh)
	}
}
