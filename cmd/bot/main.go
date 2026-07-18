package main

import (
	"context"
	"cu-timepad-bot/internal/botsetup"
	"cu-timepad-bot/internal/config"
	"cu-timepad-bot/internal/handler"
	"cu-timepad-bot/internal/store"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	ctx := context.Background()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	logLevel := &slog.LevelVar{}
	logLevel.Set(slog.LevelWarn)
	log_opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, log_opts))
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		slog.LogAttrs(ctx,
			slog.LevelWarn,
			".env not found",
			slog.String("error", err.Error()),
		)
	}
	cfg, err := config.Load()
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			"Can't load config",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	logLevel.Set(getSlogLevel(cfg.LogLevel))
	slog.LogAttrs(ctx,
		slog.LevelDebug,
		"Got config",
		slog.String("config", fmt.Sprintf("%+v", cfg)),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = config.InjectConfig(ctx, cfg)

	mongo_opts := options.Client().ApplyURI(cfg.MongoURI).SetTimeout(1 * time.Second)
	client, err := mongo.Connect(mongo_opts)
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			"Can't connect to mongodb",
			slog.Any("error", err.Error()),
			slog.String("mongodb_uri", cfg.MongoURI),
		)
		os.Exit(1)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			"Can't connect to mongodb",
			slog.Any("error", err.Error()),
			slog.String("mongodb_uri", cfg.MongoURI),
		)
		os.Exit(1)
	}
	st := store.New(ctx, client)
	h := handler.New(st)

	b, err := botsetup.Handle(ctx, h)

	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			"Can't start bot",
			slog.Any("error", err.Error()),
		)
		os.Exit(1)
	}

	if cfg.WebhookURL == "" {
		go b.Start(ctx)
	} else {
		go b.StartWebhook(ctx)
	}

	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		"Server started",
	)

	signal := <-signals
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		"Got signal",
		slog.String("signal", signal.String()),
	)

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	b.Close(ctx)
	client.Disconnect(ctx)
}

func getSlogLevel(str string) slog.Level {
	str = strings.ToLower(str)
	switch str {
	case "error":
		return slog.LevelError
	case "warn", "warning":
		return slog.LevelWarn
	case "info":
		return slog.LevelInfo
	case "debug":
		return slog.LevelDebug
	}
	return slog.LevelWarn
}
