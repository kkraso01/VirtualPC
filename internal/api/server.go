package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"virtualpc/internal/daemon"
)

type Server struct{ svc *daemon.Service }

func New(svc *daemon.Service) *Server { return &Server{svc: svc} }

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/daemon/status", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.svc.Status()) })
	mux.HandleFunc("GET /v1/config", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.svc.Status()["control_plane"]) })
	mux.HandleFunc("GET /v1/profiles", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.svc.Profiles()) })
	mux.HandleFunc("POST /v1/machines", s.createMachine)
	mux.HandleFunc("GET /v1/machines", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.svc.ListMachines()) })
	mux.HandleFunc("GET /v1/machines/", s.machineGet)
	mux.HandleFunc("POST /v1/machines/", s.machineActions)
	mux.HandleFunc("POST /v1/projects", s.createProject)
	mux.HandleFunc("GET /v1/projects", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.svc.ListProjects()) })
	mux.HandleFunc("POST /v1/assign", s.assignMachine)
	mux.HandleFunc("POST /v1/services", s.createService)
	mux.HandleFunc("GET /v1/services", s.listServices)
	mux.HandleFunc("POST /v1/snapshots", s.createSnapshot)
	mux.HandleFunc("GET /v1/snapshots", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.svc.ListSnapshots()) })
	mux.HandleFunc("GET /v1/snapshots/", s.getSnapshot)
	mux.HandleFunc("POST /v1/fork", s.forkMachine)
	mux.HandleFunc("POST /v1/tasks", s.createTask)
	mux.HandleFunc("POST /v1/tasks/run", s.runTask)
	mux.HandleFunc("GET /v1/tasks/", s.getTask)
	return mux
}

func (s *Server) createMachine(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Profile string `json:"profile"`
	}
	if !decode(w, r, &in) {
		return
	}
	m, err := s.svc.CreateMachine(in.Profile)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, m)
}
func (s *Server) machineGet(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/machines/")
	m, err := s.svc.GetMachine(id)
	if err != nil {
		writeErr(w, 404, err.Error())
		return
	}
	writeJSON(w, 200, m)
}
func (s *Server) machineActions(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Path, "/v1/machines/")
	parts := strings.Split(p, "/")
	if len(parts) < 2 {
		writeErr(w, 400, "invalid path")
		return
	}
	id, action := parts[0], parts[1]
	switch action {
	case "start":
		m, err := s.svc.StartMachine(id)
		if err != nil {
			writeErr(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, m)
	case "stop":
		m, err := s.svc.StopMachine(id)
		if err != nil {
			writeErr(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, m)
	case "destroy":
		if err := s.svc.DestroyMachine(id); err != nil {
			writeErr(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, map[string]string{"status": "deleted"})
	case "exec":
		var in struct {
			Command []string `json:"command"`
		}
		if !decode(w, r, &in) {
			return
		}
		out, err := s.svc.Exec(id, in.Command)
		if err != nil {
			writeErr(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, map[string]string{"output": out})
	case "logs":
		out, err := s.svc.Logs(id)
		if err != nil {
			writeErr(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, map[string]string{"logs": out})
	case "ps":
		out, err := s.svc.PS(id)
		if err != nil {
			writeErr(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, map[string]any{"processes": out})
	default:
		writeErr(w, 404, "unknown action")
	}
}
func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name string `json:"name"`
	}
	if !decode(w, r, &in) {
		return
	}
	p, err := s.svc.CreateProject(in.Name)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, p)
}
func (s *Server) assignMachine(w http.ResponseWriter, r *http.Request) {
	var in struct {
		MachineID string `json:"machine_id"`
		ProjectID string `json:"project_id"`
	}
	if !decode(w, r, &in) {
		return
	}
	m, err := s.svc.AssignMachine(in.MachineID, in.ProjectID)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, m)
}
func (s *Server) createService(w http.ResponseWriter, r *http.Request) {
	var in struct{ MachineID, Name, Image string }
	if !decode(w, r, &in) {
		return
	}
	svc, err := s.svc.CreateService(in.MachineID, in.Name, in.Image)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, svc)
}
func (s *Server) listServices(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.svc.ListServices(r.URL.Query().Get("machine_id")))
}
func (s *Server) createSnapshot(w http.ResponseWriter, r *http.Request) {
	var in struct {
		MachineID string `json:"machine_id"`
	}
	if !decode(w, r, &in) {
		return
	}
	snap, err := s.svc.CreateSnapshot(in.MachineID)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, snap)
}
func (s *Server) getSnapshot(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/snapshots/")
	snap, err := s.svc.GetSnapshot(id)
	if err != nil {
		writeErr(w, 404, err.Error())
		return
	}
	writeJSON(w, 200, snap)
}
func (s *Server) forkMachine(w http.ResponseWriter, r *http.Request) {
	var in struct {
		SnapshotID string `json:"snapshot_id"`
	}
	if !decode(w, r, &in) {
		return
	}
	m, err := s.svc.ForkMachine(in.SnapshotID)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, m)
}
func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var in struct{ MachineID, Goal string }
	if !decode(w, r, &in) {
		return
	}
	t, err := s.svc.CreateTask(in.MachineID, in.Goal)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, t)
}
func (s *Server) runTask(w http.ResponseWriter, r *http.Request) {
	var in struct {
		TaskID string `json:"task_id"`
	}
	if !decode(w, r, &in) {
		return
	}
	t, err := s.svc.RunTask(in.TaskID)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, t)
}
func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
	t, err := s.svc.GetTask(id)
	if err != nil {
		writeErr(w, 404, err.Error())
		return
	}
	writeJSON(w, 200, t)
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeErr(w, 400, "invalid json")
		return false
	}
	return true
}
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
