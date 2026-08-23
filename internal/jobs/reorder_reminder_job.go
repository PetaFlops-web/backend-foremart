package jobs

import (
	"context"
	"time"

	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/notification/src/usecase"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// StartReorderReminderScheduler registers the nightly reorder-reminder job.
// Spec is standard cron: "0 8 * * *" runs daily at 08:00 local time.
func StartReorderReminderScheduler(log *logrus.Logger, notificationUseCase *usecase.NotificationUseCase, spec string) (*cron.Cron, error) {
	if spec == "" {
		spec = "0 8 * * *"
	}

	c := cron.New(cron.WithLocation(time.Local))

	_, err := c.AddFunc(spec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		start := time.Now()
		sent, err := notificationUseCase.RunReminder(ctx)
		if err != nil {
			log.WithError(err).Error("reorder reminder job failed")
			return
		}
		log.WithFields(logrus.Fields{
			"sent":     sent,
			"duration": time.Since(start).String(),
		}).Info("reorder reminder job completed")
	})
	if err != nil {
		return nil, err
	}

	c.Start()
	log.WithField("spec", spec).Info("reorder reminder scheduler started")
	return c, nil
}
