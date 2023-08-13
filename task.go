package jointwork

// Task represents a task of JointWork.
//
// The Stop method can be called several times.
type Task interface {
	Run() error
	Stop() error
}
