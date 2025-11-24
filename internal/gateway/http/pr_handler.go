package http

import (
	"encoding/json"
	nethttp "net/http"
	"strings"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
	log "github.com/sirupsen/logrus"
)

type PullRequestResponse struct {
	PR PullRequest `json:"pr"`
}

type PullRequestReassignResponse struct {
	PR         PullRequest `json:"pr"`
	ReplacedBy string      `json:"replaced_by"`
}

func (s *Server) prToRepsponse(pr *entity.PR, assigned []string) PullRequestResponse {
	var prStatus PullRequestStatus
	switch strings.ToUpper(pr.Status) {
	case entity.PRStatusOpen:
		prStatus = PullRequestStatusOPEN
	case entity.PRStatusMerged:
		prStatus = PullRequestStatusMERGED
	default:
		s.log.WithFields(log.Fields{
			"pr_id":  pr.ID,
			"status": pr.Status,
		}).Warn("Unknown pull request status")
		prStatus = PullRequestStatus(pr.Status)
	}
	return PullRequestResponse{
		PR: PullRequest{
			AssignedReviewers: assigned,
			AuthorId:          pr.AuthorID,
			CreatedAt:         &pr.CreatedAt,
			MergedAt:          pr.MergedAt,
			PullRequestId:     pr.ID,
			PullRequestName:   pr.Title,
			Status:            prStatus,
		},
	}
}

func (s *Server) PostPullRequestCreate(w nethttp.ResponseWriter, r *nethttp.Request) {
	s.log.Info("Received request to create pull request")
	var body PostPullRequestCreateJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeError(w, nethttp.StatusBadRequest, NOTFOUND, "invalid JSON body")
		return
	}

	if body.AuthorId == "" || body.PullRequestId == "" || body.PullRequestName == "" {
		s.writeError(w, nethttp.StatusBadRequest, NOTFOUND, "author_id, pull_request_id and pull_request_name are required")
		return
	}

	pr := *entity.NewPR(body.PullRequestId, body.PullRequestName, body.AuthorId)

	assigned, err := s.PRUseCase.CreatePullRequest(r.Context(), pr)
	if err != nil {
		status, code := mapDomainError(err)
		s.writeError(w, status, code, err.Error())
		return
	}

	resp := s.prToRepsponse(&pr, assigned)

	s.writeJSON(w, nethttp.StatusCreated, resp)
	s.log.Info("Request to create pull request processed successfully")
}

func (s *Server) PostPullRequestMerge(w nethttp.ResponseWriter, r *nethttp.Request) {
	s.log.Info("Received request to merge pull request")
	var body PostPullRequestMergeJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeError(w, nethttp.StatusBadRequest, NOTFOUND, "invalid JSON body")
		return
	}

	if body.PullRequestId == "" {
		s.writeError(w, nethttp.StatusBadRequest, NOTFOUND, "pull_request_id is required")
		return
	}

	pr, err := s.PRUseCase.MergePullRequest(r.Context(), body.PullRequestId)
	if err != nil {
		status, code := mapDomainError(err)
		s.writeError(w, status, code, err.Error())
		return
	}

	assigned, err := s.PRUseCase.GetAssignedReviewers(r.Context(), body.PullRequestId)
	if err != nil {
		status, code := mapDomainError(err)
		s.writeError(w, status, code, err.Error())
		return
	}

	resp := s.prToRepsponse(pr, assigned)
	s.writeJSON(w, nethttp.StatusOK, resp)
	s.log.Info("Request to merge pull request processed successfully")
}

func (s *Server) PostPullRequestReassign(w nethttp.ResponseWriter, r *nethttp.Request) {
	s.log.Info("Received request to reassign pull request reviewer")
	var body PostPullRequestReassignJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeError(w, nethttp.StatusBadRequest, NOTFOUND, "invalid JSON body")
		return
	}

	if body.PullRequestId == "" || body.OldUserId == "" {
		s.writeError(w, nethttp.StatusBadRequest, NOTFOUND, "pull_request_id and old_user_id are required")
		return
	}

	newUserID, pr, err := s.PRUseCase.ReassignReviewer(r.Context(), body.PullRequestId, body.OldUserId)
	if err != nil {
		status, code := mapDomainError(err)
		s.writeError(w, status, code, err.Error())
		return
	}

	assigned, err := s.PRUseCase.GetAssignedReviewers(r.Context(), body.PullRequestId)
	if err != nil {
		status, code := mapDomainError(err)
		s.writeError(w, status, code, err.Error())
		return
	}

	prResp := s.prToRepsponse(pr, assigned)
	resp := PullRequestReassignResponse{
		PR:         prResp.PR,
		ReplacedBy: newUserID,
	}

	s.writeJSON(w, nethttp.StatusOK, resp)
	s.log.Info("Request to reassign pull request reviewer processed successfully")
}
