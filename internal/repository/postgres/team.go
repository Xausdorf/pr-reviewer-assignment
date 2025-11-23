package postgres

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/apperror"
	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
	repo "github.com/Xausdorf/pr-reviewer-assignment/internal/repository"
)

type TeamRepository struct {
	pool *pgxpool.Pool
	sb   sq.StatementBuilderType
	log  *log.Logger
}

func NewTeamRepository(pool *pgxpool.Pool, logger *log.Logger) repo.TeamRepository {
	return &TeamRepository{pool: pool, sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar), log: logger}
}

func (r *TeamRepository) CreateOrUpdateTeam(ctx context.Context, team *entity.Team, members []entity.TeamMember) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q1 := r.sb.Insert("teams").Columns("name", "created_at").Values(team.Name, team.CreatedAt).Suffix("ON CONFLICT (name) DO NOTHING")
	sql1, args1, _ := q1.ToSql()
	if _, err := tx.Exec(ctx, sql1, args1...); err != nil {
		return err
	}

	q2 := r.sb.Delete("team_members").Where(sq.Eq{"team_name": team.Name})
	sql2, args2, _ := q2.ToSql()
	if _, err := tx.Exec(ctx, sql2, args2...); err != nil {
		return err
	}

	for _, m := range members {
		q := r.sb.Insert("team_members").Columns("team_name", "user_id").Values(m.TeamName, m.UserID)
		sql, args, _ := q.ToSql()
		if _, err := tx.Exec(ctx, sql, args...); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *TeamRepository) GetTeam(ctx context.Context, name string) (*entity.Team, []entity.TeamMember, error) {
	q := r.sb.Select("name", "created_at").From("teams").Where(sq.Eq{"name": name})
	sql, args, _ := q.ToSql()
	row := r.pool.QueryRow(ctx, sql, args...)
	var team entity.Team
	if err := row.Scan(&team.Name, &team.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.WithField("team", name).Debug("team not found")
			return nil, nil, apperror.ErrNotFound
		}
		r.log.WithError(err).WithField("team", name).Error("failed to scan team")
		return nil, nil, err
	}

	q2 := r.sb.Select("team_name", "user_id").From("team_members").Where(sq.Eq{"team_name": name})
	sql2, args2, _ := q2.ToSql()
	rows, err := r.pool.Query(ctx, sql2, args2...)
	if err != nil {
		r.log.WithError(err).WithField("team", name).Error("failed to query team members")
		return &team, nil, err
	}
	defer rows.Close()
	members := make([]entity.TeamMember, 0)
	for rows.Next() {
		var m entity.TeamMember
		if err := rows.Scan(&m.TeamName, &m.UserID); err != nil {
			return &team, nil, err
		}
		members = append(members, m)
	}
	return &team, members, nil
}

func (r *TeamRepository) GetTeamsForUser(ctx context.Context, userID string) ([]string, error) {
	q := r.sb.Select("team_name").From("team_members").Where(sq.Eq{"user_id": userID})
	sql, args, _ := q.ToSql()
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		r.log.WithError(err).WithField("user_id", userID).Error("failed to query teams for user")
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var tn string
		if err := rows.Scan(&tn); err != nil {
			r.log.WithError(err).Error("failed to scan team_name")
			return nil, err
		}
		out = append(out, tn)
	}
	return out, nil
}
