package usecase

import (
	"context"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
)

type PRUseCase interface {
	CreatePullRequest(ctx context.Context, pr *entity.PR) ([]string, error)
	MergePullRequest(ctx context.Context, prID string) (*entity.PR, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID, newUserID string) (string, []string, error)
	GetByID(ctx context.Context, prID string) (*entity.PR, error)
	GetAssignedTo(ctx context.Context, userID string) ([]*entity.PR, error)
}

type TeamUseCase interface {
	AddOrUpdateTeam(ctx context.Context, team *entity.Team, members []entity.TeamMember, users []*entity.User) error
	GetTeam(ctx context.Context, name string) (*entity.Team, []entity.TeamMember, error)
}

type UserUseCase interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*entity.User, error)
	GetByID(ctx context.Context, id string) (*entity.User, error)
}
