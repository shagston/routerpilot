package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/shagston/routerpilot/internal/registry"
	"github.com/shagston/routerpilot/sdk/plugin"
)

type Loader struct {
	dir    string
	mu     sync.RWMutex
	loaded []plugin.Plugin
}

func NewLoader(dir string) *Loader {
	return &Loader{dir: dir}
}

func (l *Loader) LoadAll(ctx context.Context, reg *registry.ToolRegistry) error {
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading plugin dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !isPluginBinary(entry) {
			continue
		}

		path := filepath.Join(l.dir, entry.Name())
		plug, loadErr := loadSubprocess(ctx, path)
		if loadErr != nil {
			continue
		}

		t := plug.Tool()
		if regErr := reg.Register(t); regErr != nil {
			continue
		}

		l.mu.Lock()
		l.loaded = append(l.loaded, plug)
		l.mu.Unlock()
	}
	return nil
}

func (l *Loader) List() []plugin.Plugin {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]plugin.Plugin, len(l.loaded))
	copy(result, l.loaded)
	return result
}

func isPluginBinary(entry os.DirEntry) bool {
	if entry.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return strings.HasSuffix(strings.ToLower(entry.Name()), ".exe")
	}
	info, err := entry.Info()
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0
}
