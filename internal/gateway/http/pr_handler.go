package http

import (
	"encoding/json"
	nethttp "net/http"
	"time"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
)

func (s *Server) PostPullRequestCreate(w nethttp.ResponseWriter, r *nethttp.Request) {
	var body PostPullRequestCreateJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, nethttp.StatusBadRequest, NOTFOUND, "invalid JSON body")
		return
	}

	if body.AuthorId == "" || body.PullRequestId == "" || body.PullRequestName == "" {
		writeError(w, nethttp.StatusBadRequest, NOTFOUND, "author_id, pull_request_id and pull_request_name are required")
		return
	}

	pr := &entity.PR{
		ID:        body.PullRequestId,
		Title:     body.PullRequestName,
		AuthorID:  body.AuthorId,
		Status:    string(PullRequestStatusOPEN),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	assigned, err := s.PRUseCase.CreatePullRequest(r.Context(), pr)
	if err != nil {
		status, code := mapDomainError(err)
		writeError(w, status, code, err.Error())
		return
	}

	createdAt := pr.CreatedAt
	mergedAt := (*time.Time)(nil)

	resp := PullRequest{
		AssignedReviewers: assigned,
		AuthorId:          pr.AuthorID,
		CreatedAt:         &createdAt,
		MergedAt:          mergedAt,
		PullRequestId:     pr.ID,
		PullRequestName:   pr.Title,
		Status:            PullRequestStatus(pr.Status),
	}

	writeJSON(w, nethttp.StatusCreated, struct {
		Pr PullRequest `json:"pr"`
	}{Pr: resp})
}

func (s *Server) PostPullRequestMerge(w nethttp.ResponseWriter, r *nethttp.Request) {
	var body PostPullRequestMergeJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, nethttp.StatusBadRequest, NOTFOUND, "invalid JSON body")
		return
	}
	if body.PullRequestId == "" {
		writeError(w, nethttp.StatusBadRequest, NOTFOUND, "pull_request_id is required")
		return
	}

	pr, err := s.PRUseCase.MergePullRequest(r.Context(), body.PullRequestId)
	if err != nil {
		status, code := mapDomainError(err)
		writeError(w, status, code, err.Error())
		return
	}

	assigned := make([]string, 0)
	createdAt := pr.CreatedAt
	mergedAt := pr.UpdatedAt
	resp := PullRequest{
		AssignedReviewers: assigned,
		AuthorId:          pr.AuthorID,
		CreatedAt:         &createdAt,
		MergedAt:          &mergedAt,
		PullRequestId:     pr.ID,
		PullRequestName:   pr.Title,
		Status:            PullRequestStatus(pr.Status),
	}
	writeJSON(w, nethttp.StatusOK, struct {
		Pr PullRequest `json:"pr"`
	}{Pr: resp})
}

func (s *Server) PostPullRequestReassign(w nethttp.ResponseWriter, r *nethttp.Request) {
	var body PostPullRequestReassignJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, nethttp.StatusBadRequest, NOTFOUND, "invalid JSON body")
		return
	}
	if body.PullRequestId == "" || body.OldUserId == "" {
		writeError(w, nethttp.StatusBadRequest, NOTFOUND, "pull_request_id and old_user_id are required")
		return
	}

	newID, assigned, err := s.PRUseCase.ReassignReviewer(r.Context(), body.PullRequestId, body.OldUserId, "")
	if err != nil {
		status, code := mapDomainError(err)
		writeError(w, status, code, err.Error())
		return
	}

	prEntity, err := s.PRUseCase.GetByID(r.Context(), body.PullRequestId)
	var prResp PullRequest
	if err == nil && prEntity != nil {
		createdAt := prEntity.CreatedAt
		var mergedAt *time.Time
		if prEntity.Status == "MERGED" {
			mergedAt = &prEntity.UpdatedAt
		}
		prResp = PullRequest{AssignedReviewers: assigned, AuthorId: prEntity.AuthorID, CreatedAt: &createdAt, MergedAt: mergedAt, PullRequestId: prEntity.ID, PullRequestName: prEntity.Title, Status: PullRequestStatus(prEntity.Status)}
	} else {
		prResp = PullRequest{AssignedReviewers: assigned, PullRequestId: body.PullRequestId}
	}

	writeJSON(w, nethttp.StatusOK, struct {
		Pr         PullRequest `json:"pr"`
		ReplacedBy string      `json:"replaced_by"`
	}{Pr: prResp, ReplacedBy: newID})
}
