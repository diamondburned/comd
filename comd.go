// Package comd implements a Command Daemon that listens for HTTP requests to
// execute commands.
package comd

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Config struct {
	// BasePath is the base HTTP path for the handler.
	BasePath string `json:"base_path"`
	// Execute is the configuration for executing commands.
	Execute ExecuteOpts `json:"execute"`
	// Commands is a map of command names to Commands.
	Commands map[string]CommandOpts `json:"commands"`
}

// CommandOpts is the configuration for executing a command.
//
// When parsed from the config, it may either be:
// - a string which describes a command to be executed using the shell, or
// - an array of strings which describes a command to be executed directly, or
// - the CommandOpts struct itself, which allows for more control over the command execution.
type CommandOpts struct {
	// Command is the command to execute.
	Command Command `json:"command"`
	// WorkingDir is the working directory for the command.
	// If empty, the command will be executed in the current working directory.
	WorkingDir string `json:"working_dir"`
	// Environment is a list of environment variables to set for the command.
	Environment []string `json:"environment"`
}

// Command describes a command to be executed.
// It may either be [ShellCommand] or [DirectCommand].
type Command interface{ commandArgs() }

// ShellCommand is a command that will be executed using the shell.
type ShellCommand string

// DirectCommand is a command that will be executed directly.
type DirectCommand []string

func (ShellCommand) commandArgs()  {}
func (DirectCommand) commandArgs() {}

func (c *CommandOpts) UnmarshalJSON(blob []byte) error {
	switch {
	case bytes.HasPrefix(blob, []byte(`{`)):
		type RawOpts CommandOpts
		var raw struct {
			Command json.RawMessage `json:"command"`
			RawOpts
		}
		if err := json.Unmarshal(blob, &raw); err != nil {
			return err
		}

		*c = CommandOpts(raw.RawOpts)
		if err := c.unmarshalArgs(raw.Command); err != nil {
			return fmt.Errorf("comd: invalid command args: %v", err)
		}

		return nil
	default:
		*c = CommandOpts{}
		return c.unmarshalArgs(blob)
	}
}

func (c *CommandOpts) unmarshalArgs(blob []byte) error {
	switch {
	case bytes.HasPrefix(blob, []byte(`"`)):
		var cmd ShellCommand
		if err := json.Unmarshal(blob, &cmd); err != nil {
			return fmt.Errorf("comd: invalid command string: %v", err)
		}

		c.Command = cmd
		return nil

	case bytes.HasPrefix(blob, []byte(`[`)):
		var cmd DirectCommand
		if err := json.Unmarshal(blob, &cmd); err != nil {
			return fmt.Errorf("comd: invalid command array: %v", err)
		}
		if len(cmd) == 0 {
			return fmt.Errorf("comd: invalid command array: empty")
		}

		c.Command = cmd
		return nil
	}

	return fmt.Errorf("comd: invalid command json: %s", blob)
}
