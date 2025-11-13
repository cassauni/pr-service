package app

import (
	"context"
	"pr-service/config"
	"pr-service/internal/domain/delivery/http"
	"pr-service/internal/domain/repository"
	"pr-service/internal/domain/usecase"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func New() *fx.App {
	return fx.New(
		fx.Options(
			repository.New(),
			usecase.New(),
			http.New(),
		),
		fx.Provide(
			context.Background,
			config.NewConfig,
			zap.NewDevelopment,
		),
		fx.WithLogger(
			func(log *zap.Logger) fxevent.Logger {
				return &fxevent.ZapLogger{Logger: log}
			},
		),
	)
}
