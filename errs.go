package jointwork

import (
	"fmt"
)

// NewTaskError creates a new TaskError.
func NewTaskError(index int, cause error) *TaskError {
	return &TaskError{index, cause}
}

// TaskError represents a Task error.
type TaskError struct {
	taskIndex int
	cause     error
}

// TaskIndex is a sequential Task index.
func (err *TaskError) TaskIndex() int {
	return err.taskIndex
}

// Cause returns the error that caused the Task to stop.
func (err *TaskError) Cause() error {
	return err.cause
}

func (err *TaskError) Error() string {
	return fmt.Sprintf("%d task failed, cause: %s", err.taskIndex, err.cause)
}
