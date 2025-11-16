package postgres

import (
	"context"
	"pr-service/internal/domain/entities"
)

func (r *Repository) GetAssignmentsStats(ctx context.Context) ([]entities.ReviewerAssignmentsStat, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT reviewer_id AS user_id, COUNT(*) AS assignments
		FROM pull_request_reviewers
		GROUP BY reviewer_id
		ORDER BY reviewer_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]entities.ReviewerAssignmentsStat, 0)
	for rows.Next() {
		var s entities.ReviewerAssignmentsStat
		if err := rows.Scan(&s.UserID, &s.Assignments); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}
