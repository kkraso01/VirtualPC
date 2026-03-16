package safety

import (
	"fmt"
	"path/filepath"
	"strings"
)

type FilesystemGuard struct {
	WritableRoot string
	BlockedRoots []string
}

func NewFilesystemGuard(writableRoot string) FilesystemGuard {
	return FilesystemGuard{WritableRoot: filepath.Clean(writableRoot), BlockedRoots: []string{"/proc", "/sys", "/dev"}}
}

func (g FilesystemGuard) Validate(path string) error {
	clean := filepath.Clean(path)
	for _, blocked := range g.BlockedRoots {
		if clean == blocked || strings.HasPrefix(clean, blocked+"/") {
			return fmt.Errorf("filesystem guard: access blocked for %s", clean)
		}
	}
	if clean != g.WritableRoot && !strings.HasPrefix(clean, g.WritableRoot+"/") {
		return fmt.Errorf("filesystem guard: path %s is outside writable root %s", clean, g.WritableRoot)
	}
	return nil
}
