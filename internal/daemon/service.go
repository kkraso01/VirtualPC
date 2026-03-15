package daemon

import (
	"errors"
	"fmt"
	"log"
	"virtualpc/internal/config"
	"virtualpc/internal/models"
	"virtualpc/internal/runtime/firecracker"
	"virtualpc/internal/state"
)

type Service struct {
	cfg     config.Config
	store   *state.Store
	runtime firecracker.Runtime
}

func New(cfg config.Config) (*Service, error) {
	st, err := state.New(cfg.DataPath)
	if err != nil {
		return nil, err
	}
	rt := firecracker.NewManager(cfg.FirecrackerDir)
	log.Printf("control-plane targets: postgres=%s nats=%s minio=%s temporal=%s", cfg.PostgresDSN, cfg.NATSURL, cfg.MinIOEndpoint, cfg.TemporalHost)
	return &Service{cfg: cfg, store: st, runtime: rt}, nil
}

func (s *Service) Status() map[string]any {
	st := s.store.Status()
	st["unix_socket"] = s.cfg.UnixSocket
	st["control_plane"] = map[string]string{"postgres": s.cfg.PostgresDSN, "nats": s.cfg.NATSURL, "minio": s.cfg.MinIOEndpoint, "temporal": s.cfg.TemporalHost}
	return st
}

func (s *Service) Profiles() []models.MachineProfile { return s.store.Profiles() }
func (s *Service) CreateMachine(profile string) (models.Machine, error) {
	return s.store.CreateMachine(profile)
}
func (s *Service) ListMachines() []models.Machine               { return s.store.ListMachines() }
func (s *Service) GetMachine(id string) (models.Machine, error) { return s.store.GetMachine(id) }

func validTransition(cur, next models.MachineState) bool {
	allowed := map[models.MachineState]map[models.MachineState]bool{
		models.MachinePending:  {models.MachineBooting: true, models.MachineDeleted: true},
		models.MachineBooting:  {models.MachineRunning: true, models.MachineFailed: true, models.MachineStopping: true},
		models.MachineRunning:  {models.MachineStopping: true, models.MachineFailed: true},
		models.MachineStopping: {models.MachineStopped: true, models.MachineFailed: true},
		models.MachineStopped:  {models.MachineBooting: true, models.MachineDeleted: true},
		models.MachineFailed:   {models.MachineDeleted: true, models.MachineBooting: true},
	}
	return allowed[cur][next]
}

func (s *Service) StartMachine(id string) (models.Machine, error) {
	m, err := s.store.GetMachine(id)
	if err != nil {
		return models.Machine{}, err
	}
	if !validTransition(m.State, models.MachineBooting) && !validTransition(m.State, models.MachineRunning) {
		return models.Machine{}, fmt.Errorf("invalid state transition: %s -> running", m.State)
	}
	runtimeID, err := s.runtime.Start(id)
	if err != nil {
		return models.Machine{}, err
	}
	m, err = s.store.UpdateMachine(id, func(in models.Machine) (models.Machine, error) {
		in.State = models.MachineRunning
		in.RuntimeID = runtimeID
		return in, nil
	})
	return m, err
}
func (s *Service) StopMachine(id string) (models.Machine, error) {
	m, err := s.store.GetMachine(id)
	if err != nil {
		return models.Machine{}, err
	}
	if m.State != models.MachineRunning {
		return models.Machine{}, errors.New("machine is not running")
	}
	if err := s.runtime.Stop(id); err != nil {
		return models.Machine{}, err
	}
	return s.store.UpdateMachine(id, func(in models.Machine) (models.Machine, error) { in.State = models.MachineStopped; return in, nil })
}
func (s *Service) DestroyMachine(id string) error {
	_, err := s.store.UpdateMachine(id, func(in models.Machine) (models.Machine, error) { in.State = models.MachineDeleted; return in, nil })
	if err != nil {
		return err
	}
	_ = s.runtime.Destroy(id)
	return s.store.DeleteMachine(id)
}
func (s *Service) Exec(id string, cmd []string) (string, error) { return s.runtime.Exec(id, cmd) }
func (s *Service) Logs(id string) (string, error)               { return s.runtime.Logs(id) }
func (s *Service) PS(id string) ([]string, error)               { return s.runtime.PS(id) }

func (s *Service) CreateProject(name string) (models.Project, error) {
	return s.store.CreateProject(name)
}
func (s *Service) ListProjects() []models.Project { return s.store.ListProjects() }
func (s *Service) AssignMachine(machineID, projectID string) (models.Machine, error) {
	return s.store.AssignMachine(machineID, projectID)
}
func (s *Service) CreateService(machineID, name, image string) (models.Service, error) {
	return s.store.CreateService(machineID, name, image)
}
func (s *Service) ListServices(machineID string) []models.Service {
	return s.store.ListServices(machineID)
}
func (s *Service) CreateSnapshot(machineID string) (models.Snapshot, error) {
	return s.store.CreateSnapshot(machineID)
}
func (s *Service) ListSnapshots() []models.Snapshot               { return s.store.ListSnapshots() }
func (s *Service) GetSnapshot(id string) (models.Snapshot, error) { return s.store.GetSnapshot(id) }
func (s *Service) ForkMachine(snapshotID string) (models.Machine, error) {
	return s.store.ForkMachine(snapshotID)
}
func (s *Service) CreateTask(machineID, goal string) (models.Task, error) {
	return s.store.CreateTask(machineID, goal)
}
func (s *Service) RunTask(taskID string) (models.Task, error) { return s.store.RunTask(taskID) }
func (s *Service) GetTask(taskID string) (models.Task, error) { return s.store.GetTask(taskID) }
