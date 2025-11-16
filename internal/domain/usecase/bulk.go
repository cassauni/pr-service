package usecase

import (
	"context"
	"errors"
	"pr-service/internal/domain/entities"
	"pr-service/internal/domain/repository/postgres"

	"go.uber.org/zap"
)

func (u *Usecase) BulkDeactivateTeamUsers(
	ctx context.Context,
	teamName string,
	userIDs []string,
) (entities.BulkDeactivateResult, error) {
	exists, err := u.repo.TeamExists(ctx, teamName)
	if err != nil {
		u.log.Error("failed to check team exists before bulk deactivate", zap.Error(err))
		return entities.BulkDeactivateResult{}, err
	}
	if !exists {
		return entities.BulkDeactivateResult{}, &entities.DomainError{
			Code:    entities.ErrorCodeNotFound,
			Message: "resource not found",
		}
	}

	res, err := u.repo.BulkDeactivateTeamUsers(ctx, teamName, userIDs)
	if err != nil {
		if errors.Is(err, postgres.ErrNoReplacementCandidate) {
			return entities.BulkDeactivateResult{}, &entities.DomainError{
				Code:    entities.ErrorCodeNoCandidate,
				Message: "no active replacement candidate in team",
			}
		}
		u.log.Error("failed to bulk deactivate users", zap.Error(err))
		return entities.BulkDeactivateResult{}, err
	}

	return res, nil
}
