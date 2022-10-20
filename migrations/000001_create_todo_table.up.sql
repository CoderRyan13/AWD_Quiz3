-- Filename: migrations/000001_create_todo_table.up.sql

CREATE TABLE IF NOT EXISTS todos (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    task text NOT NULL,
    complete text NOT NULL DEFAULT 'NO',
    version integer NOT NULL DEFAULT 1
);