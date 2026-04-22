ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS periodicity_settings JSONB;
