//go:build unix

package comd

import (
	"log/slog"
	"os/exec"
	"time"

	"golang.org/x/sys/unix"
)

func cancelCommand(cmd *exec.Cmd, logger *slog.Logger) error {
	wait := make(chan error, 1)
	go func() {
		wait <- cmd.Wait()
	}()

	// SIGTERM for 2 seconds, then SIGKILL.
	if err := cmd.Process.Signal(unix.SIGTERM); err == nil {
		termTimeout := time.NewTimer(2 * time.Second)
		defer termTimeout.Stop()

		select {
		case <-termTimeout.C:
			logger.Debug(
				"timeout waiting for command to exit after SIGTERM, sending SIGKILL",
				"signal", "SIGKILL",
				"timeout", "2s")

		case err := <-wait:
			logger.Debug(
				"command exited after SIGTERM",
				"signal", "SIGTERM")
			return err
		}
	} else {
		logger.Warn(
			"failing to send SIGTERM to command",
			"signal", "SIGTERM",
			"err", err)
	}

	if err := cmd.Process.Kill(); err != nil {
		logger.Warn(
			"failing to send SIGKILL to command",
			"signal", "SIGKILL",
			"err", err)
	}

	return <-wait
}
