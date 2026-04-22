package task

import "time"

type PeriodicityType string

const (
	PeriodicityTypeDaily          PeriodicityType = "daily"
	PeriodicityTypeMonthlyDays    PeriodicityType = "monthly_days"
	PeriodicityTypeSpecificDates  PeriodicityType = "specific_datetimes"
	PeriodicityTypeMonthDayParity PeriodicityType = "month_day_parity"
)

type DayParity string

const (
	DayParityEven DayParity = "even"
	DayParityOdd  DayParity = "odd"
)

type PeriodicitySettings struct {
	Type              PeriodicityType `json:"type"`
	EveryNDays        *int            `json:"every_n_days,omitempty"`
	DaysOfMonth       []int           `json:"days_of_month,omitempty"`
	SpecificDateTimes []time.Time     `json:"specific_datetimes,omitempty"`
	DayParity         *DayParity      `json:"day_parity,omitempty"`
}

func (t PeriodicityType) Valid() bool {
	switch t {
	case PeriodicityTypeDaily, PeriodicityTypeMonthlyDays, PeriodicityTypeSpecificDates, PeriodicityTypeMonthDayParity:
		return true
	default:
		return false
	}
}

func (p DayParity) Valid() bool {
	switch p {
	case DayParityEven, DayParityOdd:
		return true
	default:
		return false
	}
}
