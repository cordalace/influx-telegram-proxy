package main

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/cordalace/influx-telegram-proxy/config"
	"github.com/cordalace/influx-telegram-proxy/proxy"
	"github.com/cordalace/influx-telegram-proxy/telegram"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	logConfig := zap.NewProductionConfig()
	logConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	logger, err := logConfig.Build()
	if err != nil {
		panic(err)
	}

	cfg, err := config.FromEnv()
	if err != nil {
		logger.Fatal("error parsing environment", zap.Error(err))
	}

	httpClient := &http.Client{}

	tg := telegram.NewTelegram(httpClient, cfg.TelegramBotToken, cfg.TelegramChatID, logger)

	proxyApp := proxy.NewApplication(tg, logger)

	router := chi.NewRouter()
	proxyApp.Route(router)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// run the http server in a separate goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			logger.Error("error running http server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("start graceful shutdown")

	if err := server.Close(); err != nil {
		logger.Error("error closing http server", zap.Error(err))
	}

	logger.Info("exiting")
}
