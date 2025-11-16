package postgres

import (
	"context"
	"pr-service/internal/domain/entities"
)

func (r *Repository) SetUserIsActive(ctx context.Context, userID string, isActive bool) (entities.User, error) {
	var u entities.User
	err := r.DB.QueryRow(ctx, `
		UPDATE users
		SET is_active=$2
		WHERE user_id=$1
		RETURNING user_id, username, team_name, is_active
	`, userID, isActive).
		Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		return entities.User{}, err
	}
	return u, nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID string) (entities.User, error) {
	var u entities.User
	err := r.DB.QueryRow(ctx, `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id=$1
	`, userID).
		Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		return entities.User{}, err
	}
	return u, nil
}

func (r *Repository) ListTeamActiveUsersExcept(ctx context.Context, teamName, exceptUserID string) ([]entities.User, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name=$1 AND is_active=TRUE AND user_id <> $2
	`, teamName, exceptUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]entities.User, 0)
	for rows.Next() {
		var u entities.User
		if err := rows.Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
