package postgres

import (
	"context"
	"errors"
	"pr-service/internal/domain/entities"

	"github.com/jackc/pgx/v4"
)

var ErrNoReplacementCandidate = errors.New("no active replacement candidate in team")

func (r *Repository) BulkDeactivateTeamUsers(
	ctx context.Context,
	teamName string,
	userIDs []string,
) (res entities.BulkDeactivateResult, err error) {
	res.TeamName = teamName

	if len(userIDs) == 0 {
		return res, nil
	}

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return res, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	cmd, err := tx.Exec(ctx, `
		UPDATE users
		SET is_active = FALSE
		WHERE team_name = $1 AND user_id = ANY($2)
	`, teamName, userIDs)
	if err != nil {
		return res, err
	}
	res.Deactivated = int(cmd.RowsAffected())

	rows, err := tx.Query(ctx, `
		SELECT user_id
		FROM users
		WHERE team_name = $1 AND is_active = TRUE
	`, teamName)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	activeCandidates := make([]string, 0)
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return res, err
		}
		activeCandidates = append(activeCandidates, id)
	}
	if err = rows.Err(); err != nil {
		return res, err
	}
	if len(activeCandidates) == 0 {
		err = ErrNoReplacementCandidate
		return res, err
	}

	type assignment struct {
		PRID        string
		AuthorID    string
		OldReviewer string
	}

	rows, err = tx.Query(ctx, `
		SELECT pr.pull_request_id, pr.author_id, rpr.reviewer_id
		FROM pull_request_reviewers rpr
		JOIN pull_requests pr ON pr.pull_request_id = rpr.pull_request_id
		WHERE pr.status = 'OPEN'
		  AND rpr.reviewer_id = ANY($1)
		ORDER BY pr.pull_request_id
	`, userIDs)
	if err != nil {
		return res, err
	}

	assignments := make([]assignment, 0)
	prIDsSet := make(map[string]struct{})
	for rows.Next() {
		var a assignment
		if err = rows.Scan(&a.PRID, &a.AuthorID, &a.OldReviewer); err != nil {
			return res, err
		}
		assignments = append(assignments, a)
		prIDsSet[a.PRID] = struct{}{}
	}
	if err = rows.Err(); err != nil {
		return res, err
	}
	rows.Close()

	if len(assignments) == 0 {

		if err = tx.Commit(ctx); err != nil {
			return res, err
		}
		return res, nil
	}

	prIDs := make([]string, 0, len(prIDsSet))
	for id := range prIDsSet {
		prIDs = append(prIDs, id)
	}

	rows, err = tx.Query(ctx, `
		SELECT pull_request_id, reviewer_id
		FROM pull_request_reviewers
		WHERE pull_request_id = ANY($1)
	`, prIDs)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	prReviewers := make(map[string]map[string]struct{})
	for rows.Next() {
		var prID, reviewerID string
		if err = rows.Scan(&prID, &reviewerID); err != nil {
			return res, err
		}
		if _, ok := prReviewers[prID]; !ok {
			prReviewers[prID] = make(map[string]struct{})
		}
		prReviewers[prID][reviewerID] = struct{}{}
	}
	if err = rows.Err(); err != nil {
		return res, err
	}

	type replacement struct {
		PRID        string
		OldReviewer string
		NewReviewer string
	}
	replacements := make([]replacement, 0, len(assignments))

	for _, a := range assignments {
		reviewersForPR := prReviewers[a.PRID]

		var chosen string
		for _, cand := range activeCandidates {
			if cand == a.AuthorID {
				continue
			}
			if _, already := reviewersForPR[cand]; already {
				continue
			}
			chosen = cand
			break
		}

		if chosen == "" {
			err = ErrNoReplacementCandidate
			return res, err
		}

		reviewersForPR[chosen] = struct{}{}
		replacements = append(replacements, replacement{
			PRID:        a.PRID,
			OldReviewer: a.OldReviewer,
			NewReviewer: chosen,
		})
	}

	for _, rep := range replacements {
		tag, execErr := tx.Exec(ctx, `
			UPDATE pull_request_reviewers
			SET reviewer_id = $3
			WHERE pull_request_id = $1 AND reviewer_id = $2
		`, rep.PRID, rep.OldReviewer, rep.NewReviewer)
		if execErr != nil {
			err = execErr
			return res, err
		}
		if tag.RowsAffected() == 0 {
			err = pgx.ErrNoRows
			return res, err
		}
		res.ReassignedCount += int(tag.RowsAffected())
	}

	if err = tx.Commit(ctx); err != nil {
		return res, err
	}
	return res, nil
}
