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

type UserRepository struct {
	pool *pgxpool.Pool
	sb   sq.StatementBuilderType
	log  *log.Logger
}

func NewUserRepository(pool *pgxpool.Pool, logger *log.Logger) repo.UserRepository {
	return &UserRepository{pool: pool, sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar), log: logger}
}

func (r *UserRepository) CreateOrUpdateUser(ctx context.Context, u *entity.User) error {
	q := r.sb.
		Insert("users").
		Columns("id", "name", "is_active", "created_at").
		Values(u.ID, u.Name, u.IsActive, u.CreatedAt).
		Suffix("ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, is_active = EXCLUDED.is_active")

	sql, args, _ := q.ToSql()
	_, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		r.log.WithError(err).WithField("user_id", u.ID).Error("failed to create/update user")
	}
	return err
}

func (r *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	q := r.sb.
		Update("users").
		Set("is_active", isActive).
		Where(sq.Eq{"id": userID})

	sql, args, _ := q.ToSql()
	_, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		r.log.WithError(err).WithField("user_id", userID).Error("failed to set is_active")
	}
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	q := r.sb.
		Select("id", "name", "is_active", "created_at").
		From("users").
		Where(sq.Eq{"id": id})

	sql, args, _ := q.ToSql()
	row := r.pool.QueryRow(ctx, sql, args...)

	var u entity.User
	if err := row.Scan(&u.ID, &u.Name, &u.IsActive, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.WithField("user_id", id).Debug("user not found")
			return nil, apperror.ErrNotFound
		}
		r.log.WithError(err).WithField("user_id", id).Error("failed to scan user")
		return nil, err
	}
	return &u, nil
}
