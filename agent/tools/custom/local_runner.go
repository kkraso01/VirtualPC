package custom

import (
	"context"
	"os/exec"
	"time"
)

func RunLocal(ctx context.Context, cmdPath string, args []string, timeoutSeconds int) (Result, error) {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 10
	}
	cctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(cctx, cmdPath, args...)
	b, err := cmd.CombinedOutput()
	return Result{Output: string(b)}, err
}
