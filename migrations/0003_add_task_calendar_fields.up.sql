ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS schedule_start_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS schedule_end_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS task_completed_occurrences (
	task_id BIGINT NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
	scheduled_for TIMESTAMPTZ NOT NULL,
	completed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (task_id, scheduled_for)
);
