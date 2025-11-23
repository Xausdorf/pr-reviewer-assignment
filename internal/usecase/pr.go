package usecase

import (
	"context"
	"math/rand"
	"time"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/apperror"
	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
	repo "github.com/Xausdorf/pr-reviewer-assignment/internal/repository"
	log "github.com/sirupsen/logrus"
)

type PRService struct {
	prRepo   repo.PRRepository
	teamRepo repo.TeamRepository
	userRepo repo.UserRepository
	log      *log.Logger
}

func NewPRService(pr repo.PRRepository, team repo.TeamRepository, user repo.UserRepository, logger *log.Logger) *PRService {
	return &PRService{prRepo: pr, teamRepo: team, userRepo: user, log: logger}
}

func (s *PRService) CreatePullRequest(ctx context.Context, pr *entity.PR) ([]string, error) {
	if existing, _ := s.prRepo.GetByID(ctx, pr.ID); existing != nil {
		return nil, apperror.ErrPRExists
	}

	teams, err := s.teamRepo.GetTeamsForUser(ctx, pr.AuthorID)
	if err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		if err := s.prRepo.Create(ctx, pr); err != nil {
			return nil, err
		}
		return nil, nil
	}

	teamName := teams[0]
	team, members, err := s.teamRepo.GetTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}

	candidates := make([]string, 0, len(members))
	for _, m := range members {
		if m.UserID == pr.AuthorID {
			continue
		}
		u, err := s.userRepo.GetByID(ctx, m.UserID)
		if err != nil {
			continue
		}
		if u.IsActive {
			candidates = append(candidates, m.UserID)
		}
	}

	if err := s.prRepo.Create(ctx, pr); err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	n := 2
	if len(candidates) < n {
		n = len(candidates)
	}
	selected := candidates[:n]

	now := time.Now().UTC()
	for _, rid := range selected {
		rv := &entity.PRReviewer{PRID: pr.ID, ReviewerID: rid, AssignedFromTeam: team.Name, AssignedAt: now}
		if err := s.prRepo.AddReviewer(ctx, rv); err != nil {
			s.log.WithError(err).WithField("pr", pr.ID).Error("failed to add reviewer")
		}
	}

	return selected, nil
}

func (s *PRService) MergePullRequest(ctx context.Context, prID string) (*entity.PR, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, apperror.ErrNotFound
	}
	if pr.Status == "MERGED" {
		return pr, nil
	}
	if err := s.prRepo.UpdateStatus(ctx, prID, "MERGED"); err != nil {
		return nil, err
	}
	updated, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *PRService) ReassignReviewer(ctx context.Context, prID, oldUserID, newUserID string) (string, []string, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return "", nil, apperror.ErrNotFound
	}
	if pr.Status == "MERGED" {
		return "", nil, apperror.ErrPRMerged
	}

	reviewers, err := s.prRepo.ListReviewers(ctx, prID)
	if err != nil {
		return "", nil, err
	}
	var found bool
	var fromTeam string
	assigned := map[string]struct{}{}
	for _, r := range reviewers {
		assigned[r.ReviewerID] = struct{}{}
		if r.ReviewerID == oldUserID {
			found = true
			fromTeam = r.AssignedFromTeam
		}
	}
	if !found {
		return "", nil, apperror.ErrNotAssigned
	}

	_, members, err := s.teamRepo.GetTeam(ctx, fromTeam)
	if err != nil {
		return "", nil, err
	}
	candidates := make([]string, 0)
	for _, m := range members {
		if m.UserID == oldUserID {
			continue
		}
		if _, ok := assigned[m.UserID]; ok {
			continue
		}
		u, err := s.userRepo.GetByID(ctx, m.UserID)
		if err != nil {
			continue
		}
		if u.IsActive {
			candidates = append(candidates, m.UserID)
		}
	}
	if len(candidates) == 0 {
		return "", nil, apperror.ErrNoCandidate
	}
	newID := candidates[rand.Intn(len(candidates))]

	if err := s.prRepo.RemoveReviewer(ctx, prID, oldUserID); err != nil {
		return "", nil, err
	}
	now := time.Now().UTC()
	if err := s.prRepo.AddReviewer(ctx, &entity.PRReviewer{PRID: prID, ReviewerID: newID, AssignedFromTeam: fromTeam, AssignedAt: now}); err != nil {
		return "", nil, err
	}
	revs, err := s.prRepo.ListReviewers(ctx, prID)
	if err != nil {
		return newID, nil, nil
	}
	ids := make([]string, 0, len(revs))
	for _, r := range revs {
		ids = append(ids, r.ReviewerID)
	}
	return newID, ids, nil
}

func (s *PRService) GetAssignedTo(ctx context.Context, userID string) ([]*entity.PR, error) {
	return s.prRepo.ListAssignedTo(ctx, userID)
}

func (s *PRService) GetByID(ctx context.Context, prID string) (*entity.PR, error) {
	return s.prRepo.GetByID(ctx, prID)
}
