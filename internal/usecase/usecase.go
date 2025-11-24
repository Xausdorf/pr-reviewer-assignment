package usecase

import (
	"context"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
)

type PRRepository interface {
	Create(ctx context.Context, pr entity.PR) error
	GetByID(ctx context.Context, id string) (*entity.PR, error)
	UpdateStatus(ctx context.Context, id, status string) (*entity.PR, error)
	AssignReviewer(ctx context.Context, prID, teamName string) (reviewerID string, err error)
	RemoveReviewer(ctx context.Context, prID, reviewerID string) error
	DeleteByID(ctx context.Context, prID string) error
	GetAssignedReviewers(ctx context.Context, prID string) (assignedIDs []string, err error)
}

type TeamRepository interface {
	CreateTeam(ctx context.Context, team entity.Team, users []entity.User) error
	GetTeam(ctx context.Context, name string) (*entity.Team, []entity.User, error)
	GetTeamForUser(ctx context.Context, userID string) (string, error)
}

type UserRepository interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*entity.User, error)
	ListAssignedTo(ctx context.Context, userID string) ([]entity.PR, error)
	IsAssignedToPR(ctx context.Context, userID, prID string) (bool, error)
	GetByID(ctx context.Context, userID string) (*entity.User, error)
}
