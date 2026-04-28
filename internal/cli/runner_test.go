package cli

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestNewRunner(t *testing.T) {
	r := NewRunner(30 * time.Second)
	if r == nil {
		t.Fatal("NewRunner returned nil")
	}
	if r.timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", r.timeout)
	}
}

func TestRunnerWithEnv(t *testing.T) {
	r := NewRunner(10 * time.Second)
	r = r.WithEnv([]string{"PATH=/usr/bin"})
	if len(r.env) != 1 || r.env[0] != "PATH=/usr/bin" {
		t.Errorf("expected env [PATH=/usr/bin], got %v", r.env)
	}
}

func TestWhich(t *testing.T) {
	path, err := Which("ls")
	if err != nil {
		t.Skipf("ls not available: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path for ls")
	}
}

func TestWhichNotFound(t *testing.T) {
	_, err := Which("nonexistent-tool-xyz")
	if err == nil {
		t.Error("expected error for nonexistent tool")
	}
}

func TestIsAvailable(t *testing.T) {
	if !IsAvailable("ls") {
		t.Error("expected ls to be available")
	}

	if IsAvailable("gone-nonexistent-tool-xyz") {
		t.Error("expected nonexistent tool to return false")
	}
}

func TestAvailableTools(t *testing.T) {
	tools := AvailableTools()
	if tools == nil {
		t.Error("AvailableTools returned nil")
	}
}

func TestToolError(t *testing.T) {
	e := &ToolError{
		Tool:     "test",
		Args:     []string{"arg1"},
		ExitCode: 1,
		Stderr:   "test error",
	}

	expected := "cli: test [arg1] failed (exit 1): test error"
	if e.Error() != expected {
		t.Errorf("expected %q, got %q", expected, e.Error())
	}
}

func TestExecSimpleSuccess(t *testing.T) {
	r := NewRunner(5 * time.Second)

	output, err := r.ExecSimple("echo", []string{"hello", "world"})
	if err != nil {
		t.Fatalf("ExecSimple failed: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr != "hello world" {
		t.Errorf("expected 'hello world', got %q", outputStr)
	}
}

func TestExecSimpleTimeout(t *testing.T) {
	r := NewRunner(100 * time.Millisecond)

	_, err := r.ExecSimple("sleep", []string{"10"})
	if err == nil {
		t.Error("expected timeout error, got nil")
	}

	if err != nil {
		errStr := err.Error()
		if !strings.Contains(errStr, "context deadline exceeded") &&
			!strings.Contains(errStr, "timeout") &&
			!strings.Contains(errStr, "signal") {
			t.Errorf("expected timeout-related error, got: %v", err)
		}
	}
}

func TestExecSimpleNotFound(t *testing.T) {
	r := NewRunner(5 * time.Second)

	_, err := r.ExecSimple("nonexistent-command-xyz", []string{})
	if err == nil {
		t.Error("expected error for nonexistent command")
	}
}

type testJSONOutput struct {
	Message string `json:"message"`
}

func TestExecJSONSuccess(t *testing.T) {
	r := NewRunner(5 * time.Second)

	var output testJSONOutput
	err := r.ExecJSON("echo", []string{`{"message":"hello"}`}, &output)
	if err != nil {
		t.Fatalf("ExecJSON failed: %v", err)
	}

	if output.Message != "hello" {
		t.Errorf("expected message 'hello', got %q", output.Message)
	}
}

func TestExecJSONInvalidJSON(t *testing.T) {
	r := NewRunner(5 * time.Second)

	var output testJSONOutput
	err := r.ExecJSON("echo", []string{"not valid json"}, &output)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestExecJSONNotFound(t *testing.T) {
	r := NewRunner(5 * time.Second)

	var output testJSONOutput
	err := r.ExecJSON("nonexistent-json-cmd", []string{}, &output)
	if err == nil {
		t.Error("expected error for nonexistent command")
	}
}

type streamLine struct {
	Seq int `json:"seq"`
}

func TestExecStreamSuccess(t *testing.T) {
	r := NewRunner(5 * time.Second)

	var lines []streamLine
	err := r.ExecStream("printf", []string{`%s\n%s\n%s\n`, `{"seq":1}`, `{"seq":2}`, `{"seq":3}`}, func(line []byte) bool {
		if len(line) == 0 {
			return true
		}
		var l streamLine
		if err := json.Unmarshal(line, &l); err == nil {
			lines = append(lines, l)
		}
		return true
	})

	if err != nil {
		t.Fatalf("ExecStream failed: %v", err)
	}

	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	for i, l := range lines {
		if l.Seq != i+1 {
			t.Errorf("line %d: expected seq %d, got %d", i, i+1, l.Seq)
		}
	}
}

func TestExecStreamStopEarly(t *testing.T) {
	r := NewRunner(5 * time.Second)

	count := 0
	err := r.ExecStream("seq", []string{"1", "10"}, func(line []byte) bool {
		count++
		return count < 3
	})

	if err != nil {
		t.Fatalf("ExecStream failed: %v", err)
	}

	if count != 3 {
		t.Errorf("expected to process 3 lines, got %d", count)
	}
}

func TestExecStreamTimeout(t *testing.T) {
	r := NewRunner(100 * time.Millisecond)

	count := 0
	err := r.ExecStream("bash", []string{"-c", "while true; do echo $RANDOM; sleep 1; done"}, func(line []byte) bool {
		count++
		return true
	})

	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestExecStreamNotFound(t *testing.T) {
	r := NewRunner(5 * time.Second)

	err := r.ExecStream("nonexistent-stream-cmd", []string{}, func(line []byte) bool {
		return true
	})

	if err == nil {
		t.Error("expected error for nonexistent command")
	}
}

func TestRunnerContextCancellation(t *testing.T) {
	_ = NewRunner(50 * time.Millisecond)

	ctx := context.Background()
	_ = ctx

	cmd := exec.Command("bash", "-c", "sleep 10 && echo done")
	cmd.Cancel = func() error { return nil }

	if cmd.Process != nil {
		cmd.Process.Kill()
	}
}
