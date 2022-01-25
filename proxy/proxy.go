package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cordalace/influx-telegram-proxy/notification"
	"github.com/cordalace/influx-telegram-proxy/telegram"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Application struct {
	telegram *telegram.Telegram
	logger   *zap.Logger
}

func NewApplication(telegram *telegram.Telegram, logger *zap.Logger) *Application {
	return &Application{telegram: telegram, logger: logger}
}

func (a *Application) Route(router *chi.Mux) {
	router.Post("/", a.handler)
}

func (a *Application) handler(w http.ResponseWriter, r *http.Request) {
	n, err := a.parseNotification(r)
	if err != nil {
		a.writeError(w, http.StatusBadRequest, "request parse error")
		return
	}

	a.logger.Info("received notification", zap.Any("notification", n))

	if err := a.telegram.SendNotificationMessage(r.Context(), n); err != nil {
		a.logger.Info("telegram send error", zap.Error(err))
		a.writeError(w, http.StatusInternalServerError, "telegram send error")
		return
	}

	a.writeSuccess(w, http.StatusOK, "sent to telegram")
}

type notificationJSON struct {
	CheckID   string `json:"_check_id"`
	CheckName string `json:"_check_name"`
	Level     string `json:"_level"`
	Message   string `json:"_message"`
	Time      string `json:"_time"`
	Type      string `json:"_type"`
}

func (j *notificationJSON) parse() (*notification.Notification, error) {
	t, err := time.Parse(time.RFC3339, j.Time)
	if err != nil {
		return nil, fmt.Errorf("error parsing RFC3339 time: %w", err)
	}

	return &notification.Notification{
		CheckID:   j.CheckID,
		CheckName: j.CheckName,
		Level:     j.Level,
		Message:   j.Message,
		Time:      t,
		Type:      j.Type,
	}, nil
}

func (a *Application) parseNotification(r *http.Request) (*notification.Notification, error) {
	j := &notificationJSON{}
	if err := json.NewDecoder(r.Body).Decode(j); err != nil {
		a.logger.Error("error unmarshaling request json", zap.Error(err))
		return nil, fmt.Errorf("error unmarshaling request json: %w", err)
	}

	return j.parse()
}

type errorResponseJSON struct {
	Error string `json:"error"`
}

func (a *Application) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(errorResponseJSON{Error: message}); err != nil {
		a.logger.Error("error writing json response", zap.Error(err))
	}
}

type successResponseJSON struct {
	Message string `json:"message"`
}

func (a *Application) writeSuccess(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(successResponseJSON{Message: message}); err != nil {
		a.logger.Error("error writing json response", zap.Error(err))
	}
}
