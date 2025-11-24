package usecase

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
)

type UserUseCase struct {
	userRepo UserRepository
	log      *log.Logger
}

func NewUserUseCase(user UserRepository, logger *log.Logger) *UserUseCase {
	return &UserUseCase{userRepo: user, log: logger}
}

func (s *UserUseCase) SetIsActive(ctx context.Context, userID string, isActive bool) (*entity.User, error) {
	s.log.WithFields(log.Fields{
		"userID":   userID,
		"isActive": isActive,
	}).Info("UserUseCase - setting user active status")
	return s.userRepo.SetIsActive(ctx, userID, isActive)
}

func (s *UserUseCase) GetAssignedTo(ctx context.Context, userID string) ([]entity.PR, error) {
	s.log.WithField("userID", userID).Info("UserUseCase - getting assigned PRs")
	return s.userRepo.ListAssignedTo(ctx, userID)
}
