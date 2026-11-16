package domain

import (
	"fmt"
	"strings"
)

// forbiddenBranchRunes are characters Git rejects in a ref name.
const forbiddenBranchRunes = " ~^:?*[\\"

// ValidateBranchName checks a branch name against Git's ref rules before any
// command is executed.
func ValidateBranchName(name string) error {
	name = strings.TrimSpace(name)
	switch {
	case name == "":
		return fmt.Errorf("%w: name is empty", ErrInvalidBranch)
	case len(name) > 255:
		return fmt.Errorf("%w: name is too long", ErrInvalidBranch)
	case name == "@":
		return fmt.Errorf("%w: %q is reserved", ErrInvalidBranch, name)
	case strings.HasPrefix(name, "-"):
		return fmt.Errorf("%w: must not start with '-'", ErrInvalidBranch)
	case strings.HasPrefix(name, "/"), strings.HasSuffix(name, "/"):
		return fmt.Errorf("%w: must not start or end with '/'", ErrInvalidBranch)
	case strings.HasSuffix(name, ".lock"):
		return fmt.Errorf("%w: must not end with '.lock'", ErrInvalidBranch)
	case strings.Contains(name, ".."):
		return fmt.Errorf("%w: must not contain '..'", ErrInvalidBranch)
	case strings.Contains(name, "@{"):
		return fmt.Errorf("%w: must not contain '@{'", ErrInvalidBranch)
	case strings.Contains(name, "//"):
		return fmt.Errorf("%w: must not contain '//'", ErrInvalidBranch)
	case strings.HasSuffix(name, "."):
		return fmt.Errorf("%w: must not end with '.'", ErrInvalidBranch)
	}

	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("%w: control characters are not allowed", ErrInvalidBranch)
		}
		if strings.ContainsRune(forbiddenBranchRunes, r) {
			return fmt.Errorf("%w: character %q is not allowed", ErrInvalidBranch, r)
		}
	}

	return nil
}

// ValidateCommitMessage checks that a commit message has content.
func ValidateCommitMessage(message string) error {
	if strings.TrimSpace(message) == "" {
		return ErrEmptyMessage
	}
	return nil
}
