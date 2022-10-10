-- Filename: migrations/000002_add_forums_check_constraints.up.sql

ALTER TABLE forums ADD CONSTRAINT mode_length_check CHECK (array_length(mode, 1) BETWEEN 1 AND 5);