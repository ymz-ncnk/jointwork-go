package jointwork

import (
	"sync"
	"sync/atomic"

	"github.com/ymz-ncnk/multierr-go"
)

// New creates a new JointWork.
func New(tasks []Task) JointWork {
	var stopped uint32 = 0
	return JointWork{tasks: tasks, stopped: &stopped, done: make(chan error, 1),
		mu: &sync.Mutex{}}
}

// JointWork performs several Tasks in different goroutines simultaneously.
//
// If one Task completes with an error, JointWork stops remaining Tasks
// (it will panic if it cannot stop all of them). If Task completes without
// an error, JointWork continues execution. Also JointWork is a Task
// itself.
type JointWork struct {
	tasks        []Task
	stopped      *uint32
	stoppedCount int
	done         chan error
	mu           *sync.Mutex
}

// Run runs JointWork.
//
// Returns a multierr.multiError which contains all Task errors or nil if all
// Tasks were completed without errors.
//
// Run will panic if one of Tasks fails and it can't stop them all.
func (jw *JointWork) Run() (err error) {
	errs := make(chan error, len(jw.tasks))
	if len(jw.tasks) == 0 {
		return
	}
	for i := 0; i < len(jw.tasks); i++ {
		go runTask(jw, i, errs)
	}
	if err := <-jw.done; err != nil {
		panic(err)
	}
	return multierr.New(toSlice(errs))
}

// Stop stops JointWork, it stops all Tasks.
//
// Returns a multierr.multiError or nil if all Tasks were stopped without
// errors.
func (jw *JointWork) Stop() (err error) {
	if swapped := atomic.CompareAndSwapUint32(jw.stopped, 0, 1); !swapped {
		return
	}
	errs := []error{}
	for _, task := range jw.tasks {
		if err = task.Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	return multierr.New(errs)
}

func runTask(jw *JointWork, i int, errs chan<- error) {
	defer incrementStopped(jw)
	var (
		task = jw.tasks[i]
		err  = task.Run()
	)
	if err != nil {
		errs <- NewTaskError(i, err)
		if err := jw.Stop(); err != nil {
			jw.done <- err
		}
	}
}

func incrementStopped(jw *JointWork) {
	jw.mu.Lock()
	jw.stoppedCount += 1
	if jw.stoppedCount == len(jw.tasks) {
		close(jw.done)
	}
	jw.mu.Unlock()
}

func toSlice(errs chan error) (es []error) {
	close(errs)
	for err := range errs {
		es = append(es, err)
	}
	return
}
