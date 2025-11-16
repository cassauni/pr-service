package postgres

import (
	"context"
	"pr-service/internal/domain/entities"
)

func (r *Repository) TeamExists(ctx context.Context, teamName string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name=$1)`, teamName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) CreateTeam(ctx context.Context, team entities.Team) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if _, err = tx.Exec(ctx,
		`INSERT INTO teams (team_name) VALUES ($1)`,
		team.TeamName,
	); err != nil {
		return err
	}

	for _, m := range team.Members {
		if _, err = tx.Exec(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) DO UPDATE
			SET username = EXCLUDED.username,
			    team_name = EXCLUDED.team_name,
			    is_active = EXCLUDED.is_active
		`, m.UserID, m.Username, team.TeamName, m.IsActive); err != nil {
			return err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetTeam(ctx context.Context, teamName string) (entities.Team, error) {
	var name string
	if err := r.DB.QueryRow(ctx,
		`SELECT team_name FROM teams WHERE team_name=$1`,
		teamName,
	).Scan(&name); err != nil {
		return entities.Team{}, err
	}

	rows, err := r.DB.Query(ctx, `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_name=$1
		ORDER BY user_id
	`, teamName)
	if err != nil {
		return entities.Team{}, err
	}
	defer rows.Close()

	members := make([]entities.TeamMember, 0)
	for rows.Next() {
		var m entities.TeamMember
		if err := rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return entities.Team{}, err
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return entities.Team{}, err
	}

	return entities.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}
