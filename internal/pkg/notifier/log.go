package notifier

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// LogNotifier logs the message instead of sending it. Useful for local
// development and as a fallback when no gateway token is configured.
type LogNotifier struct {
	Log *logrus.Logger
}

func (n *LogNotifier) SendReminder(ctx context.Context, r Reminder) error {
	log := n.Log
	if log == nil {
		log = logrus.New()
	}
	log.WithContext(ctx).
		WithField("to", r.To).
		WithField("message", r.Message).
		Info("reorder reminder (not sent, log-only)")
	return nil
}

// NewNotifier returns a Fonnte notifier when a token is configured, otherwise
// a log-only fallback so the system keeps running without a gateway.
func NewNotifier(token string, target string, log *logrus.Logger) (Notifier, error) {
	if token == "" {
		return &LogNotifier{Log: log}, fmt.Errorf("fonnte token empty; using log-only notifier")
	}
	return NewFonnte(FonnteConfig{Token: token, Target: target}), nil
}
