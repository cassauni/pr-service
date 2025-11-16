package usecase

import (
	"context"
	"errors"
	"math/rand"
	"pr-service/internal/domain/entities"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

func (u *Usecase) CreatePullRequest(ctx context.Context, req entities.CreatePullRequestRequest) (entities.PullRequest, error) {
	author, err := u.repo.GetUserByID(ctx, req.AuthorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.PullRequest{}, &entities.DomainError{
				Code:    entities.ErrorCodeNotFound,
				Message: "resource not found",
			}
		}
		u.log.Error("failed to get author", zap.Error(err))
		return entities.PullRequest{}, err
	}

	candidates, err := u.repo.ListTeamActiveUsersExcept(ctx, author.TeamName, author.UserID)
	if err != nil {
		u.log.Error("failed to list reviewer candidates", zap.Error(err))
		return entities.PullRequest{}, err
	}

	reviewers := pickReviewers(candidates, 2)

	pr := entities.PullRequest{
		PullRequestID:   req.PullRequestID,
		PullRequestName: req.PullRequestName,
		AuthorID:        req.AuthorID,
		Status:          "OPEN",
	}

	if err := u.repo.CreatePullRequest(ctx, pr, reviewers); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entities.PullRequest{}, &entities.DomainError{
				Code:    entities.ErrorCodePRExists,
				Message: "PR id already exists",
			}
		}
		u.log.Error("failed to create pull request", zap.Error(err))
		return entities.PullRequest{}, err
	}

	created, assigned, err := u.repo.GetPullRequest(ctx, req.PullRequestID)
	if err != nil {
		u.log.Error("failed to reload created pull request", zap.Error(err))
		return entities.PullRequest{}, err
	}
	created.AssignedReviewers = assigned

	return created, nil
}

func (u *Usecase) MergePullRequest(ctx context.Context, prID string) (entities.PullRequest, error) {
	pr, reviewers, err := u.repo.MarkPullRequestMerged(ctx, prID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.PullRequest{}, &entities.DomainError{
				Code:    entities.ErrorCodeNotFound,
				Message: "resource not found",
			}
		}
		u.log.Error("failed to merge pull request", zap.Error(err))
		return entities.PullRequest{}, err
	}
	pr.AssignedReviewers = reviewers
	return pr, nil
}

func (u *Usecase) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (entities.PullRequest, string, error) {
	pr, reviewers, err := u.repo.GetPullRequest(ctx, prID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.PullRequest{}, "", &entities.DomainError{
				Code:    entities.ErrorCodeNotFound,
				Message: "resource not found",
			}
		}
		u.log.Error("failed to get pull request", zap.Error(err))
		return entities.PullRequest{}, "", err
	}

	if pr.Status == "MERGED" {

		return entities.PullRequest{}, "", &entities.DomainError{
			Code:    entities.ErrorCodePRMerged,
			Message: "cannot reassign on merged PR",
		}
	}

	assignedSet := make(map[string]struct{}, len(reviewers))
	for _, id := range reviewers {
		assignedSet[id] = struct{}{}
	}
	if _, ok := assignedSet[oldReviewerID]; !ok {
		return entities.PullRequest{}, "", &entities.DomainError{
			Code:    entities.ErrorCodeNotAssigned,
			Message: "reviewer is not assigned to this PR",
		}
	}

	oldUser, err := u.repo.GetUserByID(ctx, oldReviewerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.PullRequest{}, "", &entities.DomainError{
				Code:    entities.ErrorCodeNotFound,
				Message: "resource not found",
			}
		}
		u.log.Error("failed to get old reviewer", zap.Error(err))
		return entities.PullRequest{}, "", err
	}

	candidates, err := u.repo.ListTeamActiveUsersExcept(ctx, oldUser.TeamName, oldReviewerID)
	if err != nil {
		u.log.Error("failed to list replacement candidates", zap.Error(err))
		return entities.PullRequest{}, "", err
	}

	filtered := make([]entities.User, 0, len(candidates))
	for _, cand := range candidates {
		if cand.UserID == pr.AuthorID {
			continue
		}
		if _, already := assignedSet[cand.UserID]; already {
			continue
		}
		filtered = append(filtered, cand)
	}

	if len(filtered) == 0 {
		return entities.PullRequest{}, "", &entities.DomainError{
			Code:    entities.ErrorCodeNoCandidate,
			Message: "no active replacement candidate in team",
		}
	}

	newReviewer := filtered[rand.Intn(len(filtered))].UserID

	if err := u.repo.ReplaceReviewer(ctx, prID, oldReviewerID, newReviewer); err != nil {
		u.log.Error("failed to replace reviewer", zap.Error(err))
		return entities.PullRequest{}, "", err
	}

	updated, updatedReviewers, err := u.repo.GetPullRequest(ctx, prID)
	if err != nil {
		u.log.Error("failed to reload pull request after reassign", zap.Error(err))
		return entities.PullRequest{}, "", err
	}
	updated.AssignedReviewers = updatedReviewers

	return updated, newReviewer, nil
}

func (u *Usecase) GetUserReviews(ctx context.Context, userID string) (entities.GetUserReviewsResponse, error) {
	if _, err := u.repo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.GetUserReviewsResponse{}, &entities.DomainError{
				Code:    entities.ErrorCodeNotFound,
				Message: "resource not found",
			}
		}
		u.log.Error("failed to get user", zap.Error(err))
		return entities.GetUserReviewsResponse{}, err
	}

	prs, err := u.repo.ListPullRequestsByReviewer(ctx, userID)
	if err != nil {
		u.log.Error("failed to list user PRs", zap.Error(err))
		return entities.GetUserReviewsResponse{}, err
	}

	return entities.GetUserReviewsResponse{
		UserID:       userID,
		PullRequests: prs,
	}, nil
}
