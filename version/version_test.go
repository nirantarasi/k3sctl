package version

import "testing"

// IsPre reports whether the build is a pre-release. It reads the package-level
// Version, which the linker sets from the build tag, so each case sets it and
// restores the previous value afterwards; otherwise one case would leak into
// the next.
func TestIsPre(t *testing.T) {
	original := Version
	t.Cleanup(func() { Version = original })

	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{name: "release", version: "v1.0.0", want: false},
		{name: "release without the v prefix", version: "1.0.0", want: false},
		{name: "release candidate", version: "1.0.0-rc1", want: true},
		{name: "semver pre-release with a dot", version: "2.0.0-beta.1", want: true},
		{name: "alpha", version: "v0.1.0-alpha.3", want: true},
		{name: "development build", version: "", want: false},
		{name: "development build as versioninfo writes it", version: "(devel)", want: false},
		// Semver keeps build metadata after "+" separate from the pre-release after "-", so this one
		// is a false positive today. The test pins what the function does; changing it is a separate
		// decision, not a test concern.
		{name: "a dash in build metadata is read as pre-release (current behaviour)", version: "1.0.0+build-7", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			if got := IsPre(); got != tt.want {
				t.Errorf("IsPre() with Version %q = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}
