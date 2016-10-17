CREATE TABLE tasks (
  id UUID PRIMARY KEY,
  title TEXT NOT NULL,
  start_date TIMESTAMP,
  end_date TIMESTAMP,
  done BOOL NOT NULL DEFAULT false
);
