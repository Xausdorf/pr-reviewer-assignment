package entity

import "time"

type User struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
}

type Team struct {
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

type TeamMember struct {
	TeamName string `db:"team_name"`
	UserID   string `db:"user_id"`
}
