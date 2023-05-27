package response

import (
	"remindme/internal/core/domain/user"
)

type Limits struct {
	EmailChannelCount        *int64   `json:"email_channel_count,omitempty"`
	TelegramChannelCount     *int64   `json:"telegram_channel_count,omitempty"`
	ActiveReminderCount      *int64   `json:"active_reminder_count,omitempty"`
	MonthlySentReminderCount *int64   `json:"monthly_sent_reminder_count,omitempty"`
	ReminderEveryPerDayCount *float64 `json:"reminder_every_per_day_count,omitempty"`
}

func (l *Limits) FromDomainLimits(dl user.Limits) {
	if dl.EmailChannelCount.IsPresent {
		v := int64(dl.EmailChannelCount.Value)
		l.EmailChannelCount = &v
	}
	if dl.TelegramChannelCount.IsPresent {
		v := int64(dl.TelegramChannelCount.Value)
		l.TelegramChannelCount = &v
	}
	if dl.ActiveReminderCount.IsPresent {
		v := int64(dl.ActiveReminderCount.Value)
		l.ActiveReminderCount = &v
	}
	if dl.MonthlySentReminderCount.IsPresent {
		v := int64(dl.MonthlySentReminderCount.Value)
		l.MonthlySentReminderCount = &v
	}
	if dl.ReminderEveryPerDayCount.IsPresent {
		l.ReminderEveryPerDayCount = &dl.ReminderEveryPerDayCount.Value
	}
}
