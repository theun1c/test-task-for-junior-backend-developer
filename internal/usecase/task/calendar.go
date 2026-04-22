package task

import (
	"fmt"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func validateSchedule(scheduleStartAt, scheduleEndAt *time.Time, settings *taskdomain.PeriodicitySettings) error {
	if settings == nil {
		if scheduleEndAt != nil {
			return fmt.Errorf("%w: schedule_end_at requires periodicity settings", ErrInvalidInput)
		}

		return nil
	}

	if scheduleStartAt == nil {
		return fmt.Errorf("%w: schedule_start_at is required for periodic task", ErrInvalidInput)
	}

	if scheduleEndAt != nil && scheduleEndAt.Before(*scheduleStartAt) {
		return fmt.Errorf("%w: schedule_end_at must be greater than or equal to schedule_start_at", ErrInvalidInput)
	}

	if !settings.Type.Valid() {
		return fmt.Errorf("%w: invalid periodicity type", ErrInvalidInput)
	}

	switch settings.Type {
	case taskdomain.PeriodicityTypeDaily:
		if settings.EveryNDays == nil || *settings.EveryNDays <= 0 {
			return fmt.Errorf("%w: every_n_days must be greater than zero", ErrInvalidInput)
		}
	case taskdomain.PeriodicityTypeMonthlyDays:
		if len(settings.DaysOfMonth) == 0 {
			return fmt.Errorf("%w: days_of_month is required", ErrInvalidInput)
		}

		seenDays := make(map[int]struct{}, len(settings.DaysOfMonth))
		for _, day := range settings.DaysOfMonth {
			if day < 1 || day > 30 {
				return fmt.Errorf("%w: days_of_month values must be in range 1..30", ErrInvalidInput)
			}
			if _, exists := seenDays[day]; exists {
				return fmt.Errorf("%w: duplicate days_of_month values are not allowed", ErrInvalidInput)
			}
			seenDays[day] = struct{}{}
		}
	case taskdomain.PeriodicityTypeSpecificDates:
		if len(settings.SpecificDateTimes) == 0 {
			return fmt.Errorf("%w: specific_datetimes is required", ErrInvalidInput)
		}

		for _, specificDateTime := range settings.SpecificDateTimes {
			if specificDateTime.Before(*scheduleStartAt) {
				return fmt.Errorf("%w: specific_datetime must be greater than or equal to schedule_start_at", ErrInvalidInput)
			}

			if scheduleEndAt != nil && specificDateTime.After(*scheduleEndAt) {
				return fmt.Errorf("%w: specific_datetime must be less than or equal to schedule_end_at", ErrInvalidInput)
			}
		}
	case taskdomain.PeriodicityTypeMonthDayParity:
		if settings.DayParity == nil || !settings.DayParity.Valid() {
			return fmt.Errorf("%w: valid day_parity is required", ErrInvalidInput)
		}
	}

	return nil
}

func matchesCalendarDateTime(task *taskdomain.Task, at time.Time) bool {
	if task.State != taskdomain.StateActive {
		return false
	}

	if task.PeriodicitySettings == nil {
		return matchesOneTimeTask(task, at)
	}

	if task.ScheduleStartAt == nil {
		return false
	}

	if !isWithinScheduleRange(task.ScheduleStartAt, task.ScheduleEndAt, at) {
		return false
	}

	switch task.PeriodicitySettings.Type {
	case taskdomain.PeriodicityTypeDaily:
		return matchesDaily(task, at)
	case taskdomain.PeriodicityTypeMonthlyDays:
		return matchesMonthlyDays(task, at)
	case taskdomain.PeriodicityTypeSpecificDates:
		return matchesSpecificDateTimes(task, at)
	case taskdomain.PeriodicityTypeMonthDayParity:
		return matchesMonthDayParity(task, at)
	default:
		return false
	}
}

func matchesOneTimeTask(task *taskdomain.Task, at time.Time) bool {
	if task.ScheduleStartAt == nil {
		return false
	}

	return toLocalMinute(*task.ScheduleStartAt).Equal(toLocalMinute(at))
}

func isWithinScheduleRange(scheduleStartAt, scheduleEndAt *time.Time, at time.Time) bool {
	if scheduleStartAt == nil {
		return false
	}

	localAt := toLocalMinute(at)
	localStart := toLocalMinute(*scheduleStartAt)
	if localAt.Before(localStart) {
		return false
	}

	if scheduleEndAt != nil {
		localEnd := toLocalMinute(*scheduleEndAt)
		if localAt.After(localEnd) {
			return false
		}
	}

	return true
}

func matchesDaily(task *taskdomain.Task, at time.Time) bool {
	if task.ScheduleStartAt == nil || task.PeriodicitySettings == nil || task.PeriodicitySettings.EveryNDays == nil {
		return false
	}

	localStart := toLocalMinute(*task.ScheduleStartAt)
	localAt := toLocalMinute(at)
	if !sameClockTime(localStart, localAt) {
		return false
	}

	startDate := dateOnly(localStart)
	atDate := dateOnly(localAt)
	dayDiff := int(atDate.Sub(startDate).Hours() / 24)
	if dayDiff < 0 {
		return false
	}

	return dayDiff%*task.PeriodicitySettings.EveryNDays == 0
}

func matchesMonthlyDays(task *taskdomain.Task, at time.Time) bool {
	if task.ScheduleStartAt == nil || task.PeriodicitySettings == nil {
		return false
	}

	localStart := toLocalMinute(*task.ScheduleStartAt)
	localAt := toLocalMinute(at)
	if !sameClockTime(localStart, localAt) {
		return false
	}

	for _, day := range task.PeriodicitySettings.DaysOfMonth {
		if localAt.Day() == day {
			return true
		}
	}

	return false
}

func matchesSpecificDateTimes(task *taskdomain.Task, at time.Time) bool {
	if task.PeriodicitySettings == nil {
		return false
	}

	localAt := toLocalMinute(at)
	for _, specificDateTime := range task.PeriodicitySettings.SpecificDateTimes {
		if toLocalMinute(specificDateTime).Equal(localAt) {
			return true
		}
	}

	return false
}

func matchesMonthDayParity(task *taskdomain.Task, at time.Time) bool {
	if task.ScheduleStartAt == nil || task.PeriodicitySettings == nil || task.PeriodicitySettings.DayParity == nil {
		return false
	}

	localStart := toLocalMinute(*task.ScheduleStartAt)
	localAt := toLocalMinute(at)
	if !sameClockTime(localStart, localAt) {
		return false
	}

	isEvenDay := localAt.Day()%2 == 0

	switch *task.PeriodicitySettings.DayParity {
	case taskdomain.DayParityEven:
		return isEvenDay
	case taskdomain.DayParityOdd:
		return !isEvenDay
	default:
		return false
	}
}

func toLocalMinute(t time.Time) time.Time {
	localTime := t.In(time.Local)

	return time.Date(
		localTime.Year(),
		localTime.Month(),
		localTime.Day(),
		localTime.Hour(),
		localTime.Minute(),
		0,
		0,
		time.Local,
	)
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

func sameClockTime(a, b time.Time) bool {
	return a.Hour() == b.Hour() && a.Minute() == b.Minute()
}
