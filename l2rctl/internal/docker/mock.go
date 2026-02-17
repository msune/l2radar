package docker

// MockRunner records calls for testing.
type MockRunner struct {
	Calls    [][]string
	StdoutFn func(args []string) string
	StderrFn func(args []string) string
	ErrFn    func(args []string) error
}

var _ Runner = &MockRunner{}

func (m *MockRunner) Run(args ...string) (string, string, error) {
	m.Calls = append(m.Calls, args)
	var stdout, stderr string
	var err error
	if m.StdoutFn != nil {
		stdout = m.StdoutFn(args)
	}
	if m.StderrFn != nil {
		stderr = m.StderrFn(args)
	}
	if m.ErrFn != nil {
		err = m.ErrFn(args)
	}
	return stdout, stderr, err
}

func (m *MockRunner) RunAttached(args ...string) error {
	m.Calls = append(m.Calls, args)
	if m.ErrFn != nil {
		return m.ErrFn(args)
	}
	return nil
}
