package apperror

import "errors"

var (
	ErrNoCandidate = errors.New("no candidate available")
	ErrPRExists    = errors.New("pr already exists")
	ErrNotAssigned = errors.New("not assigned")
	ErrNotFound    = errors.New("not found")
	ErrPRMerged    = errors.New("pr merged")
	ErrTeamExists  = errors.New("team already exists")
)
