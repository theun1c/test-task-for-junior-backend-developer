ALTER TABLE tasks
RENAME COLUMN status TO state;

UPDATE tasks
SET state = CASE state
	WHEN 'new' THEN 'active'
	WHEN 'in_progress' THEN 'active'
	WHEN 'done' THEN 'archived'
	ELSE state
END;

DROP INDEX IF EXISTS idx_tasks_status;
CREATE INDEX IF NOT EXISTS idx_tasks_state ON tasks (state);
