package env

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// PlaneDir returns the directory where tidefly-plane config lives.
// Matches TIDEFLY_DIR in install.sh (default: /etc/tidefly-plane)
func PlaneDir() string {
	if p := os.Getenv("TIDEFLY_DIR"); p != "" {
		return p
	}
	return "/etc/tidefly-plane"
}

// PlanePath is kept for backwards compatibility.
// Deprecated: use PlaneDir instead.
func PlanePath() string {
	return PlaneDir()
}

func Load(extraPaths ...string) error {
	candidates := buildCandidates(extraPaths...)
	for _, p := range candidates {
		if err := loadFile(p); err == nil {
			return nil
		}
	}
	return nil
}

func GetOrLoad(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	_ = Load()
	return os.Getenv(key)
}

func buildCandidates(extra ...string) []string {
	dir := PlaneDir()
	var paths []string
	paths = append(paths, extra...)
	paths = append(paths, filepath.Join(dir, ".env"))
	return paths
}

func loadFile(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	f, err := os.Open(abs)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
	return scanner.Err()
}
