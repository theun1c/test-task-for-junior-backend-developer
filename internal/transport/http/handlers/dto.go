package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type periodicitySettingsDTO struct {
	Type              taskdomain.PeriodicityType `json:"type"`
	EveryNDays        *int                       `json:"every_n_days,omitempty"`
	DaysOfMonth       []int                      `json:"days_of_month,omitempty"`
	SpecificDateTimes []string                   `json:"specific_datetimes,omitempty"`
	DayParity         *taskdomain.DayParity      `json:"day_parity,omitempty"`
}

type taskMutationDTO struct {
	Title               string                  `json:"title"`
	Description         string                  `json:"description"`
	State               taskdomain.State        `json:"state"`
	ScheduleStartAt     *string                 `json:"schedule_start_at,omitempty"`
	ScheduleEndAt       *string                 `json:"schedule_end_at,omitempty"`
	PeriodicitySettings *periodicitySettingsDTO `json:"periodicity_settings,omitempty"`
}

type taskDTO struct {
	ID                  int64                   `json:"id"`
	Title               string                  `json:"title"`
	Description         string                  `json:"description"`
	State               taskdomain.State        `json:"state"`
	ScheduleStartAt     *string                 `json:"schedule_start_at,omitempty"`
	ScheduleEndAt       *string                 `json:"schedule_end_at,omitempty"`
	PeriodicitySettings *periodicitySettingsDTO `json:"periodicity_settings,omitempty"`
	CreatedAt           time.Time               `json:"created_at"`
	UpdatedAt           time.Time               `json:"updated_at"`
}

type completeOccurrenceDTO struct {
	ScheduledFor string `json:"scheduled_for"`
}

func (dto taskMutationDTO) toCreateInput() (taskusecase.CreateInput, error) {
	scheduleStartAt, scheduleEndAt, settings, err := dto.toTaskMutationData()
	if err != nil {
		return taskusecase.CreateInput{}, err
	}

	return taskusecase.CreateInput{
		Title:               dto.Title,
		Description:         dto.Description,
		State:               dto.State,
		ScheduleStartAt:     scheduleStartAt,
		ScheduleEndAt:       scheduleEndAt,
		PeriodicitySettings: settings,
	}, nil
}

func (dto taskMutationDTO) toUpdateInput() (taskusecase.UpdateInput, error) {
	scheduleStartAt, scheduleEndAt, settings, err := dto.toTaskMutationData()
	if err != nil {
		return taskusecase.UpdateInput{}, err
	}

	return taskusecase.UpdateInput{
		Title:               dto.Title,
		Description:         dto.Description,
		State:               dto.State,
		ScheduleStartAt:     scheduleStartAt,
		ScheduleEndAt:       scheduleEndAt,
		PeriodicitySettings: settings,
	}, nil
}

func (dto taskMutationDTO) toTaskMutationData() (*time.Time, *time.Time, *taskdomain.PeriodicitySettings, error) {
	scheduleStartAt, err := parseOptionalCalendarDateTime(dto.ScheduleStartAt)
	if err != nil {
		return nil, nil, nil, err
	}

	scheduleEndAt, err := parseOptionalCalendarDateTime(dto.ScheduleEndAt)
	if err != nil {
		return nil, nil, nil, err
	}

	settings, err := toDomainPeriodicitySettings(dto.PeriodicitySettings)
	if err != nil {
		return nil, nil, nil, err
	}

	return scheduleStartAt, scheduleEndAt, settings, nil
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:                  task.ID,
		Title:               task.Title,
		Description:         task.Description,
		State:               task.State,
		ScheduleStartAt:     formatOptionalCalendarDateTime(task.ScheduleStartAt),
		ScheduleEndAt:       formatOptionalCalendarDateTime(task.ScheduleEndAt),
		PeriodicitySettings: newPeriodicitySettingsDTO(task.PeriodicitySettings),
		CreatedAt:           task.CreatedAt,
		UpdatedAt:           task.UpdatedAt,
	}
}

func toDomainPeriodicitySettings(dto *periodicitySettingsDTO) (*taskdomain.PeriodicitySettings, error) {
	if dto == nil {
		return nil, nil
	}

	specificDateTimes := make([]time.Time, 0, len(dto.SpecificDateTimes))
	for _, rawDateTime := range dto.SpecificDateTimes {
		parsed, err := parseCalendarDateTime(rawDateTime)
		if err != nil {
			return nil, err
		}

		specificDateTimes = append(specificDateTimes, parsed)
	}

	return &taskdomain.PeriodicitySettings{
		Type:              dto.Type,
		EveryNDays:        dto.EveryNDays,
		DaysOfMonth:       dto.DaysOfMonth,
		SpecificDateTimes: specificDateTimes,
		DayParity:         dto.DayParity,
	}, nil
}

func newPeriodicitySettingsDTO(settings *taskdomain.PeriodicitySettings) *periodicitySettingsDTO {
	if settings == nil {
		return nil
	}

	specificDateTimes := make([]string, 0, len(settings.SpecificDateTimes))
	for _, specificDateTime := range settings.SpecificDateTimes {
		specificDateTimes = append(specificDateTimes, formatCalendarDateTime(specificDateTime))
	}

	return &periodicitySettingsDTO{
		Type:              settings.Type,
		EveryNDays:        settings.EveryNDays,
		DaysOfMonth:       settings.DaysOfMonth,
		SpecificDateTimes: specificDateTimes,
		DayParity:         settings.DayParity,
	}
}
