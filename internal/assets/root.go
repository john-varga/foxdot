package assets

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnvOverride, if set, takes precedence over every other resolution
// strategy — handy for running from an unusual layout or pointing at a
// different art pack entirely.
const EnvOverride = "FOXDOT_ASSETS_DIR"

// FindRoot locates the assets directory across the ways this binary might
// be run: a normal `go run`/`go build` from the repo (cwd-relative), or a
// packaged binary shipped with an assets/ folder next to it. Returns an
// error listing everywhere it looked if none pan out, so a missing/misnamed
// folder is easy to diagnose on any of macOS/Windows/Linux.
func FindRoot() (string, error) {
	var tried []string

	if dir := os.Getenv(EnvOverride); dir != "" {
		tried = append(tried, dir)
		if looksLikeAssetsRoot(dir) {
			return dir, nil
		}
	}

	if cwd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(cwd, "assets")
		tried = append(tried, candidate)
		if looksLikeAssetsRoot(candidate) {
			return candidate, nil
		}
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		for _, rel := range []string{"assets", filepath.Join("..", "assets")} {
			candidate := filepath.Join(exeDir, rel)
			tried = append(tried, candidate)
			if looksLikeAssetsRoot(candidate) {
				return candidate, nil
			}
		}
	}

	return "", fmt.Errorf("assets: could not find an assets/ directory (tried %v; set %s to override)", tried, EnvOverride)
}

// looksLikeAssetsRoot is a light sanity check rather than validating every
// catalog file, so it stays cheap to call speculatively.
func looksLikeAssetsRoot(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, "models"))
	return err == nil && info.IsDir()
}
