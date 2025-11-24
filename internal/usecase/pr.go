package usecase

import (
	"context"
	"errors"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/apperror"
	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
	log "github.com/sirupsen/logrus"
)

const MaxReviewersCount = 2

type PRUseCase struct {
	prRepo   PRRepository
	userRepo UserRepository
	log      *log.Logger
}

func NewPRUseCase(pr PRRepository, user UserRepository, logger *log.Logger) *PRUseCase {
	return &PRUseCase{prRepo: pr, userRepo: user, log: logger}
}

func (s *PRUseCase) CreatePullRequest(ctx context.Context, pr entity.PR) ([]string, error) {
	s.log.WithFields(log.Fields{
		"prID":     pr.ID,
		"prName":   pr.Title,
		"authorID": pr.AuthorID,
	}).Info("PRUseCase - creating pull request")
	author, err := s.userRepo.GetByID(ctx, pr.AuthorID)
	if err != nil {
		return nil, err
	}

	if err := s.prRepo.Create(ctx, pr); err != nil {
		return nil, err
	}

	assigned := make([]string, 0, MaxReviewersCount)
	for i := 0; i < MaxReviewersCount; i++ {
		reviewerID, err := s.prRepo.AssignReviewer(ctx, pr.ID, author.TeamName)
		if errors.Is(err, apperror.ErrNoCandidate) {
			break
		}
		if err != nil {
			_ = s.prRepo.DeleteByID(ctx, pr.ID)
			return nil, err
		}
		assigned = append(assigned, reviewerID)
	}

	return assigned, nil
}

func (s *PRUseCase) MergePullRequest(ctx context.Context, prID string) (*entity.PR, error) {
	s.log.WithField("prID", prID).Info("PRUseCase - merging pull request")
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, err
	}
	if pr.Status == entity.PRStatusMerged {
		return pr, nil
	}

	return s.prRepo.UpdateStatus(ctx, prID, entity.PRStatusMerged)
}

func (s *PRUseCase) ReassignReviewer(ctx context.Context, prID, oldUserID string) (string, *entity.PR, error) {
	s.log.WithFields(log.Fields{
		"prID":      prID,
		"oldUserID": oldUserID,
	}).Info("PRUseCase - reassigning reviewer")
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return "", nil, err
	}
	if pr.Status == entity.PRStatusMerged {
		return "", nil, apperror.ErrPRMerged
	}

	oldUser, err := s.userRepo.GetByID(ctx, oldUserID)
	if err != nil {
		return "", nil, err
	}

	isAssigned, err := s.userRepo.IsAssignedToPR(ctx, oldUserID, prID)
	if err != nil {
		return "", nil, err
	}
	if !isAssigned {
		return "", nil, apperror.ErrNotAssigned
	}

	newUserID, err := s.prRepo.AssignReviewer(ctx, prID, oldUser.TeamName)
	if err != nil {
		return "", nil, err
	}
	if err := s.prRepo.RemoveReviewer(ctx, prID, oldUserID); err != nil {
		_ = s.prRepo.RemoveReviewer(ctx, prID, newUserID)
		return "", nil, err
	}

	return newUserID, pr, nil
}

func (s *PRUseCase) GetAssignedReviewers(ctx context.Context, prID string) (assignedIDs []string, err error) {
	s.log.WithField("prID", prID).Info("PRUseCase - getting assigned reviewers")
	return s.prRepo.GetAssignedReviewers(ctx, prID)
}
