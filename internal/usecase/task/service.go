package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:               normalized.Title,
		Description:         normalized.Description,
		Status:              normalized.Status,
		ScheduleStartAt:     normalized.ScheduleStartAt,
		ScheduleEndAt:       normalized.ScheduleEndAt,
		PeriodicitySettings: normalized.PeriodicitySettings,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:                  id,
		Title:               normalized.Title,
		Description:         normalized.Description,
		Status:              normalized.Status,
		ScheduleStartAt:     normalized.ScheduleStartAt,
		ScheduleEndAt:       normalized.ScheduleEndAt,
		PeriodicitySettings: normalized.PeriodicitySettings,
		UpdatedAt:           s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func (s *Service) ListForCalendar(ctx context.Context, at time.Time) ([]taskdomain.Task, error) {
	candidates, err := s.repo.ListCalendarCandidates(ctx, at)
	if err != nil {
		return nil, err
	}

	tasks := make([]taskdomain.Task, 0, len(candidates))
	for i := range candidates {
		if matchesCalendarDateTime(&candidates[i], at) {
			tasks = append(tasks, candidates[i])
		}
	}

	return tasks, nil
}

func (s *Service) CompleteOccurrence(ctx context.Context, id int64, scheduledFor time.Time) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !matchesCalendarDateTime(task, scheduledFor) {
		return fmt.Errorf("%w: task does not occur at the specified date and time", ErrInvalidInput)
	}

	return s.repo.CompleteOccurrence(ctx, id, scheduledFor, s.now())
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if err := validateSchedule(input.ScheduleStartAt, input.ScheduleEndAt, input.PeriodicitySettings); err != nil {
		return CreateInput{}, err
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if err := validateSchedule(input.ScheduleStartAt, input.ScheduleEndAt, input.PeriodicitySettings); err != nil {
		return UpdateInput{}, err
	}

	return input, nil
}
