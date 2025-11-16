package postgres

import (
	"context"
	"pr-service/internal/domain/entities"

	"github.com/jackc/pgx/v4"
)

func (r *Repository) CreatePullRequest(ctx context.Context, pr entities.PullRequest, reviewers []string) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if _, err = tx.Exec(ctx, `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
		VALUES ($1, $2, $3, $4)
	`, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status); err != nil {
		return err
	}

	for _, rid := range reviewers {
		if _, err = tx.Exec(ctx, `
			INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
			VALUES ($1, $2)
		`, pr.PullRequestID, rid); err != nil {
			return err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetPullRequest(ctx context.Context, prID string) (entities.PullRequest, []string, error) {
	var pr entities.PullRequest
	err := r.DB.QueryRow(ctx, `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id=$1
	`, prID).
		Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		return entities.PullRequest{}, nil, err
	}

	rows, err := r.DB.Query(ctx, `
		SELECT reviewer_id
		FROM pull_request_reviewers
		WHERE pull_request_id=$1
		ORDER BY reviewer_id
	`, prID)
	if err != nil {
		return entities.PullRequest{}, nil, err
	}
	defer rows.Close()

	reviewers := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return entities.PullRequest{}, nil, err
		}
		reviewers = append(reviewers, id)
	}
	if err := rows.Err(); err != nil {
		return entities.PullRequest{}, nil, err
	}

	return pr, reviewers, nil
}

func (r *Repository) MarkPullRequestMerged(ctx context.Context, prID string) (entities.PullRequest, []string, error) {
	var pr entities.PullRequest
	err := r.DB.QueryRow(ctx, `
		UPDATE pull_requests
		SET status = 'MERGED',
		    merged_at = COALESCE(merged_at, NOW())
		WHERE pull_request_id=$1
		RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at
	`, prID).
		Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		return entities.PullRequest{}, nil, err
	}

	rows, err := r.DB.Query(ctx, `
		SELECT reviewer_id
		FROM pull_request_reviewers
		WHERE pull_request_id=$1
		ORDER BY reviewer_id
	`, prID)
	if err != nil {
		return entities.PullRequest{}, nil, err
	}
	defer rows.Close()

	reviewers := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return entities.PullRequest{}, nil, err
		}
		reviewers = append(reviewers, id)
	}
	if err := rows.Err(); err != nil {
		return entities.PullRequest{}, nil, err
	}

	return pr, reviewers, nil
}

func (r *Repository) ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	tag, err := r.DB.Exec(ctx, `
		UPDATE pull_request_reviewers
		SET reviewer_id=$3
		WHERE pull_request_id=$1 AND reviewer_id=$2
	`, prID, oldUserID, newUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) ListPullRequestsByReviewer(ctx context.Context, reviewerID string) ([]entities.PullRequestShort, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT p.pull_request_id, p.pull_request_name, p.author_id, p.status
		FROM pull_requests p
		JOIN pull_request_reviewers rpr ON rpr.pull_request_id = p.pull_request_id
		WHERE rpr.reviewer_id = $1
		ORDER BY p.created_at DESC
	`, reviewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]entities.PullRequestShort, 0)
	for rows.Next() {
		var pr entities.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		res = append(res, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
