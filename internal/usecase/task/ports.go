package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
	ListCalendarCandidates(ctx context.Context, at time.Time) ([]taskdomain.Task, error)
	CompleteOccurrence(ctx context.Context, taskID int64, scheduledFor, completedAt time.Time) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
	ListForCalendar(ctx context.Context, at time.Time) ([]taskdomain.Task, error)
	CompleteOccurrence(ctx context.Context, id int64, scheduledFor time.Time) error
}

type CreateInput struct {
	Title               string
	Description         string
	State               taskdomain.State
	ScheduleStartAt     *time.Time
	ScheduleEndAt       *time.Time
	PeriodicitySettings *taskdomain.PeriodicitySettings
}

type UpdateInput struct {
	Title               string
	Description         string
	State               taskdomain.State
	ScheduleStartAt     *time.Time
	ScheduleEndAt       *time.Time
	PeriodicitySettings *taskdomain.PeriodicitySettings
}
