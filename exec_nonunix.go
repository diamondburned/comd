//go:build !unix

package comd

import (
	"log/slog"
	"os/exec"
)

func cancelCommand(cmd *exec.Cmd, logger *slog.Logger) error {
	return cmd.Process.Kill()
}
