BEGIN;

ALTER TABLE gitserver_repos ADD COLUMN IF NOT EXISTS last_changed TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now();

COMMIT;
