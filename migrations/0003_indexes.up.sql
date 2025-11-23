CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_pr_author ON prs(author_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer ON pr_reviewers(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_team_members_team ON team_members(team_name);
CREATE INDEX IF NOT EXISTS idx_team_members_user_active ON team_members(user_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_pr_assigned_team ON pr_reviewers(pr_id, assigned_from_team);
CREATE INDEX IF NOT EXISTS idx_prs_status ON prs(status);