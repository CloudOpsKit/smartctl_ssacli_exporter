package command

import (
	"testing"
	"time"
)

func TestRunReturnsOutput(t *testing.T) {
	out, err := Run("sh", "-c", "echo out; echo err >&2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "out\nerr\n" {
		t.Errorf("got %q, want combined stdout and stderr", out)
	}
}

func TestRunTimesOut(t *testing.T) {
	old := Timeout
	Timeout = 100 * time.Millisecond
	defer func() { Timeout = old }()

	start := time.Now()
	if _, err := Run("sleep", "10"); err == nil {
		t.Fatal("expected error for command exceeding timeout")
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Errorf("command was not killed in time: took %v", elapsed)
	}
}
