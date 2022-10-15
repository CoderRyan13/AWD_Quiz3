-- Filename: migrations/000003_add_forums_indexes.down.sql
DROP INDEX IF EXISTS forums_name_idx;
DROP INDEX IF EXISTS forums_level_idx;
DROP INDEX IF EXISTS forums_mode_idx;