package firecracker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const defaultGuestAgentPort uint32 = 10789

type VMStatus string

const (
	VMStatusStarting VMStatus = "starting"
	VMStatusRunning  VMStatus = "running"
	VMStatusStopped  VMStatus = "stopped"
	VMStatusFailed   VMStatus = "failed"
)

type VMProcessState struct {
	MachineID      string    `json:"machine_id"`
	RuntimeID      string    `json:"runtime_id"`
	Firecracker    int       `json:"firecracker_pid"`
	AgentPID       int       `json:"agent_pid"`
	AgentSocket    string    `json:"agent_socket"`
	AgentVSockCID  uint32    `json:"agent_vsock_cid"`
	AgentVSockPort uint32    `json:"agent_vsock_port"`
	Status         VMStatus  `json:"status"`
	Failure        string    `json:"failure,omitempty"`
	StartedAt      time.Time `json:"started_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	RuntimeDir     string    `json:"runtime_dir"`
	LogFile        string    `json:"log_file"`
	StateFile      string    `json:"state_file"`
	GuestFS        string    `json:"guest_fs"`
	SnapshotDir    string    `json:"snapshot_dir"`
	NetworkState   string    `json:"network_state"`
}

type ProcessManager struct {
	baseDir         string
	firecrackerBin  string
	agentBin        string
	mu              sync.RWMutex
	states          map[string]*VMProcessState
	healthCancelers map[string]context.CancelFunc
}

func NewProcessManager(baseDir, firecrackerBin, agentBin string) *ProcessManager {
	pm := &ProcessManager{baseDir: baseDir, firecrackerBin: firecrackerBin, agentBin: agentBin, states: map[string]*VMProcessState{}, healthCancelers: map[string]context.CancelFunc{}}
	pm.loadPersisted()
	return pm
}

func (m *ProcessManager) loadPersisted() {
	_ = filepath.Walk(filepath.Join(m.baseDir), func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() || info.Name() != "vm_state.json" {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		var st VMProcessState
		if json.Unmarshal(b, &st) == nil {
			if st.Status == VMStatusRunning && !isAlive(st.Firecracker) {
				st.Status = VMStatusFailed
				st.Failure = "process missing after daemon restart"
			}
			cp := st
			m.states[st.MachineID] = &cp
		}
		return nil
	})
}

func (m *ProcessManager) StartVM(machineID string) (*VMProcessState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if st, ok := m.states[machineID]; ok && st.Status == VMStatusRunning && isAlive(st.Firecracker) {
		return st, nil
	}
	runtimeDir := filepath.Join(m.baseDir, machineID)
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		return nil, err
	}
	guestFS := filepath.Join(runtimeDir, "guestfs")
	_ = os.MkdirAll(guestFS, 0o755)
	agentSock := filepath.Join(runtimeDir, "agent.sock")
	_ = os.Remove(agentSock)
	logFile := filepath.Join(runtimeDir, "firecracker.log")
	lf, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	defer lf.Close()
	if m.firecrackerBin == "" {
		return nil, errors.New("firecracker binary not configured")
	}
	if _, err := os.Stat(m.firecrackerBin); err != nil {
		return nil, fmt.Errorf("firecracker binary unavailable: %w", err)
	}
	fcCmd := exec.Command(m.firecrackerBin, "--api-sock", filepath.Join(runtimeDir, "firecracker.sock"))
	fcCmd.Stdout = lf
	fcCmd.Stderr = lf
	if err := fcCmd.Start(); err != nil {
		return nil, err
	}
	if m.agentBin == "" {
		_ = fcCmd.Process.Kill()
		return nil, errors.New("guest agent binary not configured")
	}
	if _, err := os.Stat(m.agentBin); err != nil {
		_ = fcCmd.Process.Kill()
		return nil, fmt.Errorf("guest agent binary unavailable: %w", err)
	}
	agentCmd := exec.Command(m.agentBin, "server", "--socket", agentSock, "--machine-id", machineID, "--root", guestFS)
	agentCmd.Stdout = lf
	agentCmd.Stderr = lf
	if err := agentCmd.Start(); err != nil {
		_ = fcCmd.Process.Kill()
		return nil, err
	}
	now := time.Now().UTC()
	runtimeID := "fc-" + machineID
	if len(machineID) > 12 {
		runtimeID = "fc-" + machineID[:12]
	}
	st := &VMProcessState{MachineID: machineID, RuntimeID: runtimeID, Firecracker: fcCmd.Process.Pid, AgentPID: agentCmd.Process.Pid, AgentSocket: agentSock, AgentVSockCID: 3, AgentVSockPort: defaultGuestAgentPort, Status: VMStatusRunning, StartedAt: now, UpdatedAt: now, RuntimeDir: runtimeDir, LogFile: logFile, StateFile: filepath.Join(runtimeDir, "vm_state.json"), GuestFS: guestFS, SnapshotDir: filepath.Join(runtimeDir, "snapshots"), NetworkState: filepath.Join(runtimeDir, "network.json")}
	m.states[machineID] = st
	if err := m.persist(st); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.healthCancelers[machineID] = cancel
	go m.healthLoop(ctx, machineID)
	return st, nil
}

func (m *ProcessManager) StopVM(machineID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	st, ok := m.states[machineID]
	if !ok {
		return errors.New("vm not found")
	}
	if c, ok := m.healthCancelers[machineID]; ok {
		c()
		delete(m.healthCancelers, machineID)
	}
	_ = signalTERM(st.AgentPID)
	_ = signalTERM(st.Firecracker)
	st.Status = VMStatusStopped
	st.UpdatedAt = time.Now().UTC()
	return m.persist(st)
}

func (m *ProcessManager) KillVM(machineID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	st, ok := m.states[machineID]
	if !ok {
		return nil
	}
	if c, ok := m.healthCancelers[machineID]; ok {
		c()
		delete(m.healthCancelers, machineID)
	}
	_ = syscall.Kill(st.AgentPID, syscall.SIGKILL)
	_ = syscall.Kill(st.Firecracker, syscall.SIGKILL)
	st.Status = VMStatusStopped
	st.UpdatedAt = time.Now().UTC()
	_ = os.Remove(st.AgentSocket)
	_ = os.Remove(filepath.Join(st.RuntimeDir, "firecracker.sock"))
	delete(m.states, machineID)
	_ = os.Remove(st.StateFile)
	return nil
}

func (m *ProcessManager) InspectVM(machineID string) (*VMProcessState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st, ok := m.states[machineID]
	if !ok {
		return nil, errors.New("vm not found")
	}
	cp := *st
	return &cp, nil
}

func (m *ProcessManager) VMStatus(machineID string) (VMStatus, error) {
	st, err := m.InspectVM(machineID)
	if err != nil {
		return "", err
	}
	return st.Status, nil
}

func (m *ProcessManager) healthLoop(ctx context.Context, machineID string) {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			m.mu.Lock()
			st, ok := m.states[machineID]
			if !ok {
				m.mu.Unlock()
				return
			}
			if st.Status == VMStatusRunning && (!isAlive(st.Firecracker) || !isAlive(st.AgentPID)) {
				st.Status = VMStatusFailed
				st.Failure = "vm process crashed"
				st.UpdatedAt = time.Now().UTC()
				_ = m.persist(st)
			}
			m.mu.Unlock()
		}
	}
}

func (m *ProcessManager) persist(st *VMProcessState) error {
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(st.StateFile, b, 0o644)
}

func isAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}

func signalTERM(pid int) error {
	if pid <= 0 {
		return nil
	}
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return err
	}
	for i := 0; i < 10; i++ {
		if !isAlive(pid) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("pid %s did not stop", strconv.Itoa(pid))
}
