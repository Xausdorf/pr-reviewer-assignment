package usecase

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
)

type TeamUseCase struct {
	teamRepo TeamRepository
	log      *log.Logger
}

func NewTeamUseCase(team TeamRepository, logger *log.Logger) *TeamUseCase {
	return &TeamUseCase{teamRepo: team, log: logger}
}

func (s *TeamUseCase) AddTeam(ctx context.Context, team entity.Team, users []entity.User) error {
	s.log.WithFields(log.Fields{
		"team":  team.Name,
		"count": len(users),
	}).Info("TeamUseCase - adding team")
	return s.teamRepo.CreateTeam(ctx, team, users)
}

func (s *TeamUseCase) GetTeam(ctx context.Context, name string) (*entity.Team, []entity.User, error) {
	s.log.WithField("team", name).Info("TeamUseCase - getting team")
	return s.teamRepo.GetTeam(ctx, name)
}
