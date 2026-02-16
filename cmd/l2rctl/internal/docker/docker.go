package docker

import (
	"bytes"
	"os"
	"os/exec"
)

// Runner abstracts Docker CLI execution for testability.
type Runner interface {
	Run(args ...string) (stdout, stderr string, err error)
	RunAttached(args ...string) error
}

// RealRunner executes docker commands via os/exec.
type RealRunner struct{}

func (r *RealRunner) Run(args ...string) (string, string, error) {
	cmd := exec.Command("docker", args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

func (r *RealRunner) RunAttached(args ...string) error {
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
