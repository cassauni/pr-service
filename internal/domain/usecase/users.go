package usecase

import (
	"context"
	"errors"
	"pr-service/internal/domain/entities"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

func (u *Usecase) SetUserIsActive(ctx context.Context, userID string, isActive bool) (entities.User, error) {
	user, err := u.repo.SetUserIsActive(ctx, userID, isActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.User{}, &entities.DomainError{
				Code:    entities.ErrorCodeNotFound,
				Message: "resource not found",
			}
		}
		u.log.Error("failed to set user active", zap.Error(err))
		return entities.User{}, err
	}
	return user, nil
}
