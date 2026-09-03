package main

import (
	"context"
	"cu-timepad-bot/internal/app"
	"log/slog"
	"os"
)

func main() {
	ctx := context.Background()
	if err := app.Run(ctx); err != nil {
		slog.Error(
			"Fatal error: teminating server",
			slog.Any("error", err),
		)
		os.Exit(1)
	}
}
