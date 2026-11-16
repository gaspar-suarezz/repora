package domain

import "testing"

func TestNormalizeID(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"simple", "api", "api"},
		{"spaces", "Product X", "product-x"},
		{"trim and case", "  API  ", "api"},
		{"multiple separators", "Client A / Backend", "client-a-backend"},
		{"leading separators", "///Web", "web"},
		{"accents", "Café Service", "caf-service"},
		{"digits", "v2 Platform", "v2-platform"},
		{"empty", "   ", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeID(tc.in); got != tc.want {
				t.Fatalf("NormalizeID(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestIDValidate(t *testing.T) {
	valid := []string{"api", "product-x", "v2", "a"}
	for _, id := range valid {
		if err := RepositoryID(id).Validate(); err != nil {
			t.Errorf("RepositoryID(%q).Validate() = %v, want nil", id, err)
		}
	}

	invalid := []string{"", "API", "product x", "-api", "api-", "api--x", "a/b", ".", ".."}
	for _, id := range invalid {
		if err := WorkspaceID(id).Validate(); err == nil {
			t.Errorf("WorkspaceID(%q).Validate() = nil, want error", id)
		}
	}
}

func TestValidateName(t *testing.T) {
	if err := ValidateName(""); err == nil {
		t.Error("empty name accepted")
	}
	if err := ValidateName("   "); err == nil {
		t.Error("whitespace-only name accepted")
	}
	if err := ValidateName("Product X"); err != nil {
		t.Errorf("valid name rejected: %v", err)
	}
}
