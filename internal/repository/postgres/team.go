package postgres

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/apperror"
	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
)

type TeamRepository struct {
	pool *pgxpool.Pool
	sb   sq.StatementBuilderType
}

func NewTeamRepository(pool *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{pool: pool, sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar)}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, team entity.Team, members []entity.User) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := r.sb.
		Insert("teams").
		Columns("name", "created_at").
		Values(team.Name, team.CreatedAt)

	if err = tryExec(ctx, query, tx); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperror.ErrTeamExists
		}
		return fmt.Errorf("TeamRepository.CreateTeam failed to insert team: %w", err)
	}

	if len(members) > 0 {
		queryAddUsers := r.sb.
			Insert("users").
			Columns("id", "team_name", "name", "is_active", "created_at")

		for _, m := range members {
			queryAddUsers = queryAddUsers.Values(m.ID, team.Name, m.Name, m.IsActive, m.CreatedAt)
		}
		queryAddUsers = queryAddUsers.Suffix("ON CONFLICT (id) DO UPDATE SET " +
			"team_name = EXCLUDED.team_name, name = EXCLUDED.name, is_active = EXCLUDED.is_active")

		if err = tryExec(ctx, queryAddUsers, tx); err != nil {
			return fmt.Errorf("TeamRepository.CreateTeam failed to insert or update team members: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *TeamRepository) GetTeam(ctx context.Context, name string) (*entity.Team, []entity.User, error) {
	query := r.sb.
		Select("name", "created_at").
		From("teams").
		Where(sq.Eq{"name": name})

	row := tryQueryRow(ctx, query, r.pool)

	var team entity.Team
	if err := row.Scan(&team.Name, &team.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, apperror.ErrNotFound
		}
		return nil, nil, fmt.Errorf("TeamRepository.GetTeam failed to select team: %w", err)
	}

	queryUsers := r.sb.
		Select("id", "team_name", "name", "is_active", "created_at").
		From("users").
		Where(sq.Eq{"team_name": name})

	rows, err := tryQuery(ctx, queryUsers, r.pool)
	if err != nil {
		return &team, nil, fmt.Errorf("TeamRepository.GetTeam failed to select team members: %w", err)
	}
	defer rows.Close()

	users := make([]entity.User, 0)
	for rows.Next() {
		var u entity.User
		if err = rows.Scan(&u.ID, &u.TeamName, &u.Name, &u.IsActive, &u.CreatedAt); err != nil {
			return &team, nil, fmt.Errorf("TeamRepository.GetTeam failed to scan team member: %w", err)
		}
		users = append(users, u)
	}

	return &team, users, nil
}

func (r *TeamRepository) GetTeamForUser(ctx context.Context, userID string) (string, error) {
	query := r.sb.
		Select("team_name").
		From("users").
		Where(sq.Eq{"id": userID})

	row := tryQueryRow(ctx, query, r.pool)

	var teamName string
	if err := row.Scan(&teamName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperror.ErrNotFound
		}
		return "", err
	}

	return teamName, nil
}
