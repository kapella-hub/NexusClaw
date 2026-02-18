package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/kapella-hub/NexusClaw/internal/app"
	"github.com/kapella-hub/NexusClaw/internal/platform/config"
	"github.com/kapella-hub/NexusClaw/internal/platform/database"
	"github.com/kapella-hub/NexusClaw/internal/platform/server"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the NexusClaw HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Setup logger
		var level slog.Level
		switch cfg.Log.Level {
		case "debug":
			level = slog.LevelDebug
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		default:
			level = slog.LevelInfo
		}
		logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
		slog.SetDefault(logger)

		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		// Initialize database pool
		pool, err := database.NewPostgres(ctx, cfg.Database.DSN)
		if err != nil {
			return fmt.Errorf("connecting to database: %w", err)
		}
		defer pool.Close()
		slog.Info("database connected", "dsn", cfg.Database.DSN)

		handler := app.New(cfg, logger, pool)

		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		slog.Info("starting server", "addr", addr)

		return server.Run(ctx, handler, addr)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntP("port", "p", 8080, "port to listen on")
}
