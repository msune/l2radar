package docker

import (
	"testing"
)

func TestRealRunnerImplementsInterface(t *testing.T) {
	var _ Runner = &RealRunner{}
}

func TestMockRunnerRecordsCalls(t *testing.T) {
	m := &MockRunner{}
	_, _, _ = m.Run("ps")
	_, _, _ = m.Run("stop", "foo")

	if len(m.Calls) != 2 {
		t.Fatalf("got %d calls, want 2", len(m.Calls))
	}
	if m.Calls[0][0] != "ps" {
		t.Errorf("call 0 = %v, want [ps]", m.Calls[0])
	}
	if m.Calls[1][0] != "stop" || m.Calls[1][1] != "foo" {
		t.Errorf("call 1 = %v, want [stop foo]", m.Calls[1])
	}
}

func TestMockRunnerAttachedRecordsCalls(t *testing.T) {
	m := &MockRunner{}
	_ = m.RunAttached("exec", "l2radar", "dump")

	if len(m.Calls) != 1 {
		t.Fatalf("got %d calls, want 1", len(m.Calls))
	}
	if m.Calls[0][0] != "exec" {
		t.Errorf("call 0 = %v, want [exec ...]", m.Calls[0])
	}
}
