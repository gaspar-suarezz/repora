package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// WorkspaceID identifies a workspace. It is an immutable, lowercase slug.
type WorkspaceID string

// RepositoryID identifies a registered repository. It is an immutable, lowercase slug.
type RepositoryID string

const maxIDLength = 64

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

var (
	// ErrEmptyID is returned when an identifier is missing.
	ErrEmptyID = errors.New("id is empty")
	// ErrInvalidID is returned when an identifier is not a valid slug.
	ErrInvalidID = errors.New("id must match ^[a-z0-9]+(-[a-z0-9]+)*$")
	// ErrIDTooLong is returned when an identifier exceeds the length limit.
	ErrIDTooLong = errors.New("id is too long")
)

// NormalizeID converts a human-readable name into a lowercase kebab-case slug.
//
//	"Product X"  -> "product-x"
//	"  API  "    -> "api"
//	"café/serve" -> "caf-serve"
func NormalizeID(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))

	var b strings.Builder
	lastDash := false
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}

	return strings.Trim(b.String(), "-")
}

func validateID(kind, id string) error {
	if id == "" {
		return fmt.Errorf("%s %w", kind, ErrEmptyID)
	}
	if len(id) > maxIDLength {
		return fmt.Errorf("%s %w: %d > %d", kind, ErrIDTooLong, len(id), maxIDLength)
	}
	if !slugPattern.MatchString(id) {
		return fmt.Errorf("%s %w: %q", kind, ErrInvalidID, id)
	}
	return nil
}

// Validate reports whether the workspace identifier is well formed.
func (id WorkspaceID) Validate() error { return validateID("workspace", string(id)) }

// Validate reports whether the repository identifier is well formed.
func (id RepositoryID) Validate() error { return validateID("repository", string(id)) }
