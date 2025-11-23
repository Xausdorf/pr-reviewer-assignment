package repository

import (
	"context"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
)

type PRRepository interface {
	Create(ctx context.Context, pr *entity.PR) error
	GetByID(ctx context.Context, id string) (*entity.PR, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	AddReviewer(ctx context.Context, r *entity.PRReviewer) error
	RemoveReviewer(ctx context.Context, prID, reviewerID string) error
	ListAssignedTo(ctx context.Context, userID string) ([]*entity.PR, error)
	ListReviewers(ctx context.Context, prID string) ([]entity.PRReviewer, error)
}

type TeamRepository interface {
	CreateOrUpdateTeam(ctx context.Context, team *entity.Team, members []entity.TeamMember) error
	GetTeam(ctx context.Context, name string) (*entity.Team, []entity.TeamMember, error)
	GetTeamsForUser(ctx context.Context, userID string) ([]string, error)
}

type UserRepository interface {
	CreateOrUpdateUser(ctx context.Context, u *entity.User) error
	SetIsActive(ctx context.Context, userID string, isActive bool) error
	GetByID(ctx context.Context, id string) (*entity.User, error)
}
