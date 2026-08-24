package notifier

import "context"

// Reminder is a single outgoing reorder reminder/promo message.
type Reminder struct {
	To      string // phone number (WhatsApp E.164, e.g. 6281234567890)
	Message string // rendered message body
}

// Notifier sends reorder reminder notifications to a customer.
type Notifier interface {
	SendReminder(ctx context.Context, r Reminder) error
}
