package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cordalace/influx-telegram-proxy/notification"
	"go.uber.org/zap"
)

var ErrTelegramBadStatus = errors.New("telegram API bad status")

type Telegram struct {
	httpClient *http.Client
	baseURL    string
	chatID     string
	logger     *zap.Logger
}

func NewTelegram(httpClient *http.Client, botToken string, chatID string, logger *zap.Logger) *Telegram {
	return &Telegram{
		httpClient: httpClient,
		baseURL:    fmt.Sprintf("https://api.telegram.org/bot%v", botToken),
		chatID:     chatID,
		logger:     logger,
	}
}

func (t *Telegram) SendNotificationMessage(ctx context.Context, n *notification.Notification) error {
	telegramMessageText := t.formatTelegramMessageText(n)

	j := &telegramRequestBodyJSON{
		ChatID:    t.chatID,
		Text:      telegramMessageText,
		ParseMode: "MarkdownV2",
	}
	bodyBytes, err := json.Marshal(j)
	if err != nil {
		return fmt.Errorf("error marshaling telegram request json: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL+"/sendMessage", bytes.NewBuffer(bodyBytes))
	if err != nil {
		t.logger.Error("error creating http request instance", zap.Error(err))
		return fmt.Errorf("error creating http request instance: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		t.logger.Error("error making telegram HTTP request", zap.Error(err))
		return fmt.Errorf("error making telegram HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		var untypedResponse interface{}
		if err := json.NewDecoder(resp.Body).Decode(&untypedResponse); err != nil {
			t.logger.Error("error decoding bad response body", zap.Int("statusCode", resp.StatusCode), zap.Error(err))
		}
		t.logger.Error(
			"bad telegram bot api status code",
			zap.Int("statusCode", resp.StatusCode),
			zap.Any("body", untypedResponse),
		)
		return ErrTelegramBadStatus
	}

	return nil
}

func (t *Telegram) formatTelegramMessageText(n *notification.Notification) string {
	return fmt.Sprintf(
		"%v on %v at %v\n\n%v",
		markdownBold(strings.ToUpper(n.Level)),
		markdownSafe(n.CheckName),
		markdownSafe(n.Time.Format(time.RFC3339)),
		markdownSafe(n.Message),
	)
}

// According to official telegram bot API docs:
//
// > Any character with code between 1 and 126 inclusively can be escaped
// > anywhere with a preceding '\\' character, in which case it is treated as
// > an ordinary character and not a part of the markup.
//
// See: https://core.telegram.org/bots/api#markdownv2-style
func markdownSafe(text string) string {
	safe := make([]rune, 0, len(text)+len(text))
	for _, r := range text {
		if r >= 1 && r <= 126 {
			safe = append(safe, '\\', r)
		} else {
			safe = append(safe, r)
		}
	}
	return string(safe)
}

func markdownBold(text string) string {
	return "*" + markdownSafe(text) + "*"
}

type telegramRequestBodyJSON struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}
