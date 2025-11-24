package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/Xausdorf/pr-reviewer-assignment/internal/entity"
)

type TeamAddResponse struct {
	Team Team `json:"team"`
}

func TeamFromEntity(e entity.Team, members []entity.User) Team {
	team := Team{
		TeamName: e.Name,
		Members:  make([]TeamMember, 0, len(members)),
	}
	for _, m := range members {
		team.Members = append(team.Members, TeamMember{
			UserId:   m.ID,
			Username: m.Name,
			IsActive: m.IsActive,
		})
	}
	return team
}

func (s *Server) PostTeamAdd(w nethttp.ResponseWriter, r *nethttp.Request) {
	s.log.Info("Received request to add team")
	var body Team
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeError(w, nethttp.StatusBadRequest, NOTFOUND, "invalid JSON body")
		return
	}

	if body.TeamName == "" {
		s.writeError(w, nethttp.StatusBadRequest, NOTFOUND, "team_name is required")
		return
	}

	team := *entity.NewTeam(body.TeamName)

	users := make([]entity.User, 0, len(body.Members))
	for _, m := range body.Members {
		if m.UserId == "" {
			s.writeError(w, nethttp.StatusBadRequest, NOTFOUND, "user_id is required for each member")
			return
		}
		user := *entity.NewUser(m.UserId, m.Username, body.TeamName, m.IsActive)
		users = append(users, user)
	}

	if err := s.TeamUseCase.AddTeam(r.Context(), team, users); err != nil {
		status, code := mapDomainError(err)
		s.writeError(w, status, code, err.Error())
		return
	}

	resp := TeamAddResponse{Team: TeamFromEntity(team, users)}
	s.writeJSON(w, nethttp.StatusCreated, resp)
	s.log.Info("Request to add team processed successfully")
}

func (s *Server) GetTeamGet(w nethttp.ResponseWriter, r *nethttp.Request, params GetTeamGetParams) {
	s.log.Info("Received request to get team")
	if params.TeamName == "" {
		s.writeError(w, nethttp.StatusBadRequest, NOTFOUND, "team_name is required")
		return
	}

	team, members, err := s.TeamUseCase.GetTeam(r.Context(), params.TeamName)
	if err != nil {
		status, code := mapDomainError(err)
		s.writeError(w, status, code, err.Error())
		return
	}

	resp := TeamFromEntity(*team, members)

	s.writeJSON(w, nethttp.StatusOK, resp)
	s.log.Info("Request to get team processed successfully")
}
