package gitcli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Executor runs the Git binary.
type Executor interface {
	// Execute runs git with args inside dir and returns stdout and stderr
	// separately. A non-zero exit is reported as *ExitError.
	Execute(ctx context.Context, dir string, args ...string) (stdout, stderr []byte, err error)
}

// CLI is the os/exec implementation of Executor.
type CLI struct {
	bin string
	env []string
}

// NewExecutor returns an executor bound to the git binary on PATH.
func NewExecutor() *CLI { return &CLI{bin: "git"} }

// NewExecutorWithBinary returns an executor bound to an explicit binary.
func NewExecutorWithBinary(bin string) *CLI { return &CLI{bin: bin} }

// Execute implements Executor.
func (c *CLI) Execute(ctx context.Context, dir string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, c.bin, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = c.environment()
	// A nil Stdin is equivalent to /dev/null: git must never wait for input.
	cmd.Stdin = nil

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if ctxErr := ctx.Err(); ctxErr != nil {
		return stdout.Bytes(), stderr.Bytes(), ctxErr
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return stdout.Bytes(), stderr.Bytes(), &ExitError{
				Dir:      dir,
				Args:     args,
				ExitCode: exitErr.ExitCode(),
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				Err:      err,
			}
		}

		var execErr *exec.Error
		if errors.As(err, &execErr) {
			return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("%w: %s", ErrGitNotFound, c.bin)
		}

		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("cannot run git in %q: %w", dir, err)
		}

		return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("running git: %w", err)
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}

func (c *CLI) environment() []string {
	overrides := map[string]string{
		// Never block waiting for credentials or confirmation.
		"GIT_TERMINAL_PROMPT": "0",
		// Stable, machine-parseable output.
		"LC_ALL": "C",
		"LANG":   "C",
		// No pagers or editors inside the TUI.
		"GIT_PAGER":          "cat",
		"PAGER":              "cat",
		"GIT_EDITOR":         "true",
		"GIT_MERGE_AUTOEDIT": "no",
	}

	base := c.env
	if base == nil {
		base = os.Environ()
	}

	out := make([]string, 0, len(base)+len(overrides))
	for _, entry := range base {
		key, _, ok := strings.Cut(entry, "=")
		if ok {
			if _, overridden := overrides[key]; overridden {
				continue
			}
		}
		out = append(out, entry)
	}
	for key, value := range overrides {
		out = append(out, key+"="+value)
	}
	return out
}
