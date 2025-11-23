package http

import (
	"encoding/json"
	nethttp "net/http"
)

func (s *Server) GetUsersGetReview(w nethttp.ResponseWriter, r *nethttp.Request, params GetUsersGetReviewParams) {
	if params.UserId == "" {
		writeError(w, nethttp.StatusBadRequest, NOTFOUND, "user_id required")
		return
	}

	prs, err := s.PRUseCase.GetAssignedTo(r.Context(), params.UserId)
	if err != nil {
		status, code := mapDomainError(err)
		writeError(w, status, code, err.Error())
		return
	}

	out := make([]PullRequestShort, 0, len(prs))
	for _, p := range prs {
		out = append(out, PullRequestShort{AuthorId: p.AuthorID, PullRequestId: p.ID, PullRequestName: p.Title, Status: PullRequestShortStatus(p.Status)})
	}
	resp := struct {
		UserId       string             `json:"user_id"`
		PullRequests []PullRequestShort `json:"pull_requests"`
	}{UserId: params.UserId, PullRequests: out}
	writeJSON(w, nethttp.StatusOK, resp)
}

func (s *Server) PostUsersSetIsActive(w nethttp.ResponseWriter, r *nethttp.Request) {
	var body PostUsersSetIsActiveJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, nethttp.StatusBadRequest, NOTFOUND, "invalid JSON body")
		return
	}
	if body.UserId == "" {
		writeError(w, nethttp.StatusBadRequest, NOTFOUND, "user_id required")
		return
	}

	u, err := s.UserUseCase.SetIsActive(r.Context(), body.UserId, body.IsActive)
	if err != nil {
		status, code := mapDomainError(err)
		writeError(w, status, code, err.Error())
		return
	}

	resp := struct {
		User User `json:"user"`
	}{User: User{UserId: u.ID, Username: u.Name, TeamName: "", IsActive: u.IsActive}}

	writeJSON(w, nethttp.StatusOK, resp)
}
