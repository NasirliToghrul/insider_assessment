package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"case_study/internal/config"
	"case_study/internal/models"
)

type WebhookRequest struct {
	To      string `json:"to"`
	Content string `json:"content"`
}

type WebhookResponse struct {
	Message   string `json:"message"`
	MessageID string `json:"messageId"`
}

type HTTPSender struct {
	client *http.Client
	cfg    config.Config
}

func NewHTTPSender(cfg config.Config) *HTTPSender {
	return &HTTPSender{cfg: cfg, client: &http.Client{Timeout: 20 * time.Second}}
}

func (s *HTTPSender) Send(ctx context.Context, m models.Message) (string, int, []byte, error) {
	payload := WebhookRequest{To: m.To, Content: m.Content}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.WebhookURL, bytes.NewReader(b))
	if err != nil {
		return "", 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.WebhookKey != "" {
		req.Header.Set("x-ins-auth-key", s.cfg.WebhookKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", 0, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	// Expect 202 Accepted per spec
	if resp.StatusCode != http.StatusAccepted {
		return "", resp.StatusCode, body, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var wr WebhookResponse
	if err := json.Unmarshal(body, &wr); err != nil {
		return "", resp.StatusCode, body, fmt.Errorf("decode error: %w", err)
	}
	if wr.MessageID == "" {
		return "", resp.StatusCode, body, fmt.Errorf("empty messageId in response")
	}
	return wr.MessageID, resp.StatusCode, body, nil
}
