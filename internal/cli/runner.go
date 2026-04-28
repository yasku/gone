package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type Runner struct {
	timeout time.Duration
	env     []string
}

func NewRunner(timeout time.Duration) *Runner {
	return &Runner{
		timeout: timeout,
		env:     []string{},
	}
}

func (r *Runner) WithEnv(env []string) *Runner {
	r.env = env
	return r
}

func (r *Runner) ExecJSON(cmd string, args []string, output interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmdExec := exec.CommandContext(ctx, cmd, args...)
	cmdExec.Env = r.env

	stdout, err := cmdExec.StdoutPipe()
	if err != nil {
		return fmt.Errorf("StdoutPipe: %w", err)
	}

	stderr := &bytes.Buffer{}
	cmdExec.Stderr = stderr

	if err := cmdExec.Start(); err != nil {
		return fmt.Errorf("Start: %w", err)
	}

	dec := json.NewDecoder(stdout)
	if err := dec.Decode(output); err != nil {
		if err := cmdExec.Wait(); err != nil {
			return fmt.Errorf("exec %s: %v: %s", cmd, err, stderr.String())
		}
		return fmt.Errorf("Decode JSON from %s: %w", cmd, err)
	}

	if err := cmdExec.Wait(); err != nil {
		return fmt.Errorf("exec %s: %v: %s", cmd, err, stderr.String())
	}

	return nil
}

func (r *Runner) ExecStream(cmd string, args []string, handler func(line []byte) bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmdExec := exec.CommandContext(ctx, cmd, args...)
	cmdExec.Env = r.env

	stdout, err := cmdExec.StdoutPipe()
	if err != nil {
		return fmt.Errorf("StdoutPipe: %w", err)
	}

	stderr := &bytes.Buffer{}
	cmdExec.Stderr = stderr

	if err := cmdExec.Start(); err != nil {
		return fmt.Errorf("Start: %w", err)
	}

	scanner := &streamScanner{r: stdout}
	for scanner.Scan() {
		line := scanner.Bytes()
		if !handler(line) {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream: %w", err)
	}

	if err := cmdExec.Wait(); err != nil {
		return fmt.Errorf("exec %s: %v: %s", cmd, err, stderr.String())
	}

	return nil
}

func (r *Runner) ExecSimple(cmd string, args []string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmdExec := exec.CommandContext(ctx, cmd, args...)
	cmdExec.Env = r.env

	return cmdExec.CombinedOutput()
}

type ToolError struct {
	Tool     string
	Args     []string
	ExitCode int
	Stderr   string
}

func (e *ToolError) Error() string {
	return fmt.Sprintf("cli: %s %v failed (exit %d): %s", e.Tool, e.Args, e.ExitCode, e.Stderr)
}
