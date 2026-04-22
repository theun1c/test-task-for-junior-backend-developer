package task

import "time"

const (
	StateActive   State = "active"
	StatePaused   State = "paused"
	StateArchived State = "archived"
)

type Task struct {
	ID                  int64                `json:"id"`
	Title               string               `json:"title"`
	Description         string               `json:"description"`
	State               State                `json:"state"`
	ScheduleStartAt     *time.Time           `json:"schedule_start_at,omitempty"`
	ScheduleEndAt       *time.Time           `json:"schedule_end_at,omitempty"`
	PeriodicitySettings *PeriodicitySettings `json:"periodicity_settings,omitempty"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
}

type State string

func (s State) Valid() bool {
	switch s {
	case StateActive, StatePaused, StateArchived:
		return true
	default:
		return false
	}
}
