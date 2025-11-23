CREATE OR REPLACE FUNCTION prs_update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prs_update_updated_at
BEFORE UPDATE ON prs
FOR EACH ROW
EXECUTE FUNCTION prs_update_updated_at();