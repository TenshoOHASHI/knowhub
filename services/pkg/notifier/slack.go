package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// SlackNotifier sends messages to a Slack channel via webhook.
type SlackNotifier struct {
	webhookURL string
	client     *http.Client
	enabled    bool
}

// NewSlackNotifier creates a new SlackNotifier.
// If webhookURL is empty, the notifier is disabled (safe for development).
func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		enabled: webhookURL != "",
	}
}

// slackPayload represents the JSON body sent to a Slack webhook.
type slackPayload struct {
	Blocks []slackBlock `json:"blocks,omitempty"`
	Text   string       `json:"text,omitempty"`
}

type slackBlock struct {
	Type     string      `json:"type"`
	Text     *slackText  `json:"text,omitempty"`
	Elements []slackText `json:"elements,omitempty"` // context用
}

type slackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Send posts a raw text message to Slack.
func (n *SlackNotifier) Send(text string) error {
	if !n.enabled {
		return nil
	}
	payload := slackPayload{Text: text}
	return n.post(payload)
}

// NotifyArticleCreated sends a notification when an article is created.
func (n *SlackNotifier) NotifyArticleCreated(title, articleID string) error {
	if !n.enabled {
		return nil
	}
	payload := slackPayload{
		Blocks: []slackBlock{
			{Type: "header", Text: &slackText{Type: "plain_text", Text: ":memo: 記事を作成しました"}},
			{Type: "section", Text: &slackText{Type: "mrkdwn", Text: "*" + title + "*"}},
			{Type: "context", Elements: []slackText{
				{Type: "mrkdwn", Text: "ID: `" + articleID + "`"},
			}},
		},
	}
	return n.post(payload)
}

// NotifyArticleUpdated sends a notification when an article is updated.
func (n *SlackNotifier) NotifyArticleUpdated(title, articleID string) error {
	if !n.enabled {
		return nil
	}
	payload := slackPayload{
		Blocks: []slackBlock{
			{Type: "header", Text: &slackText{Type: "plain_text", Text: ":pencil: 記事を更新しました"}},
			{Type: "section", Text: &slackText{Type: "mrkdwn", Text: "*" + title + "*"}},
			{Type: "context", Elements: []slackText{
				{Type: "mrkdwn", Text: "ID: `" + articleID + "`"},
			}},
		},
	}
	return n.post(payload)
}

// NotifyArticleDeleted sends a notification when an article is deleted.
func (n *SlackNotifier) NotifyArticleDeleted(title, articleID string) error {
	if !n.enabled {
		return nil
	}
	payload := slackPayload{
		Blocks: []slackBlock{
			{Type: "header", Text: &slackText{Type: "plain_text", Text: ":wastebasket: 記事を削除しました"}},
			{Type: "section", Text: &slackText{Type: "mrkdwn", Text: "*" + title + "*"}},
			{Type: "context", Elements: []slackText{
				{Type: "mrkdwn", Text: "ID: `" + articleID + "`"},
			}},
		},
	}
	return n.post(payload)
}

// NotifyAsync runs the given notification function in a goroutine.
// Errors are logged but not propagated.
func (n *SlackNotifier) NotifyAsync(fn func() error) {
	if !n.enabled {
		return
	}
	go func() {
		if err := fn(); err != nil {
			slog.Error("slack notification failed", "error", err, "webhook_url_set", n.webhookURL != "")
		}
	}()
}

func (n *SlackNotifier) post(payload slackPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack marshal error: %w", err)
	}
	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack post error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}
	return nil
}
