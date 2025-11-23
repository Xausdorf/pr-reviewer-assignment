package entity

import "time"

type PR struct {
	ID        string    `db:"id"`
	Title     string    `db:"title"`
	AuthorID  string    `db:"author_id"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type PRReviewer struct {
	PRID             string    `db:"pr_id"`
	ReviewerID       string    `db:"reviewer_id"`
	AssignedFromTeam string    `db:"assigned_from_team"`
	AssignedAt       time.Time `db:"assigned_at"`
}
