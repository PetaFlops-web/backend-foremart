package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// FonnteConfig holds the credentials for the Fonnte WhatsApp gateway.
type FonnteConfig struct {
	Token  string // API token
	Target string // default target number (optional; per-message To takes precedence)
}

type fonnteNotifier struct {
	client *http.Client
	cfg    FonnteConfig
}

// NewFonnte builds a WhatsApp notifier backed by the Fonnte gateway.
// baseURL defaults to https://api.fonnte.com.
func NewFonnte(cfg FonnteConfig) Notifier {
	return &fonnteNotifier{
		client: &http.Client{Timeout: 15 * time.Second},
		cfg:    cfg,
	}
}

type fonntePayload struct {
	Target  string `json:"target,omitempty"`
	Message string `json:"message"`
}

func (n *fonnteNotifier) SendReminder(ctx context.Context, r Reminder) error {
	target := r.To
	if target == "" {
		target = n.cfg.Target
	}
	if target == "" {
		return fmt.Errorf("notifier: no target phone number")
	}

	payload := fonntePayload{Target: target, Message: r.Message}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("notifier: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.fonnte.com/send", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notifier: create request: %w", err)
	}
	req.Header.Set("Authorization", n.cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("notifier: call fonnte: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("notifier: fonnte returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
