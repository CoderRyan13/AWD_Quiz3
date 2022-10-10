-- Filename: migrations/000002_add_forums_check_constraints.down.sql

ALTER TABLE forums DROP CONSTRAINT IF EXISTS mode_length_check;