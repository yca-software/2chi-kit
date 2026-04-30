package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// apiGoModule must match the module line in this app's go.mod (template / fork).
const apiGoModule = "github.com/yca-software/2chi-kit/go-api"

// ModuleRoot returns the directory containing this API module's go.mod.
// Resolution walks upward from this package's source file; it does not use the process cwd.
func ModuleRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("runtime.Caller failed")
	}
	startDir := filepath.Dir(file)
	dir := startDir

	for {
		modPath := filepath.Join(dir, "go.mod")
		if root, ok := matchModuleFile(modPath, apiGoModule); ok {
			return root, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod for module %q not found (started from %s)", apiGoModule, startDir)
		}
		dir = parent
	}
}

func matchModuleFile(modPath, wantModule string) (root string, ok bool) {
	data, err := os.ReadFile(modPath)
	if err != nil {
		return "", false
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		rest, found := strings.CutPrefix(line, "module ")
		if !found {
			continue
		}
		mod := strings.Fields(rest)[0]
		if mod != wantModule {
			return "", false
		}
		return filepath.Dir(modPath), true
	}
	return "", false
}
