package http

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"http",
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("http")
	}),
	fx.Provide(New),
	fx.Invoke(func(lc fx.Lifecycle, cfg Config, app *fiber.App, logger *zap.Logger) {
		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				go func() {
					if err := app.Listen(cfg.Address); err != nil {
						logger.Error("server failed to start", zap.Error(err))
					}
				}()
				logger.Info("server starting", zap.String("address", cfg.Address))

				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Warn("shutting down server")
				if err := app.ShutdownWithContext(ctx); err != nil {
					logger.Error("server shutdown failed", zap.Error(err))
					return fmt.Errorf("server shutdown failed: %w", err)
				}
				logger.Info("server shutdown completed")
				return nil
			},
		})
	}),
)
