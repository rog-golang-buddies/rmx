CREATE TABLE users (
  id   BIGINT  NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(255)    NOT NULL,
  email  VARCHAR(255)   NOT NULL,
  role  VARCHAR(255)    NOT NULL    DEFAULT 'basic',
  CHECK (role in ('basic', 'moderator', 'admin')),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);
