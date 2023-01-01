CREATE extension IF NOT EXISTS "pgcrypto";
CREATE extension IF NOT EXISTS "citext";
CREATE temp TABLE IF NOT EXISTS "user" (
    id uuid primary key default gen_random_uuid(),
    email citext UNIQUE NOT NULL CHECK (
        email ~ '^[a-zA-Z0-9.!#$%&’*+/=?^_\x60{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$'
    ),
    username text unique not null check (username <> ''),
    password citext not null check (PASSWORD <> ''),
    created_at timestamp not null default now()
);
-- INDEX on username and email
-- CREATE INDEX IF NOT EXISTS "user_email_username_idx" ON "user" (email,username);