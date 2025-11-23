package usecase

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
	repo "github.com/Xausdorf/pr-reviewer-assignment/internal/repository"
)

type UserService struct {
	userRepo repo.UserRepository
	log      *log.Logger
}

func NewUserService(user repo.UserRepository, logger *log.Logger) *UserService {
	return &UserService{userRepo: user, log: logger}
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*entity.User, error) {
	if err := s.userRepo.SetIsActive(ctx, userID, isActive); err != nil {
		return nil, err
	}
	return s.userRepo.GetByID(ctx, userID)
}

func (s *UserService) GetByID(ctx context.Context, id string) (*entity.User, error) {
	return s.userRepo.GetByID(ctx, id)
}
