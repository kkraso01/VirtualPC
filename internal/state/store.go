package state

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"virtualpc/internal/models"
)

type Store struct {
	mu   sync.RWMutex
	path string
	data models.DaemonState
}

func New(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	b, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.data = defaultState()
			return s.persistLocked()
		}
		return err
	}
	if len(b) == 0 {
		s.data = defaultState()
		return s.persistLocked()
	}
	if err := json.Unmarshal(b, &s.data); err != nil {
		return err
	}
	if s.data.Machines == nil {
		s.data = defaultState()
	}
	return nil
}

func defaultState() models.DaemonState {
	now := time.Now().UTC()
	profiles := map[string]models.MachineProfile{
		"minimal-shell":           {Name: "minimal-shell", BaseImage: "guest-base", VCPU: 2, MemoryMB: 2048, DiskGB: 20, ContainerdEnabled: true, NetworkMode: models.NetworkNAT, PolicyClass: "default"},
		"minimal-shell-offline":   {Name: "minimal-shell-offline", BaseImage: "guest-base", VCPU: 2, MemoryMB: 2048, DiskGB: 20, ContainerdEnabled: true, NetworkMode: models.NetworkOffline, PolicyClass: "offline"},
		"minimal-shell-allowlist": {Name: "minimal-shell-allowlist", BaseImage: "guest-base", VCPU: 2, MemoryMB: 2048, DiskGB: 20, ContainerdEnabled: true, NetworkMode: models.NetworkAllowlist, PolicyClass: "restricted", Allowlist: []string{"1.1.1.1", "8.8.8.8"}},
		"dev-python":              {Name: "dev-python", BaseImage: "guest-dev", VCPU: 2, MemoryMB: 4096, DiskGB: 30, ContainerdEnabled: true, NetworkMode: models.NetworkNAT, PolicyClass: "dev"},
		"dev-node":                {Name: "dev-node", BaseImage: "guest-dev", VCPU: 2, MemoryMB: 4096, DiskGB: 30, ContainerdEnabled: true, NetworkMode: models.NetworkNAT, PolicyClass: "dev"},
		"browser-dev":             {Name: "browser-dev", BaseImage: "guest-browser", VCPU: 4, MemoryMB: 8192, DiskGB: 40, BrowserEnabled: true, ContainerdEnabled: true, NetworkMode: models.NetworkNAT, PolicyClass: "browser"},
		"fullstack-dev":           {Name: "fullstack-dev", BaseImage: "guest-dev", VCPU: 4, MemoryMB: 8192, DiskGB: 60, BrowserEnabled: true, ContainerdEnabled: true, NetworkMode: models.NetworkNAT, PolicyClass: "fullstack"},
	}
	return models.DaemonState{
		Machines:    map[string]models.Machine{},
		Profiles:    profiles,
		Snapshots:   map[string]models.Snapshot{},
		Projects:    map[string]models.Project{},
		Services:    map[string]models.Service{},
		Tasks:       map[string]models.Task{},
		AuditEvents: []models.AuditEvent{},
		StartedAt:   now,
	}
}

func (s *Store) persistLocked() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o644)
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Store) Status() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]any{
		"started_at": s.data.StartedAt,
		"machines":   len(s.data.Machines),
		"projects":   len(s.data.Projects),
		"tasks":      len(s.data.Tasks),
		"profiles":   len(s.data.Profiles),
	}
}

func (s *Store) Profiles() []models.MachineProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.MachineProfile, 0, len(s.data.Profiles))
	for _, p := range s.data.Profiles {
		out = append(out, p)
	}
	return out
}

func (s *Store) GetProfile(name string) (models.MachineProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.data.Profiles[name]
	if !ok {
		return models.MachineProfile{}, errors.New("profile not found")
	}
	return p, nil
}

func (s *Store) CreateMachine(profile string) (models.Machine, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.Profiles[profile]; !ok {
		return models.Machine{}, errors.New("profile not found")
	}
	now := time.Now().UTC()
	m := models.Machine{ID: newID(), ProfileName: profile, State: models.MachinePending, CreatedAt: now, UpdatedAt: now}
	s.data.Machines[m.ID] = m
	s.auditLocked("machine.created", m.ID, "machine created")
	return m, s.persistLocked()
}

func (s *Store) ListMachines() []models.Machine {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Machine, 0, len(s.data.Machines))
	for _, m := range s.data.Machines {
		out = append(out, m)
	}
	return out
}
func (s *Store) GetMachine(id string) (models.Machine, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.data.Machines[id]
	if !ok {
		return models.Machine{}, errors.New("machine not found")
	}
	return m, nil
}

func (s *Store) UpdateMachine(id string, update func(models.Machine) (models.Machine, error)) (models.Machine, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.data.Machines[id]
	if !ok {
		return models.Machine{}, errors.New("machine not found")
	}
	next, err := update(m)
	if err != nil {
		return models.Machine{}, err
	}
	next.UpdatedAt = time.Now().UTC()
	s.data.Machines[id] = next
	s.auditLocked("machine.updated", id, string(next.State))
	return next, s.persistLocked()
}

func (s *Store) DeleteMachine(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.Machines[id]; !ok {
		return errors.New("machine not found")
	}
	delete(s.data.Machines, id)
	s.auditLocked("machine.deleted", id, "machine removed")
	return s.persistLocked()
}

func (s *Store) CreateProject(name string) (models.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := models.Project{ID: newID(), Name: name, CreatedAt: time.Now().UTC()}
	s.data.Projects[p.ID] = p
	s.auditLocked("project.created", p.ID, name)
	return p, s.persistLocked()
}
func (s *Store) ListProjects() []models.Project {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Project, 0, len(s.data.Projects))
	for _, p := range s.data.Projects {
		out = append(out, p)
	}
	return out
}

func (s *Store) AssignMachine(machineID, projectID string) (models.Machine, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, mok := s.data.Machines[machineID]
	if !mok {
		return models.Machine{}, errors.New("machine not found")
	}
	if _, pok := s.data.Projects[projectID]; !pok {
		return models.Machine{}, errors.New("project not found")
	}
	m.ProjectID = projectID
	m.UpdatedAt = time.Now().UTC()
	s.data.Machines[machineID] = m
	s.auditLocked("machine.assigned", machineID, projectID)
	return m, s.persistLocked()
}

func (s *Store) CreateService(machineID, name, image string) (models.Service, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.Machines[machineID]; !ok {
		return models.Service{}, errors.New("machine not found")
	}
	svc := models.Service{ID: newID(), MachineID: machineID, Name: name, Image: image, Status: "running", CreatedAt: time.Now().UTC()}
	s.data.Services[svc.ID] = svc
	s.auditLocked("service.created", svc.ID, machineID+":"+name)
	return svc, s.persistLocked()
}
func (s *Store) ListServices(machineID string) []models.Service {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []models.Service{}
	for _, svc := range s.data.Services {
		if machineID == "" || svc.MachineID == machineID {
			out = append(out, svc)
		}
	}
	return out
}

func (s *Store) DeleteService(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.Services[id]; !ok {
		return errors.New("service not found")
	}
	delete(s.data.Services, id)
	s.auditLocked("service.deleted", id, "removed")
	return s.persistLocked()
}

func (s *Store) CreateSnapshot(machineID string) (models.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.Machines[machineID]; !ok {
		return models.Snapshot{}, errors.New("machine not found")
	}
	snap := models.Snapshot{ID: newID(), MachineID: machineID, DiskRef: "state://workspace/" + machineID, Metadata: "disk+workspace metadata", CreatedAt: time.Now().UTC()}
	s.data.Snapshots[snap.ID] = snap
	s.auditLocked("snapshot.created", snap.ID, machineID)
	return snap, s.persistLocked()
}
func (s *Store) ListSnapshots() []models.Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Snapshot, 0, len(s.data.Snapshots))
	for _, snp := range s.data.Snapshots {
		out = append(out, snp)
	}
	return out
}
func (s *Store) GetSnapshot(id string) (models.Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snp, ok := s.data.Snapshots[id]
	if !ok {
		return models.Snapshot{}, errors.New("snapshot not found")
	}
	return snp, nil
}

func (s *Store) ForkMachine(snapshotID string) (models.Machine, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snp, ok := s.data.Snapshots[snapshotID]
	if !ok {
		return models.Machine{}, errors.New("snapshot not found")
	}
	parent, ok := s.data.Machines[snp.MachineID]
	if !ok {
		return models.Machine{}, errors.New("parent machine missing")
	}
	now := time.Now().UTC()
	child := models.Machine{ID: newID(), ProfileName: parent.ProfileName, ProjectID: parent.ProjectID, State: models.MachinePending, CreatedAt: now, UpdatedAt: now}
	s.data.Machines[child.ID] = child
	s.auditLocked("machine.forked", child.ID, snapshotID)
	return child, s.persistLocked()
}

func (s *Store) CreateTask(machineID, goal string) (models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.Machines[machineID]; !ok {
		return models.Task{}, errors.New("machine not found")
	}
	now := time.Now().UTC()
	t := models.Task{ID: newID(), MachineID: machineID, Goal: goal, Status: "created", Logs: []string{"task created"}, CreatedAt: now, UpdatedAt: now}
	s.data.Tasks[t.ID] = t
	s.auditLocked("task.created", t.ID, machineID)
	return t, s.persistLocked()
}
func (s *Store) GetTask(id string) (models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.data.Tasks[id]
	if !ok {
		return models.Task{}, errors.New("task not found")
	}
	return t, nil
}
func (s *Store) RunTask(id string) (models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.data.Tasks[id]
	if !ok {
		return models.Task{}, errors.New("task not found")
	}
	t.Status = "running"
	t.UpdatedAt = time.Now().UTC()
	t.Logs = append(t.Logs, "task queued", "task running")
	s.data.Tasks[id] = t
	_ = s.persistLocked()
	return t, nil
}

func (s *Store) CompleteTask(id, status, output string) (models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.data.Tasks[id]
	if !ok {
		return models.Task{}, errors.New("task not found")
	}
	t.Status = status
	t.Logs = append(t.Logs, output)
	t.UpdatedAt = time.Now().UTC()
	s.data.Tasks[id] = t
	s.auditLocked("task.completed", id, status)
	return t, s.persistLocked()
}

func (s *Store) UpdateSnapshot(snap models.Snapshot) (models.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.Snapshots[snap.ID]; !ok {
		return models.Snapshot{}, errors.New("snapshot not found")
	}
	s.data.Snapshots[snap.ID] = snap
	return snap, s.persistLocked()
}

func (s *Store) auditLocked(kind, subj, msg string) {
	s.data.AuditEvents = append(s.data.AuditEvents, models.AuditEvent{ID: newID(), Type: kind, SubjectID: subj, Message: msg, CreatedAt: time.Now().UTC()})
}
