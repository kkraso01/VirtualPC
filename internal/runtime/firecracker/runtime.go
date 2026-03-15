package firecracker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Runtime interface {
	Start(machineID string) (string, error)
	Stop(machineID string) error
	Destroy(machineID string) error
	Exec(machineID string, command []string) (string, error)
	Logs(machineID string) (string, error)
	PS(machineID string) ([]string, error)
}

type Manager struct {
	baseDir string
}

func NewManager(baseDir string) *Manager { return &Manager{baseDir: baseDir} }

func (m *Manager) Start(machineID string) (string, error) {
	dir := filepath.Join(m.baseDir, machineID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	cfg := filepath.Join(dir, "firecracker-vm.json")
	content := []byte(fmt.Sprintf(`{"machine_id":"%s","kernel":"vmlinux","rootfs":"rootfs.ext4"}`+"\n", machineID))
	if err := os.WriteFile(cfg, content, 0o644); err != nil {
		return "", err
	}
	return "fc-" + machineID[:12], nil
}
func (m *Manager) Stop(machineID string) error {
	if _, err := os.Stat(filepath.Join(m.baseDir, machineID)); err != nil {
		return err
	}
	return nil
}
func (m *Manager) Destroy(machineID string) error {
	return os.RemoveAll(filepath.Join(m.baseDir, machineID))
}
func (m *Manager) Exec(machineID string, command []string) (string, error) {
	if len(command) == 0 {
		return "", errors.New("empty command")
	}
	return "guest-exec(" + machineID + "): " + fmt.Sprint(command), nil
}
func (m *Manager) Logs(machineID string) (string, error) { return "guest logs for " + machineID, nil }
func (m *Manager) PS(machineID string) ([]string, error) {
	return []string{"1 init", "22 vpc-agent", "58 containerd"}, nil
}
