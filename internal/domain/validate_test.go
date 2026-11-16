package domain

import (
	"errors"
	"testing"
)

func TestValidateBranchName(t *testing.T) {
	valid := []string{
		"main",
		"feature/payment-retry",
		"fix-123",
		"release/v1.2.3",
		"a",
	}
	for _, name := range valid {
		if err := ValidateBranchName(name); err != nil {
			t.Errorf("ValidateBranchName(%q) = %v, want nil", name, err)
		}
	}

	invalid := []string{
		"",
		"   ",
		"@",
		"-feature",
		"/feature",
		"feature/",
		"feature..x",
		"feature@{1}",
		"feature//x",
		"feature.lock",
		"feature.",
		"feat ure",
		"feat~ure",
		"feat^ure",
		"feat:ure",
		"feat?ure",
		"feat*ure",
		"feat[ure",
		"feat\\ure",
		"feat\nure",
	}
	for _, name := range invalid {
		err := ValidateBranchName(name)
		if err == nil {
			t.Errorf("ValidateBranchName(%q) = nil, want error", name)
			continue
		}
		if !errors.Is(err, ErrInvalidBranch) {
			t.Errorf("ValidateBranchName(%q) = %v, want ErrInvalidBranch", name, err)
		}
	}
}

func TestValidateCommitMessage(t *testing.T) {
	if err := ValidateCommitMessage(""); err == nil {
		t.Error("empty message accepted")
	}
	if err := ValidateCommitMessage("   \n\t "); err == nil {
		t.Error("whitespace-only message accepted")
	}
	if err := ValidateCommitMessage("fix: retry payments"); err != nil {
		t.Errorf("valid message rejected: %v", err)
	}
}
