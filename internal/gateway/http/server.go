package http

import (
	"context"
	"encoding/json"
	"errors"

	nethttp "net/http"

	log "github.com/sirupsen/logrus"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/apperror"
	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
)

type PRUseCase interface {
	CreatePullRequest(ctx context.Context, pr entity.PR) (assignedIDs []string, err error)
	MergePullRequest(ctx context.Context, prID string) (*entity.PR, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (string, *entity.PR, error)
	GetAssignedReviewers(ctx context.Context, prID string) (assignedIDs []string, err error)
}

type TeamUseCase interface {
	AddTeam(ctx context.Context, team entity.Team, users []entity.User) error
	GetTeam(ctx context.Context, name string) (*entity.Team, []entity.User, error)
}

type UserUseCase interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*entity.User, error)
	GetAssignedTo(ctx context.Context, userID string) ([]entity.PR, error)
}

type Server struct {
	PRUseCase   PRUseCase
	TeamUseCase TeamUseCase
	UserUseCase UserUseCase
	log         *log.Logger
}

func NewServer(pr PRUseCase, team TeamUseCase, user UserUseCase, logger *log.Logger) *Server {
	return &Server{
		PRUseCase:   pr,
		TeamUseCase: team,
		UserUseCase: user,
		log:         logger,
	}
}

func (s *Server) writeJSON(w nethttp.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.log.WithError(err).Error("Failed to write JSON response")
	}
}

func (s *Server) writeError(w nethttp.ResponseWriter, status int, code ErrorResponseErrorCode, message string) {
	resp := ErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = message

	logEntry := s.log.WithFields(log.Fields{
		"status":  status,
		"code":    code,
		"message": message,
	})
	if status == nethttp.StatusInternalServerError {
		logEntry.Error("Writing internal server error response")
	} else {
		logEntry.Info("Writing error response")
	}

	s.writeJSON(w, status, resp)
}

func mapDomainError(err error) (int, ErrorResponseErrorCode) {
	if errors.Is(err, apperror.ErrNoCandidate) {
		return nethttp.StatusBadRequest, NOCANDIDATE
	}
	if errors.Is(err, apperror.ErrPRExists) {
		return nethttp.StatusConflict, PREXISTS
	}
	if errors.Is(err, apperror.ErrNotAssigned) {
		return nethttp.StatusBadRequest, NOTASSIGNED
	}
	if errors.Is(err, apperror.ErrNotFound) {
		return nethttp.StatusNotFound, NOTFOUND
	}
	if errors.Is(err, apperror.ErrPRMerged) {
		return nethttp.StatusConflict, PRMERGED
	}
	if errors.Is(err, apperror.ErrTeamExists) {
		return nethttp.StatusBadRequest, TEAMEXISTS
	}
	return nethttp.StatusInternalServerError, NOTFOUND
}
