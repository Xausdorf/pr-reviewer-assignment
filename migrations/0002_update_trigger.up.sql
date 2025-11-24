CREATE OR REPLACE FUNCTION prs_update_merged_at()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.status = 'MERGED' AND OLD.status = 'OPEN' THEN 
    NEW.updated_at = now();
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prs_update_merged_at
BEFORE UPDATE ON prs
FOR EACH ROW
WHEN (NEW.status = 'MERGED' AND OLD.status = 'OPEN')
EXECUTE FUNCTION prs_update_merged_at();