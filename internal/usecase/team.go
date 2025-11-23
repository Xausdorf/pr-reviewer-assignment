package usecase

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
	repo "github.com/Xausdorf/pr-reviewer-assignment/internal/repository"
)

type TeamService struct {
	teamRepo repo.TeamRepository
	userRepo repo.UserRepository
	log      *log.Logger
}

func NewTeamService(team repo.TeamRepository, user repo.UserRepository, logger *log.Logger) *TeamService {
	return &TeamService{teamRepo: team, userRepo: user, log: logger}
}

func (s *TeamService) AddOrUpdateTeam(ctx context.Context, team *entity.Team, members []entity.TeamMember, users []*entity.User) error {
	for _, u := range users {
		if err := s.userRepo.CreateOrUpdateUser(ctx, u); err != nil {
			s.log.WithError(err).WithField("user", u.ID).Error("failed to create/update user")
		}
	}
	return s.teamRepo.CreateOrUpdateTeam(ctx, team, members)
}

func (s *TeamService) GetTeam(ctx context.Context, name string) (*entity.Team, []entity.TeamMember, error) {
	return s.teamRepo.GetTeam(ctx, name)
}
