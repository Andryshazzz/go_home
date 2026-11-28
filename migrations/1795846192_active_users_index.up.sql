CREATE INDEX users_active_bio_age_idx ON users (age)
WHERE deleted_at IS NULL AND btrim(bio) <> '';