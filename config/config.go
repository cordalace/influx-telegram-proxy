package config

import (
	"errors"
	"os"
)

var (
	errMissingTelegramBotToken = errors.New("missing TELEGRAM_BOT_TOKEN env var")
	errMissingTelegramChatID   = errors.New("missing TELEGRAM_CHAT_ID env var")
)

type Config struct {
	TelegramBotToken string
	TelegramChatID   string
}

func FromEnv() (*Config, error) {
	telegramBotToken, ok := os.LookupEnv("TELEGRAM_BOT_TOKEN")
	if !ok {
		return nil, errMissingTelegramBotToken
	}

	telegramChatID, ok := os.LookupEnv("TELEGRAM_CHAT_ID")
	if !ok {
		return nil, errMissingTelegramChatID
	}

	return &Config{
		TelegramBotToken: telegramBotToken,
		TelegramChatID:   telegramChatID,
	}, nil
}
