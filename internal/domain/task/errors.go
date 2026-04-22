package task

import "errors"

var ErrNotFound = errors.New("task not found")
var ErrOccurrenceAlreadyCompleted = errors.New("task occurrence already completed")
