package main

import (
	"context"
	"cu-timepad-bot/internal/botsetup"
	"cu-timepad-bot/internal/config"
	"cu-timepad-bot/internal/handler"
	"cu-timepad-bot/internal/service"
	memcachestore "cu-timepad-bot/internal/store/memcache"
	mongostore "cu-timepad-bot/internal/store/mongo"
	"cu-timepad-bot/pkg/timepad"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var logLevel = &slog.LevelVar{}

const shutdownTimeout = 30 * time.Second

func main() {
	if err := run(); err != nil {
		slog.Error(
			"Fatal error: teminating server",
			slog.Any("error", err),
		)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	initLogger()

	cfg, err := initConfig(ctx)
	if err != nil {
		return err
	}
	ctx = config.InjectConfig(ctx, cfg)

	mongo_client, err := initMongoClient(ctx)
	if err != nil {
		return err
	}

	st_mongo, err := mongostore.New(ctx, mongo_client)
	if err != nil {
		return err
	}
	st := memcachestore.New(ctx, st_mongo)

	svc := service.New(st)
	h := handler.New(svc)

	b, err := botsetup.Handle(ctx, h)
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			"Can't start bot",
			slog.Any("error", err.Error()),
		)
		return err
	}

	if cfg.WebhookURL == "" {
		go b.Start(ctx)
	} else {
		go b.StartWebhook(ctx)
	}

	setupWorkers(ctx, svc, b)

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

	ctx, cancel = context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	b.Close(ctx)
	mongo_client.Disconnect(ctx)
	return nil
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

func initLogger() {
	logLevel.Set(slog.LevelWarn)
	log_opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, log_opts))
	slog.SetDefault(logger)
}

func initConfig(ctx context.Context) (*config.Config, error) {
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
		return nil, err
	}
	logLevel.Set(getSlogLevel(cfg.LogLevel))
	slog.LogAttrs(ctx,
		slog.LevelDebug,
		"Got config",
		slog.String("config", fmt.Sprintf("%+v", cfg)),
	)

	return &cfg, err
}

func initMongoClient(ctx context.Context) (*mongo.Client, error) {
	cfg := config.GetConfig(ctx)

	mongo_opts := options.Client().ApplyURI(cfg.MongoURI)
	mongo_client, err := mongo.Connect(mongo_opts)
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			"Can't connect to mongodb",
			slog.Any("error", err.Error()),
			slog.String("mongodb_uri", cfg.MongoURI),
		)
		return nil, err
	}

	ctx_ping, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = mongo_client.Ping(ctx_ping, nil)
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			"MongoDB ping failed",
			slog.Any("error", err),
		)
		return nil, err
	}
	return mongo_client, nil
}

func setupWorkers(ctx context.Context, svc *service.Service, b *bot.Bot) {
	cfg := config.GetConfig(ctx)

	http_client := &http.Client{
		Timeout: time.Duration(cfg.TimepadTimeout) * time.Second,
	}
	timepad_client := timepad.Client{
		HTTPClient: http_client,
	}

	go svc.StartTimepadWorker(ctx, &timepad_client)
	go svc.StartNotifyingWorker(ctx, b)
}
