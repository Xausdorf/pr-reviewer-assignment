package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/apperror"
	"github.com/Xausdorf/pr-reviewer-assignment/internal/usecase"
)

type Server struct {
	PRUseCase   usecase.PRUseCase
	TeamUseCase usecase.TeamUseCase
	UserUseCase usecase.UserUseCase
}

func NewServer(pr usecase.PRUseCase, team usecase.TeamUseCase, user usecase.UserUseCase) *Server {
	return &Server{
		PRUseCase:   pr,
		TeamUseCase: team,
		UserUseCase: user,
	}
}

func writeJSON(w nethttp.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w nethttp.ResponseWriter, status int, code ErrorResponseErrorCode, message string) {
	resp := ErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = message
	writeJSON(w, status, resp)
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
	return nethttp.StatusInternalServerError, NOTFOUND
}
