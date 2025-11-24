package entity

import "time"

const (
	PRStatusOpen   = "OPEN"
	PRStatusMerged = "MERGED"
)

type PR struct {
	ID        string     `db:"id"`
	Title     string     `db:"title"`
	AuthorID  string     `db:"author_id"`
	Status    string     `db:"status"`
	CreatedAt time.Time  `db:"created_at"`
	MergedAt  *time.Time `db:"merged_at"`
}

type PRReviewer struct {
	PRID       string    `db:"pr_id"`
	ReviewerID string    `db:"reviewer_id"`
	AssignedAt time.Time `db:"assigned_at"`
}

func NewPR(id, title, authorID string) *PR {
	return &PR{
		ID:        id,
		Title:     title,
		AuthorID:  authorID,
		Status:    PRStatusOpen,
		CreatedAt: time.Now().UTC(),
		MergedAt:  nil,
	}
}
