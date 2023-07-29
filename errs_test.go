package jointwork

import (
	"errors"
	"testing"
)

func TestTaskError(t *testing.T) {

	t.Run("TaskError.TaskIndex should return the index received during initialization",
		func(t *testing.T) {
			var (
				wantIndex = 5
				err       = NewTaskError(wantIndex, nil)
				taskIndex = err.TaskIndex()
			)
			if taskIndex != wantIndex {
				t.Errorf("unexpected task index, want '%v' actual '%v'", taskIndex,
					wantIndex)
			}
		})

	t.Run("TaskError.Cause should return the cause received during initialization",
		func(t *testing.T) {
			var (
				wantCause = errors.New("cause")
				err       = NewTaskError(0, wantCause)
				cause     = err.Cause()
			)
			if cause != wantCause {
				t.Errorf("unexpected task index, want '%v' actual '%v'", wantCause,
					cause)
			}
		})

}
