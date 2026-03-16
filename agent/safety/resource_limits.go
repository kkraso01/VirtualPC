package safety

import (
	"fmt"
	"time"
)

type ResourceLimits struct {
	MaxCommandsPerSession int
	MaxRuntime            time.Duration
	MaxDiskUsageMB        int
	MaxMemoryMB           int
	MaxProcesses          int
	MaxIterations         int
	MaxFailures           int
	MaxRepeatedCommand    int
}

func (r ResourceLimits) Validate(commands int, started time.Time, iterations int, failures int, repeated int) error {
	if commands > r.MaxCommandsPerSession {
		return fmt.Errorf("session command limit exceeded: %d > %d", commands, r.MaxCommandsPerSession)
	}
	if time.Since(started) > r.MaxRuntime {
		return fmt.Errorf("session runtime limit exceeded: %s > %s", time.Since(started).Round(time.Second), r.MaxRuntime)
	}
	if iterations > r.MaxIterations {
		return fmt.Errorf("iteration limit exceeded: %d > %d", iterations, r.MaxIterations)
	}
	if failures > r.MaxFailures {
		return fmt.Errorf("tool failure threshold exceeded: %d > %d", failures, r.MaxFailures)
	}
	if repeated > r.MaxRepeatedCommand {
		return fmt.Errorf("repeated command threshold exceeded: %d > %d", repeated, r.MaxRepeatedCommand)
	}
	return nil
}
