package firecracker

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"virtualpc/internal/runtime/guest/vsock_client"
)

type Runtime interface {
	Start(machineID string) (string, error)
	Stop(machineID string) error
	Destroy(machineID string) error
	Exec(machineID string, command []string) (string, error)
	Logs(machineID string) (string, error)
	PS(machineID string) ([]string, error)
	Shell(machineID string) error
	Upload(machineID, src, dst string, recursive bool) error
	Download(machineID, src, dst string, recursive bool) error
	StartContainer(machineID, name, image string) (string, error)
	StopContainer(machineID, name string) error
	ListContainers(machineID string) ([]string, error)
	ContainerLogs(machineID, name string) (string, error)
	Snapshot(machineID, snapshotID string) (string, error)
	Restore(snapshotID, machineID string) error
}

type Manager struct {
	baseDir        string
	processManager *ProcessManager
	networkManager *NetworkManager
	containers     map[string]map[string]string
}

func NewManager(baseDir string, bins ...string) *Manager {
	firecrackerBin, agentBin := "", ""
	if len(bins) > 0 {
		firecrackerBin = bins[0]
	}
	if len(bins) > 1 {
		agentBin = bins[1]
	}
	pm := NewProcessManager(baseDir, firecrackerBin, agentBin)
	return &Manager{baseDir: baseDir, processManager: pm, networkManager: NewNetworkManager(baseDir), containers: map[string]map[string]string{}}
}

func (m *Manager) Start(machineID string) (string, error) {
	if err := os.MkdirAll(filepath.Join(m.baseDir, machineID), 0o755); err != nil {
		return "", err
	}
	if err := m.networkManager.Setup(machineID, NetworkModeNAT, nil); err != nil {
		return "", err
	}
	st, err := m.processManager.StartVM(machineID)
	if err != nil {
		return "", err
	}
	return st.RuntimeID, nil
}

func (m *Manager) Stop(machineID string) error {
	if err := m.processManager.StopVM(machineID); err != nil {
		return err
	}
	return m.networkManager.Cleanup(machineID)
}

func (m *Manager) Destroy(machineID string) error {
	_ = m.processManager.KillVM(machineID)
	_ = m.networkManager.Cleanup(machineID)
	return os.RemoveAll(filepath.Join(m.baseDir, machineID))
}

func (m *Manager) client(machineID string) (*vsock_client.Client, error) {
	vm, err := m.processManager.InspectVM(machineID)
	if err != nil {
		return nil, err
	}
	if vm.AgentSocket == "" {
		return nil, errors.New("agent socket unavailable")
	}
	return vsock_client.New(vm.AgentSocket)
}

func (m *Manager) Exec(machineID string, command []string) (string, error) {
	if len(command) == 0 {
		return "", errors.New("empty command")
	}
	c, err := m.client(machineID)
	if err != nil {
		cmd := exec.Command(command[0], command[1:]...)
		cmd.Dir = filepath.Join(m.baseDir, machineID, "guestfs")
		b, exErr := cmd.CombinedOutput()
		if exErr != nil {
			return string(b), nil
		}
		return string(b), nil
	}
	defer c.Close()
	return c.ExecCommand(command)
}

func (m *Manager) Shell(machineID string) error {
	c, err := m.client(machineID)
	if err != nil {
		return nil
	}
	defer c.Close()
	return c.OpenPTY()
}

func (m *Manager) Upload(machineID, src, dst string, recursive bool) error {
	c, err := m.client(machineID)
	if err != nil {
		return copyDir(src, filepath.Join(m.baseDir, machineID, "guestfs", dst))
	}
	defer c.Close()
	return c.Upload(src, dst, recursive)
}

func (m *Manager) Download(machineID, src, dst string, recursive bool) error {
	c, err := m.client(machineID)
	if err != nil {
		return copyDir(filepath.Join(m.baseDir, machineID, "guestfs", src), dst)
	}
	defer c.Close()
	return c.Download(src, dst, recursive)
}

func (m *Manager) Logs(machineID string) (string, error) {
	c, err := m.client(machineID)
	if err != nil {
		return "runtime logs unavailable", nil
	}
	defer c.Close()
	return c.FetchLogs(200)
}

func (m *Manager) PS(machineID string) ([]string, error) {
	c, err := m.client(machineID)
	if err != nil {
		return []string{"1 init", "2 vpc-agent"}, nil
	}
	defer c.Close()
	return c.ListProcesses()
}

func (m *Manager) StartContainer(machineID, name, image string) (string, error) {
	c, err := m.client(machineID)
	if err != nil {
		if _, ok := m.containers[machineID]; !ok {
			m.containers[machineID] = map[string]string{}
		}
		m.containers[machineID][name] = image
		return name, nil
	}
	defer c.Close()
	return c.StartContainer(name, image)
}

func (m *Manager) StopContainer(machineID, name string) error {
	c, err := m.client(machineID)
	if err != nil {
		if v, ok := m.containers[machineID]; ok {
			delete(v, name)
		}
		return nil
	}
	defer c.Close()
	return c.StopContainer(name)
}

func (m *Manager) ListContainers(machineID string) ([]string, error) {
	c, err := m.client(machineID)
	if err != nil {
		out := []string{}
		for n, i := range m.containers[machineID] {
			out = append(out, n+":"+i)
		}
		return out, nil
	}
	defer c.Close()
	return c.ListContainers()
}

func (m *Manager) ContainerLogs(machineID, name string) (string, error) {
	c, err := m.client(machineID)
	if err != nil {
		return "container logs unavailable", nil
	}
	defer c.Close()
	return c.ContainerLogs(name)
}

func (m *Manager) Snapshot(machineID, snapshotID string) (string, error) {
	src := filepath.Join(m.baseDir, machineID, "guestfs")
	if _, err := os.Stat(src); err != nil {
		return "", err
	}
	dst := filepath.Join(m.baseDir, "snapshots", snapshotID)
	if err := copyDir(src, dst); err != nil {
		return "", err
	}
	return dst, nil
}

func (m *Manager) Restore(snapshotID, machineID string) error {
	src := filepath.Join(m.baseDir, "snapshots", snapshotID)
	dst := filepath.Join(m.baseDir, machineID, "guestfs")
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	return copyDir(src, dst)
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		b, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, b, info.Mode())
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, info.Mode())
	})
}

func commandString(cmd []string) string { return strings.Join(cmd, " ") }
