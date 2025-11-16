package usecase

import (
	"context"
	"pr-service/internal/domain/entities"

	"go.uber.org/zap"
)

func (u *Usecase) GetAssignmentsStats(ctx context.Context) (entities.AssignmentsStatsResponse, error) {
	stats, err := u.repo.GetAssignmentsStats(ctx)
	if err != nil {
		u.log.Error("failed to get assignments stats", zap.Error(err))
		return entities.AssignmentsStatsResponse{}, err
	}

	return entities.AssignmentsStatsResponse{
		Reviewers: stats,
	}, nil
}
