-- Filename: migrations/000003_add_forums_indexes.up.sql
CREATE INDEX IF NOT EXISTS forums_name_idx ON forums USING GIN(to_tsvector('simple', name));
CREATE INDEX IF NOT EXISTS forums_level_idx ON forums USING GIN(to_tsvector('simple', level));
CREATE INDEX IF NOT EXISTS forums_mode_idx ON forums USING GIN(mode);