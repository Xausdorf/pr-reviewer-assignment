package http

import (
	"encoding/json"
	nethttp "net/http"
	"strings"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
	log "github.com/sirupsen/logrus"
)

type UsersSetIsActiveResponse struct {
	User User `json:"user"`
}

type UsersGetReviewResponse struct {
	UserId       string             `json:"user_id"`
	PullRequests []PullRequestShort `json:"pull_requests"`
}

func (s *Server) GetUsersGetReview(w nethttp.ResponseWriter, r *nethttp.Request, params GetUsersGetReviewParams) {
	s.log.Info("Received request to get user's assigned PRs")
	if params.UserId == "" {
		s.writeError(w, nethttp.StatusBadRequest, NOTFOUND, "user_id required")
		return
	}

	prs, err := s.UserUseCase.GetAssignedTo(r.Context(), params.UserId)
	if err != nil {
		status, code := mapDomainError(err)
		s.writeError(w, status, code, err.Error())
		return
	}

	out := make([]PullRequestShort, 0, len(prs))
	for _, p := range prs {
		var status PullRequestShortStatus
		switch strings.ToUpper(p.Status) {
		case entity.PRStatusOpen:
			status = PullRequestShortStatusOPEN
		case entity.PRStatusMerged:
			status = PullRequestShortStatusMERGED
		default:
			s.log.WithFields(log.Fields{
				"pr_id":  p.ID,
				"status": p.Status,
			}).Warn("Unknown pull request status")
			continue
		}
		out = append(out, PullRequestShort{
			AuthorId:        p.AuthorID,
			PullRequestId:   p.ID,
			PullRequestName: p.Title,
			Status:          status,
		})
	}

	resp := UsersGetReviewResponse{
		UserId:       params.UserId,
		PullRequests: out,
	}

	s.writeJSON(w, nethttp.StatusOK, resp)
	s.log.Info("Request to get user's assigned PRs processed successfully")
}

func (s *Server) PostUsersSetIsActive(w nethttp.ResponseWriter, r *nethttp.Request) {
	s.log.Info("Received request to set user active status")
	var body PostUsersSetIsActiveJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeError(w, nethttp.StatusBadRequest, NOTFOUND, "invalid JSON body")
		return
	}

	if body.UserId == "" {
		s.writeError(w, nethttp.StatusBadRequest, NOTFOUND, "user_id is required")
		return
	}

	u, err := s.UserUseCase.SetIsActive(r.Context(), body.UserId, body.IsActive)
	if err != nil {
		status, code := mapDomainError(err)
		s.writeError(w, status, code, err.Error())
		return
	}

	resp := UsersSetIsActiveResponse{User: User{
		UserId:   u.ID,
		Username: u.Name,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}}

	s.writeJSON(w, nethttp.StatusOK, resp)
	s.log.Info("Request to set user active status processed successfully")
}
