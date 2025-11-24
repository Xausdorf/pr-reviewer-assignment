package postgres

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/apperror"
	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
)

type UserRepository struct {
	pool *pgxpool.Pool
	sb   sq.StatementBuilderType
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool, sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar)}
}

func (r *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*entity.User, error) {
	query := r.sb.
		Update("users").
		Set("is_active", isActive).
		Where(sq.Eq{"id": userID}).
		Suffix("RETURNING id, name, team_name, is_active, created_at")

	row := tryQueryRow(query, r.pool, ctx)

	var user entity.User
	if err := row.Scan(&user.ID, &user.Name, &user.TeamName, &user.IsActive, &user.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, fmt.Errorf("UserRepository.SetIsActive failed to update user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) ListAssignedTo(ctx context.Context, userID string) ([]entity.PR, error) {
	query := r.sb.
		Select("p.id", "p.title", "p.author_id", "p.status", "p.created_at", "p.merged_at").
		From("prs p").
		Join("pr_reviewers r ON p.id = r.pr_id").
		Where(sq.Eq{"r.reviewer_id": userID})

	rows, err := tryQuery(query, r.pool, ctx)
	if err != nil {
		return nil, fmt.Errorf("UserRepository.ListAssignedTo failed to select assigned PRs: %w", err)
	}
	defer rows.Close()

	prs := make([]entity.PR, 0)
	for rows.Next() {
		var pr entity.PR
		if err := rows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return nil, fmt.Errorf("UserRepository.ListAssignedTo failed to scan assigned PR: %w", err)
		}
		prs = append(prs, pr)
	}

	return prs, nil
}

func (r *UserRepository) IsAssignedToPR(ctx context.Context, userID, prID string) (bool, error) {
	query := sq.Select("1").
		Prefix("SELECT EXISTS (").
		From("pr_reviewers").
		Where(sq.Eq{"pr_id": prID, "reviewer_id": userID}).
		Suffix(")")

	row := tryQueryRow(query, r.pool, ctx)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("UserRepository.IsAssignedToPR failed to check assignment: %w", err)
	}

	return exists, nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*entity.User, error) {
	query := r.sb.
		Select("id", "team_name", "name", "is_active", "created_at").
		From("users").
		Where(sq.Eq{"id": userID})

	row := tryQueryRow(query, r.pool, ctx)

	var user entity.User
	if err := row.Scan(&user.ID, &user.TeamName, &user.Name, &user.IsActive, &user.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, fmt.Errorf("UserRepository.GetByID failed to select user: %w", err)
	}

	return &user, nil
}
