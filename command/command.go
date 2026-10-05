// Package command runs external tools (ssacli, smartctl) with a timeout,
// so a hung controller call cannot block a scrape forever.
package command

import (
	"context"
	"os/exec"
	"time"
)

// Timeout limits how long a single command may run.
var Timeout = 30 * time.Second

// Run executes the command and returns its combined stdout and stderr.
func Run(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	// Don't wait forever for output pipes held open by orphaned child processes
	cmd.WaitDelay = time.Second
	return cmd.CombinedOutput()
}
