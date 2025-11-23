package http

import (
	"encoding/json"
	nethttp "net/http"
	"time"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
)

func (s *Server) PostTeamAdd(w nethttp.ResponseWriter, r *nethttp.Request) {
	var body Team
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, nethttp.StatusBadRequest, NOTFOUND, "invalid JSON body")
		return
	}

	if body.TeamName == "" {
		writeError(w, nethttp.StatusBadRequest, NOTFOUND, "team_name is required")
		return
	}

	team := &entity.Team{
		Name:      body.TeamName,
		CreatedAt: time.Now().UTC(),
	}

	members := make([]entity.TeamMember, 0, len(body.Members))
	users := make([]*entity.User, 0, len(body.Members))
	for _, m := range body.Members {
		if m.UserId == "" {
			writeError(w, nethttp.StatusBadRequest, NOTFOUND, "user_id is required for each member")
			return
		}
		members = append(members, entity.TeamMember{TeamName: body.TeamName, UserID: m.UserId})
		users = append(users, &entity.User{ID: m.UserId, Name: m.Username, IsActive: m.IsActive})
	}

	if err := s.TeamUseCase.AddOrUpdateTeam(r.Context(), team, members, users); err != nil {
		status, code := mapDomainError(err)
		writeError(w, status, code, err.Error())
		return
	}

	resp := struct {
		Team Team `json:"team"`
	}{Team: Team{TeamName: team.Name, Members: make([]TeamMember, 0, len(users))}}
	for _, u := range users {
		resp.Team.Members = append(resp.Team.Members, TeamMember{UserId: u.ID, Username: u.Name, IsActive: u.IsActive})
	}
	writeJSON(w, nethttp.StatusCreated, resp)
}

func (s *Server) GetTeamGet(w nethttp.ResponseWriter, r *nethttp.Request, params GetTeamGetParams) {
	if params.TeamName == "" {
		writeError(w, nethttp.StatusBadRequest, NOTFOUND, "team_name required")
		return
	}

	team, members, err := s.TeamUseCase.GetTeam(r.Context(), params.TeamName)
	if err != nil {
		status, code := mapDomainError(err)
		writeError(w, status, code, err.Error())
		return
	}

	resp := Team{TeamName: team.Name, Members: make([]TeamMember, 0, len(members))}
	for _, m := range members {
		u, err := s.UserUseCase.GetByID(r.Context(), m.UserID)
		if err != nil {
			resp.Members = append(resp.Members, TeamMember{IsActive: false, UserId: m.UserID, Username: ""})
			continue
		}
		resp.Members = append(resp.Members, TeamMember{IsActive: u.IsActive, UserId: u.ID, Username: u.Name})
	}

	writeJSON(w, nethttp.StatusOK, resp)
}
