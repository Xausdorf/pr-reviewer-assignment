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

type PRRepository struct {
	pool *pgxpool.Pool
	sb   sq.StatementBuilderType
	log  *log.Logger
}

func NewPRRepository(pool *pgxpool.Pool, logger *log.Logger) repo.PRRepository {
	return &PRRepository{
		pool: pool,
		sb:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		log:  logger,
	}
}

func (r *PRRepository) Create(ctx context.Context, pr *entity.PR) error {
	q := r.sb.
		Insert("prs").
		Columns("id", "title", "author_id", "status", "created_at", "updated_at").
		Values(pr.ID, pr.Title, pr.AuthorID, pr.Status, pr.CreatedAt, pr.UpdatedAt)

	sql, args, _ := q.ToSql()
	_, err := r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *PRRepository) GetByID(ctx context.Context, id string) (*entity.PR, error) {
	q := r.sb.
		Select("id", "title", "author_id", "status", "created_at", "updated_at").
		From("prs").
		Where(sq.Eq{"id": id})

	sql, args, _ := q.ToSql()
	row := r.pool.QueryRow(ctx, sql, args...)

	var pr entity.PR
	if err := row.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.WithField("pr_id", id).Debug("pr not found")
			return nil, apperror.ErrNotFound
		}
		r.log.WithError(err).WithField("pr_id", id).Error("failed to scan pr")
		return nil, err
	}
	return &pr, nil
}

func (r *PRRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	q := r.sb.
		Update("prs").
		Set("status", status).
		Where(sq.Eq{"id": id})

	sql, args, _ := q.ToSql()
	_, err := r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *PRRepository) AddReviewer(ctx context.Context, rv *entity.PRReviewer) error {
	q := r.sb.
		Insert("pr_reviewers").
		Columns("pr_id", "reviewer_id", "assigned_from_team", "assigned_at").
		Values(rv.PRID, rv.ReviewerID, rv.AssignedFromTeam, rv.AssignedAt)

	sql, args, _ := q.ToSql()
	_, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		r.log.WithError(err).WithFields(log.Fields{"pr_id": rv.PRID, "reviewer_id": rv.ReviewerID}).Error("failed to insert pr_reviewer")
	}
	return err
}

func (r *PRRepository) RemoveReviewer(ctx context.Context, prID, reviewerID string) error {
	q := r.sb.
		Delete("pr_reviewers").
		Where(sq.Eq{"pr_id": prID, "reviewer_id": reviewerID})

	sql, args, _ := q.ToSql()
	_, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		r.log.WithError(err).WithFields(log.Fields{"pr_id": prID, "reviewer_id": reviewerID}).Error("failed to remove pr_reviewer")
	}
	return err
}

func (r *PRRepository) ListAssignedTo(ctx context.Context, userID string) ([]*entity.PR, error) {
	q := r.sb.
		Select("p.id", "p.title", "p.author_id", "p.status", "p.created_at", "p.updated_at").
		From("prs p").Join("pr_reviewers r ON p.id = r.pr_id").
		Where(sq.Eq{"r.reviewer_id": userID})

	sql, args, _ := q.ToSql()
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		r.log.WithError(err).WithField("user_id", userID).Error("failed to query assigned prs")
		return nil, err
	}
	defer rows.Close()

	out := make([]*entity.PR, 0)
	for rows.Next() {
		var pr entity.PR
		if err := rows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.UpdatedAt); err != nil {
			r.log.WithError(err).Error("failed to scan assigned pr")
			return nil, err
		}
		out = append(out, &pr)
	}
	return out, nil
}

func (r *PRRepository) ListReviewers(ctx context.Context, prID string) ([]entity.PRReviewer, error) {
	q := r.sb.
		Select("pr_id", "reviewer_id", "assigned_from_team", "assigned_at").
		From("pr_reviewers").
		Where(sq.Eq{"pr_id": prID})

	sql, args, _ := q.ToSql()
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.PRReviewer, 0)
	for rows.Next() {
		var rv entity.PRReviewer
		if err := rows.Scan(&rv.PRID, &rv.ReviewerID, &rv.AssignedFromTeam, &rv.AssignedAt); err != nil {
			return nil, err
		}
		out = append(out, rv)
	}
	return out, nil
}
