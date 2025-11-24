package entity

import "time"

type User struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	TeamName  string    `db:"team_name"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
}

type Team struct {
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

func NewUser(id, name, teamName string, isActive bool) *User {
	return &User{
		ID:        id,
		Name:      name,
		TeamName:  teamName,
		IsActive:  isActive,
		CreatedAt: time.Now().UTC(),
	}
}

func NewTeam(name string) *Team {
	return &Team{
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}
}
