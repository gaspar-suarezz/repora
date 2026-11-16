package gitcli

import (
	"errors"
	"fmt"
	"strings"
)

// ExitError describes a Git invocation that exited with a non-zero status.
type ExitError struct {
	Dir      string
	Args     []string
	ExitCode int
	Stdout   string
	Stderr   string
	Err      error
}

func (e *ExitError) Error() string {
	cmd := "git " + strings.Join(e.Args, " ")
	msg := strings.TrimSpace(e.Stderr)
	if msg == "" {
		msg = strings.TrimSpace(e.Stdout)
	}
	if msg == "" {
		return fmt.Sprintf("%s: exit code %d", cmd, e.ExitCode)
	}
	return fmt.Sprintf("%s: exit code %d: %s", cmd, e.ExitCode, salientLine(msg))
}

// SalientMessage returns the most useful line from stderr: the first fatal or
// error line, otherwise the first non-empty line.
func (e *ExitError) SalientMessage() string {
	msg := strings.TrimSpace(e.Stderr)
	if msg == "" {
		msg = strings.TrimSpace(e.Stdout)
	}
	return salientLine(msg)
}

// Unwrap exposes the underlying error so errors.Is/As keep working.
func (e *ExitError) Unwrap() error { return e.Err }

// StderrContains reports whether stderr contains a case-insensitive substring.
func (e *ExitError) StderrContains(s string) bool {
	return strings.Contains(strings.ToLower(e.Stderr), strings.ToLower(s))
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// salientLine prefers the line that explains the failure.
func salientLine(s string) string {
	lines := strings.Split(s, "\n")
	fallback := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if fallback == "" {
			fallback = line
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "fatal:") || strings.HasPrefix(lower, "error:") {
			return line
		}
	}
	if fallback != "" {
		return fallback
	}
	return firstLine(s)
}

var (
	// ErrGitNotFound indicates the git executable could not be started.
	ErrGitNotFound = errors.New("git executable not found")

	// ErrGitTooOld indicates the installed git is below the supported minimum.
	ErrGitTooOld = errors.New("git version is too old")

	// ErrNotARepository indicates the path is not inside a Git working tree.
	ErrNotARepository = errors.New("not a git repository")

	// ErrBareRepository indicates the path is a bare repository.
	ErrBareRepository = errors.New("bare repositories are not supported")
)

// MinGitVersion is the oldest supported Git release.
const MinGitVersion = "2.30.0"
