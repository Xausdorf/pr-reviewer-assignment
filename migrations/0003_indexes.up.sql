CREATE INDEX IF NOT EXISTS idx_users_team_name ON users(team_name);
CREATE INDEX IF NOT EXISTS idx_users_is_active_team_name ON users(is_active, team_name);
CREATE INDEX IF NOT EXISTS idx_pr_author ON prs(author_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer ON pr_reviewers(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_pr_id ON pr_reviewers(pr_id);