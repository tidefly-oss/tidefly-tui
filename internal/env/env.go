package env

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Load lädt die .env — sucht von cwd aufwärts nach deploy/dev/.env
func Load(extraPaths ...string) error {
	candidates := buildCandidates(extraPaths...)
	for _, p := range candidates {
		if err := loadFile(p); err == nil {
			return nil
		}
	}
	return nil
}

// GetOrLoad gibt einen Env-Wert zurück — lädt .env falls nötig
func GetOrLoad(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	_ = Load()
	return os.Getenv(key)
}

func buildCandidates(extra ...string) []string {
	cwd, _ := os.Getwd()
	var paths []string

	// Extra Pfade zuerst
	paths = append(paths, extra...)

	// Von cwd aufwärts nach deploy/dev/.env suchen
	dir := cwd
	for i := 0; i < 8; i++ {
		paths = append(paths, filepath.Join(dir, "deploy", "dev", ".env"))
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Von cwd aufwärts nach .env suchen
	dir = cwd
	for i := 0; i < 8; i++ {
		paths = append(paths, filepath.Join(dir, ".env"))
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

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
	defer f.Close()

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
			os.Setenv(key, value)
		}
	}
	return scanner.Err()
}
