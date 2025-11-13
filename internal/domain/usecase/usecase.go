package usecase

import (
	"pr-service/config"
	"pr-service/internal/domain/repository/postgres"

	"go.uber.org/zap"
)

type repository interface {
}

type Usecase struct {
	cfg  *config.ConfigModel
	log  *zap.Logger
	repo repository
}

func NewUsecase(
	log *zap.Logger,
	repo *postgres.Repository,
	cfg *config.ConfigModel,
) (*Usecase, error) {
	return newUsecase(log, repo, cfg), nil
}

func newUsecase(
	log *zap.Logger,
	repo repository,
	cfg *config.ConfigModel,
) *Usecase {
	return &Usecase{cfg: cfg, log: log, repo: repo}
}
