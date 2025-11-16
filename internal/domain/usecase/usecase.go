package usecase

import (
	"context"
	"math/rand"
	"pr-service/config"
	"pr-service/internal/domain/entities"
	"pr-service/internal/domain/repository/postgres"

	"go.uber.org/zap"
)

type repository interface {
	TeamExists(ctx context.Context, teamName string) (bool, error)
	CreateTeam(ctx context.Context, team entities.Team) error
	GetTeam(ctx context.Context, teamName string) (entities.Team, error)

	SetUserIsActive(ctx context.Context, userID string, isActive bool) (entities.User, error)
	GetUserByID(ctx context.Context, userID string) (entities.User, error)
	ListTeamActiveUsersExcept(ctx context.Context, teamName, exceptUserID string) ([]entities.User, error)

	CreatePullRequest(ctx context.Context, pr entities.PullRequest, reviewers []string) error
	GetPullRequest(ctx context.Context, prID string) (entities.PullRequest, []string, error)
	MarkPullRequestMerged(ctx context.Context, prID string) (entities.PullRequest, []string, error)
	ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error
	ListPullRequestsByReviewer(ctx context.Context, reviewerID string) ([]entities.PullRequestShort, error)

	GetAssignmentsStats(ctx context.Context) ([]entities.ReviewerAssignmentsStat, error)

	BulkDeactivateTeamUsers(ctx context.Context, teamName string, userIDs []string) (entities.BulkDeactivateResult, error)
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
	return &Usecase{cfg: cfg, log: log, repo: repo}, nil
}

func pickReviewers(users []entities.User, limit int) []string {
	if limit <= 0 || len(users) == 0 {
		return nil
	}

	if len(users) <= limit {
		res := make([]string, 0, len(users))
		for _, u := range users {
			res = append(res, u.UserID)
		}
		return res
	}

	idxs := make([]int, len(users))
	for i := range users {
		idxs[i] = i
	}

	rand.Shuffle(len(idxs), func(i, j int) {
		idxs[i], idxs[j] = idxs[j], idxs[i]
	})

	res := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		res = append(res, users[idxs[i]].UserID)
	}

	return res
}
