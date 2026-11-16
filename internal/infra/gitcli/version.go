package gitcli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Version returns the installed Git version, for example "2.51.0".
func (c *CLI) Version(ctx context.Context) (string, error) {
	stdout, _, err := c.Execute(ctx, "", "--version")
	if err != nil {
		if IsNotFound(err) {
			return "", fmt.Errorf("%w: install git and ensure it is on PATH", ErrGitNotFound)
		}
		return "", err
	}
	version, err := parseGitVersion(string(stdout))
	if err != nil {
		return "", err
	}
	return version, nil
}

// CheckVersion fails when the given version is older than MinGitVersion.
func CheckVersion(version string) error {
	if compareVersions(version, MinGitVersion) < 0 {
		return fmt.Errorf("%w: found %s, need >= %s", ErrGitTooOld, version, MinGitVersion)
	}
	return nil
}

// IsNotFound reports whether err means the git binary is missing.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrGitNotFound)
}

func parseGitVersion(output string) (string, error) {
	// Examples: "git version 2.51.0", "git version 2.39.3 (Apple Git-145)"
	fields := strings.Fields(strings.TrimSpace(output))
	for _, field := range fields {
		if len(field) == 0 || field[0] < '0' || field[0] > '9' {
			continue
		}
		if strings.Count(field, ".") >= 1 {
			return normalizeVersion(field), nil
		}
	}
	return "", fmt.Errorf("could not parse git version from %q", strings.TrimSpace(output))
}

func normalizeVersion(v string) string {
	v = strings.TrimSuffix(v, ".")
	parts := strings.Split(v, ".")
	for len(parts) < 3 {
		parts = append(parts, "0")
	}
	return strings.Join(parts[:3], ".")
}

func compareVersions(a, b string) int {
	as := strings.Split(normalizeVersion(a), ".")
	bs := strings.Split(normalizeVersion(b), ".")
	for i := 0; i < 3; i++ {
		ai, _ := strconv.Atoi(as[i])
		bi, _ := strconv.Atoi(bs[i])
		switch {
		case ai < bi:
			return -1
		case ai > bi:
			return 1
		}
	}
	return 0
}
