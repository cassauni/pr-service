package usecase

import (
	"context"
	"errors"
	"pr-service/internal/domain/entities"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

func (u *Usecase) CreateTeam(ctx context.Context, team entities.Team) (entities.Team, error) {
	exists, err := u.repo.TeamExists(ctx, team.TeamName)
	if err != nil {
		u.log.Error("failed to check team exists", zap.Error(err))
		return entities.Team{}, err
	}
	if exists {
		return entities.Team{}, &entities.DomainError{
			Code:    entities.ErrorCodeTeamExists,
			Message: "team_name already exists",
		}
	}

	if err := u.repo.CreateTeam(ctx, team); err != nil {
		u.log.Error("failed to create team", zap.Error(err))
		return entities.Team{}, err
	}

	created, err := u.repo.GetTeam(ctx, team.TeamName)
	if err != nil {
		u.log.Error("failed to reload created team", zap.Error(err))
		return entities.Team{}, err
	}
	return created, nil
}

func (u *Usecase) GetTeam(ctx context.Context, teamName string) (entities.Team, error) {
	team, err := u.repo.GetTeam(ctx, teamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.Team{}, &entities.DomainError{
				Code:    entities.ErrorCodeNotFound,
				Message: "resource not found",
			}
		}
		u.log.Error("failed to get team", zap.Error(err))
		return entities.Team{}, err
	}
	return team, nil
}
