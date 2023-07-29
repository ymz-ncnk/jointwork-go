package mock

import "github.com/ymz-ncnk/mok"

func NewTask() Task {
	return Task{
		Mock: mok.New("Task"),
	}
}

type Task struct {
	*mok.Mock
}

func (m Task) RegisterRun(fn func() (err error)) Task {
	m.Register("Run", fn)
	return m
}

func (m Task) RegisterStop(fn func() (err error)) Task {
	m.Register("Stop", fn)
	return m
}

func (m Task) Run() (err error) {
	result, err := m.Call("Run")
	if err != nil {
		panic(err)
	}
	err, _ = result[0].(error)
	return
}

func (m Task) Stop() (err error) {
	result, err := m.Call("Stop")
	if err != nil {
		panic(err)
	}
	err, _ = result[0].(error)
	return
}
