package jointwork

// Task represents a task of the JointWork.
//
// The Stop method can be called several times.
type Task interface {
	Run() error
	Stop() error
}
