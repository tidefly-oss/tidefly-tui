package env

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func Root() string {
	cwd, _ := os.Getwd()
	if filepath.Base(cwd) == "tui" {
		return filepath.Dir(cwd)
	}
	if _, err := os.Stat(filepath.Join(cwd, "tui")); err == nil {
		return cwd
	}
	return filepath.Dir(cwd)
}

func PlanePath() string {
	return filepath.Join(Root(), "tidefly-plane")
}

func Load(extraPaths ...string) error {
	envType := os.Getenv("ENV_CHOICE")
	if envType == "" {
		envType = "development"
	}
	candidates := buildCandidates(envType, extraPaths...)
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

func buildCandidates(envType string, extra ...string) []string {
	plane := PlanePath()
	var paths []string
	paths = append(paths, extra...)
	paths = append(paths, filepath.Join(plane, "deploy", envType, ".env"))
	paths = append(paths, filepath.Join(plane, ".env"))
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
