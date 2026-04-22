package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	periodicitySettings, err := marshalPeriodicitySettings(task.PeriodicitySettings)
	if err != nil {
		return nil, err
	}

	const query = `
		INSERT INTO tasks (
			title,
			description,
			status,
			schedule_start_at,
			schedule_end_at,
			periodicity_settings,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, title, description, status, schedule_start_at, schedule_end_at, periodicity_settings, created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.ScheduleStartAt,
		task.ScheduleEndAt,
		periodicitySettings,
		task.CreatedAt,
		task.UpdatedAt,
	)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, schedule_start_at, schedule_end_at, periodicity_settings, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	periodicitySettings, err := marshalPeriodicitySettings(task.PeriodicitySettings)
	if err != nil {
		return nil, err
	}

	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			schedule_start_at = $4,
			schedule_end_at = $5,
			periodicity_settings = $6,
			updated_at = $7
		WHERE id = $8
		RETURNING id, title, description, status, schedule_start_at, schedule_end_at, periodicity_settings, created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.ScheduleStartAt,
		task.ScheduleEndAt,
		periodicitySettings,
		task.UpdatedAt,
		task.ID,
	)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, schedule_start_at, schedule_end_at, periodicity_settings, created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *Repository) ListCalendarCandidates(ctx context.Context, at time.Time) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, schedule_start_at, schedule_end_at, periodicity_settings, created_at, updated_at
		FROM tasks
		WHERE periodicity_settings IS NOT NULL
			AND schedule_start_at IS NOT NULL
			AND schedule_start_at <= $1
			AND (schedule_end_at IS NULL OR schedule_end_at >= $1)
			AND NOT EXISTS (
				SELECT 1
				FROM task_completed_occurrences
				WHERE task_completed_occurrences.task_id = tasks.id
					AND task_completed_occurrences.scheduled_for = $1
			)
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query, at)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *Repository) CompleteOccurrence(ctx context.Context, taskID int64, scheduledFor, completedAt time.Time) error {
	const query = `
		INSERT INTO task_completed_occurrences (task_id, scheduled_for, completed_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (task_id, scheduled_for) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, taskID, scheduledFor, completedAt)

	return err
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task                   taskdomain.Task
		status                 string
		scheduleStartAt        pgtype.Timestamptz
		scheduleEndAt          pgtype.Timestamptz
		periodicitySettingsRaw []byte
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&scheduleStartAt,
		&scheduleEndAt,
		&periodicitySettingsRaw,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	periodicitySettings, err := unmarshalPeriodicitySettings(periodicitySettingsRaw)
	if err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	task.ScheduleStartAt = nullableTime(scheduleStartAt)
	task.ScheduleEndAt = nullableTime(scheduleEndAt)
	task.PeriodicitySettings = periodicitySettings

	return &task, nil
}

func marshalPeriodicitySettings(settings *taskdomain.PeriodicitySettings) ([]byte, error) {
	if settings == nil {
		return nil, nil
	}

	return json.Marshal(settings)
}

func unmarshalPeriodicitySettings(data []byte) (*taskdomain.PeriodicitySettings, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var settings taskdomain.PeriodicitySettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

func nullableTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	result := value.Time

	return &result
}
