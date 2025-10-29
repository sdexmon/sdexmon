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
		{"0.1.0", "0.1.1", true},   // without 'v' prefix
		{"v0.1.0", "0.1.1", true},  // mixed prefix
	}

	for _, tt := range tests {
		result := CompareVersions(tt.local, tt.remote)
		if result != tt.expected {
			t.Errorf("CompareVersions(%q, %q) = %v, expected %v",
				tt.local, tt.remote, result, tt.expected)
		}
	}
}
