package assets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRootUsesEnvOverrideWhenValid(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "models"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Setenv(EnvOverride, dir)

	got, err := FindRoot()
	if err != nil {
		t.Fatalf("FindRoot: %v", err)
	}
	if got != dir {
		t.Fatalf("expected FindRoot to honor %s override, got %q want %q", EnvOverride, got, dir)
	}
}

func TestFindRootIgnoresInvalidEnvOverride(t *testing.T) {
	t.Setenv(EnvOverride, t.TempDir()) // no "models" subdir inside

	// Falling back to cwd-relative "assets" should also fail from a test's
	// working directory (internal/assets has no ./assets folder of its
	// own), so we just assert the invalid override wasn't silently
	// accepted as a valid root.
	got, err := FindRoot()
	if err == nil && got == os.Getenv(EnvOverride) {
		t.Fatalf("expected invalid override to be rejected, got %q", got)
	}
}

func TestFindRootRealRepoAssetsDir(t *testing.T) {
	// From the repo root, "assets" (with a "models" subdir) should resolve.
	t.Setenv(EnvOverride, "")
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	got, err := FindRoot()
	if err != nil {
		t.Fatalf("FindRoot: %v", err)
	}
	if !looksLikeAssetsRoot(got) {
		t.Fatalf("resolved root %q doesn't look like an assets dir", got)
	}
}
