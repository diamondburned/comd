package comd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"libdb.so/comd/internal/cfgtypes"
)

// CommandExecutor defines the interface for executing commands.
type CommandExecutor interface {
	// ExecuteCommand executes the given command.
	ExecuteCommand(ctx context.Context, opts CommandOpts, stdin io.Reader, stdout io.Writer) error
}

// ExecuteOpts is the configuration for executing commands.
type ExecuteOpts struct {
	// Shell is the shell to use for executing commands.
	// It must be a list of strings that represent the command and its arguments.
	// The executor will search for a "%" string in the list and replace it with
	// the command to execute.
	// If not found, the command will be piped to the shell.
	Shell []string `json:"shell"`
	// Timeout is the maximum time to wait for the command to complete.
	Timeout cfgtypes.Duration `json:"timeout"`
}

// NewCommandExecutor creates a new CommandExecutor instance
func NewCommandExecutor(logger *slog.Logger, opts ExecuteOpts) CommandExecutor {
	return defaultCommandExecutor{logger, opts}
}

type defaultCommandExecutor struct {
	logger *slog.Logger
	opts   ExecuteOpts
}

func (e defaultCommandExecutor) ExecuteCommand(ctx context.Context, opts CommandOpts, stdin io.Reader, stdout io.Writer) error {
	var cmd *exec.Cmd
	var isShell bool

	switch command := opts.Command.(type) {
	case ShellCommand:
		isShell = true
		if i := slices.Index(e.opts.Shell, "%"); i != -1 {
			commands := append([]string(nil), e.opts.Shell...)
			commands[i] = string(command)
			cmd = exec.CommandContext(ctx, commands[0], commands[1:]...)
		} else {
			cmd = exec.CommandContext(ctx, e.opts.Shell[0], e.opts.Shell[1:]...)
			cmd.Stdin = strings.NewReader(string(command))
		}

	case DirectCommand:
		cmd = exec.CommandContext(ctx, command[0], command[1:]...)

	default:
		return fmt.Errorf("comd: unsupported command type: %T", command)
	}

	logger := e.logger.With("command", opts.Command)

	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Cancel = func() error { return cancelCommand(cmd, e.logger) }
	cmd.Dir = opts.WorkingDir
	if len(opts.Environment) > 0 {
		cmd.Env = append(os.Environ(), opts.Environment...)
	}

	logger.Info(
		"executing command",
		"is_shell", isShell,
		"working_dir", opts.WorkingDir)
	start := time.Now()

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			logger.Info(
				"command exited with error",
				"duration", time.Since(start),
				"exit_code", exitErr.ExitCode(),
				"stderr", string(exitErr.Stderr))
			return CommandError{
				ExitCode: exitErr.ExitCode(),
			}
		} else {
			logger.Error(
				"command failed",
				"err", err)
			return err
		}
	}

	logger.Info("command completed", "duration", time.Since(start))
	return nil
}

// CommandError is the error type for command execution errors.
type CommandError struct {
	ExitCode int    `json:"exit_code"`
	Stderr   string `json:"stderr"`
}

func (e CommandError) Error() string {
	return fmt.Sprintf("command failed with exit code %d", e.ExitCode)
}

func (e CommandError) IsJSONError() bool { return true }
