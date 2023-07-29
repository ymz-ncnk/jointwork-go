package jointwork

import (
	"errors"
	"testing"
	"time"

	"github.com/ymz-ncnk/jointwork-go/testdata/mock"
	"github.com/ymz-ncnk/mok"
	"github.com/ymz-ncnk/multierr-go"
)

func TestJointWork(t *testing.T) {

	t.Run("If JointWork has no Tasks, Run should return immediately",
		func(t *testing.T) {
			var (
				jw  = New([]Task{})
				err = jw.Run()
			)
			if err != nil {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
		})

	t.Run("JointWork should be able to run several Tasks", func(t *testing.T) {
		var (
			task1 = mock.NewTask().RegisterRun(
				func() error { time.Sleep(50 * time.Millisecond); return nil },
			)
			task2 = mock.NewTask().RegisterRun(
				func() error { time.Sleep(10 * time.Millisecond); return nil },
			)
			tasks = []Task{task1, task2}
			mocks = []*mok.Mock{task1.Mock, task2.Mock}
			jw    = New(tasks)
			err   = jw.Run()
		)
		if err != nil {
			t.Errorf("unexpected err, want '%v' actual '%v'", nil, err)
		}
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("If a task completes with an error, Run should return it and stop all other Tasks",
		func(t *testing.T) {
			var (
				task1Err = errors.New("first task error")
				task1    = mock.NewTask().RegisterRun(
					func() error { time.Sleep(50 * time.Millisecond); return task1Err },
				).RegisterStop(
					func() (err error) { return nil },
				)
				task2 = mock.NewTask().RegisterRun(
					func() error { time.Sleep(50 * time.Millisecond); return nil },
				).RegisterStop(
					func() (err error) { return nil },
				)
				wantErr = multierr.New([]error{NewTaskError(0, task1Err)})
				tasks   = []Task{task1, task2}
				mocks   = []*mok.Mock{task1.Mock, task2.Mock}
				jw      = New(tasks)
				err     = jw.Run()
			)
			testErr(err, wantErr, t)
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("Stop should stop all Tasks", func(t *testing.T) {
		var (
			task1 = func() mock.Task {
				done := make(chan struct{})
				task := mock.NewTask().RegisterRun(
					func() error { <-done; return nil },
				).RegisterStop(
					func() (err error) { close(done); return nil },
				)
				return task
			}()

			task2 = func() mock.Task {
				done := make(chan struct{})
				task := mock.NewTask().RegisterRun(
					func() error { <-done; return nil },
				).RegisterStop(
					func() (err error) { close(done); return nil },
				)
				return task
			}()
			tasks = []Task{task1, task2}
			mocks = []*mok.Mock{task1.Mock, task2.Mock}
			jw    = New(tasks)
		)
		go func() {
			err := jw.Stop()
			if err != nil {
				t.Error(err)
			}
		}()
		err := jw.Run()
		if err != nil {
			t.Errorf("unexpected err, want '%v' actual '%v'", nil, err)
		}
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("Run shlould return all Task errors", func(t *testing.T) {
		var (
			Task1Err = errors.New("Task1 error")
			Task1    = func() mock.Task {
				done := make(chan struct{})
				Task := mock.NewTask().RegisterRun(
					func() error { <-done; return Task1Err },
				).RegisterStop(
					func() (err error) { close(done); return nil },
				)
				return Task
			}()

			Task2Err = errors.New("Task2 error")
			Task2    = func() mock.Task {
				done := make(chan struct{})
				Task := mock.NewTask().RegisterRun(
					func() error { <-done; return Task2Err },
				).RegisterStop(
					func() (err error) { close(done); return nil },
				)
				return Task
			}()
			wantErr = multierr.New([]error{
				NewTaskError(0, Task1Err),
				NewTaskError(1, Task2Err),
			})
			tasks = []Task{Task1, Task2}
			mocks = []*mok.Mock{Task1.Mock, Task2.Mock}
			jw    = New(tasks)
		)
		go func() {
			time.Sleep(100 * time.Millisecond)
			err := jw.Stop()
			if err != nil {
				t.Error(err)
			}
		}()
		err := jw.Run()
		testErr(err, wantErr, t)
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("If several Tasks stop with errors, Stop should return them all",
		func(t *testing.T) {
			var (
				task1StopErr = errors.New("task1 stop error")
				task1        = func() mock.Task {
					done := make(chan struct{})
					task := mock.NewTask().RegisterRun(
						func() error { <-done; return nil },
					).RegisterStop(
						func() (err error) { close(done); return task1StopErr },
					)
					return task
				}()

				task2StopErr = errors.New("task2 stop error")
				task2        = func() mock.Task {
					done := make(chan struct{})
					task := mock.NewTask().RegisterRun(
						func() error { <-done; return nil },
					).RegisterStop(
						func() (err error) { close(done); return task2StopErr },
					)
					return task
				}()
				wantErr = multierr.New([]error{task1StopErr, task2StopErr})
				tasks   = []Task{task1, task2}
				mocks   = []*mok.Mock{task1.Mock, task2.Mock}
				jw      = New(tasks)
			)
			go func() {
				time.Sleep(100 * time.Millisecond)
				err := jw.Stop()
				if !err.(interface{ Similar(ae error) bool }).Similar(wantErr) {
					t.Error(err)
				}
				if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
					t.Error(infomap)
				}
			}()
			jw.Run()
		})

	t.Run("JoinWork should panic, if one of the Tasks fails, and it cann't stop all of them",
		func(t *testing.T) {
			var (
				wantErr = errors.New("task2 error")
				task1   = mock.NewTask().RegisterRun(
					func() error { return errors.New("task1 error") },
				).RegisterStop(
					func() (err error) { return nil },
				)
				task2 = func() mock.Task {
					done := make(chan struct{})
					task := mock.NewTask().RegisterRun(
						func() error { <-done; return nil },
					).RegisterStop(
						func() (err error) { close(done); return wantErr },
					)
					return task
				}()
				tasks = []Task{task1, task2}
				mocks = []*mok.Mock{task1.Mock, task2.Mock}
				jw    = New(tasks)
			)
			defer func() {
				err, ok := recover().(error)
				if !ok {
					t.Errorf("recover should return an error")
				}
				if err.Error() != wantErr.Error() {
					t.Errorf("unexpected error, want '%v' actual '%v'", err, wantErr)
				}
				if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
					t.Error(infomap)
				}
			}()
			jw.Run()
		})

}

func testErr(err, wantErr error, t *testing.T) {
	if wantErr == nil && err == nil {
		return
	}
	if (wantErr == nil && err != nil) || (wantErr != nil && err == nil) {
		t.Errorf("unexpected err, want '%v' actual '%v'", err, wantErr)
	}
	jwErr := err.(interface{ Similar(ae error) bool })
	if !jwErr.Similar(wantErr) {
		t.Errorf("unexpected err, want '%v' actual '%v'", err, wantErr)
	}
}
