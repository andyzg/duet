CREATE TABLE tasks (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  title TEXT NOT NULL,
  start_date TIMESTAMP,
  end_date TIMESTAMP,
  done BOOL NOT NULL DEFAULT false
);
