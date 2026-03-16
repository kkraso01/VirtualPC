package safety

import (
	"fmt"
	"path/filepath"
	"strings"
)

type FilesystemGuard struct {
	WritableRoots []string
	BlockedRoots  []string
}

func NewFilesystemGuard(writableRoots ...string) FilesystemGuard {
	roots := []string{}
	for _, r := range writableRoots {
		if strings.TrimSpace(r) != "" {
			roots = append(roots, filepath.Clean(r))
		}
	}
	if len(roots) == 0 {
		roots = []string{"/workspace", "/tmp"}
	}
	return FilesystemGuard{WritableRoots: roots, BlockedRoots: []string{"/etc", "/usr", "/proc", "/sys", "/dev"}}
}

func (g FilesystemGuard) Validate(path string) error {
	clean := filepath.Clean(path)
	for _, blocked := range g.BlockedRoots {
		if clean == blocked || strings.HasPrefix(clean, blocked+"/") {
			return fmt.Errorf("filesystem guard: access blocked for %s", clean)
		}
	}
	for _, root := range g.WritableRoots {
		if clean == root || strings.HasPrefix(clean, root+"/") {
			return nil
		}
	}
	return fmt.Errorf("filesystem guard: path %s is outside writable roots %v", clean, g.WritableRoots)
}
