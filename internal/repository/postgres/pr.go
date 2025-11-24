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

type PRRepository struct {
	pool *pgxpool.Pool
	sb   sq.StatementBuilderType
}

func NewPRRepository(pool *pgxpool.Pool) *PRRepository {
	return &PRRepository{pool: pool, sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar)}
}

func (r *PRRepository) Create(ctx context.Context, pr entity.PR) error {
	query := r.sb.
		Insert("prs").
		Columns("id", "title", "author_id", "status", "created_at").
		Values(pr.ID, pr.Title, pr.AuthorID, pr.Status, pr.CreatedAt)

	if _, err := tryExec(query, r.pool, ctx); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperror.ErrPRExists
		}
		return fmt.Errorf("PRRepository.Create failed to insert pr: %w", err)
	}

	return nil
}

func (r *PRRepository) GetByID(ctx context.Context, id string) (*entity.PR, error) {
	query := r.sb.
		Select("id", "title", "author_id", "status", "created_at", "merged_at").
		From("prs").
		Where(sq.Eq{"id": id})

	row := tryQueryRow(query, r.pool, ctx)

	var pr entity.PR
	if err := row.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, fmt.Errorf("PRRepository.GetByID failed to select pr: %w", err)
	}

	return &pr, nil
}

func (r *PRRepository) UpdateStatus(ctx context.Context, id, status string) (*entity.PR, error) {
	query := r.sb.
		Update("prs").
		Set("status", status).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING id, title, author_id, status, created_at, merged_at")

	row := tryQueryRow(query, r.pool, ctx)

	var pr entity.PR
	if err := row.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, fmt.Errorf("PRRepository.UpdateStatus failed to update pr status: %w", err)
	}

	return &pr, nil
}

func (r *PRRepository) AssignReviewer(ctx context.Context, prID, teamName string) (reviewerID string, err error) {
	querySelect := r.sb.
		Select("id").
		From("users").
		Where(sq.Eq{"team_name": teamName, "is_active": true}).
		Where("id NOT IN (SELECT reviewer_id FROM pr_reviewers WHERE pr_id = ?)", prID).
		Where("id NOT IN (SELECT author_id FROM prs WHERE id = ?)", prID).
		OrderBy("RANDOM()").
		Limit(1)

	row := tryQueryRow(querySelect, r.pool, ctx)

	if err := row.Scan(&reviewerID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperror.ErrNoCandidate
		}
		return "", err
	}

	queryInsert := r.sb.
		Insert("pr_reviewers").
		Columns("pr_id", "reviewer_id").
		Values(prID, reviewerID)

	if _, err := tryExec(queryInsert, r.pool, ctx); err != nil {
		return "", fmt.Errorf("PRRepository.AssignReviewer failed to insert pr_reviewer: %w", err)
	}

	return reviewerID, nil
}

func (r *PRRepository) RemoveReviewer(ctx context.Context, prID, reviewerID string) error {
	query := r.sb.
		Delete("pr_reviewers").
		Where(sq.Eq{"pr_id": prID, "reviewer_id": reviewerID})

	if _, err := tryExec(query, r.pool, ctx); err != nil {
		return fmt.Errorf("PRRepository.RemoveReviewer failed to delete pr_reviewer: %w", err)
	}

	return nil
}

func (r *PRRepository) DeleteByID(ctx context.Context, prID string) error {
	query := r.sb.
		Delete("prs").
		Where(sq.Eq{"id": prID})

	if _, err := tryExec(query, r.pool, ctx); err != nil {
		return fmt.Errorf("PRRepository.DeleteByID failed to delete pr: %w", err)
	}

	return nil
}

func (r *PRRepository) GetAssignedReviewers(ctx context.Context, prID string) ([]string, error) {
	query := r.sb.
		Select("reviewer_id").
		From("pr_reviewers").
		Where(sq.Eq{"pr_id": prID})

	rows, err := tryQuery(query, r.pool, ctx)
	if err != nil {
		return nil, fmt.Errorf("PRRepository.GetAssignedReviewers failed to select reviewers: %w", err)
	}
	defer rows.Close()

	assignedIDs := make([]string, 0)
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("PRRepository.GetAssignedReviewers failed to scan reviewer ID: %w", err)
		}
		assignedIDs = append(assignedIDs, reviewerID)
	}

	return assignedIDs, nil
}
