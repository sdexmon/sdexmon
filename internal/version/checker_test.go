package version

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		local    string
		remote   string
		expected bool
	}{
		{"v0.1.0", "v0.1.1", true},
		{"v0.1.1", "v0.1.0", false},
		{"v0.1.0", "v0.1.0", false},
		{"v0.1.9", "v0.2.0", true},
		{"v1.0.0", "v0.9.9", false},
		{"v0.9.0", "v0.10.0", true},
		{"v1.9.9", "v2.0.0", true},
		{"v1.0.0-beta.1", "v1.0.0", true},
		{"v1.0.0", "v1.0.0-beta.1", false},
		{"v1.0.0-beta.2", "v1.0.0-beta.10", true},
		{"not-a-version", "v1.0.0", false},
		{"0.1.0", "0.1.1", true},  // without 'v' prefix
		{"v0.1.0", "0.1.1", true}, // mixed prefix
	}

	for _, tt := range tests {
		result := CompareVersions(tt.local, tt.remote)
		if result != tt.expected {
			t.Errorf("CompareVersions(%q, %q) = %v, expected %v",
				tt.local, tt.remote, result, tt.expected)
		}
	}
}
