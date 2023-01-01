create extension if not exists "citext";
-- using Serial https://www.postgresqltutorial.com/postgresql-tutorial/postgresql-serial/
create table if not exists "credentials" (
    id smallserial,
    email citext unique not null check (email ~ '^[a-zA-Z0-9.!#$%&`’*+/=?^_\x60{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$'),
    password citext not null check (password <> ''),
    created_at timestamp not null default now()
);
-- INDEX on email
create index if not exists "credentials_email_idx" on "credentials" (email);