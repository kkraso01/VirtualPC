package server

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"virtualpc/internal/runtime/guest/protocol"
)

type Server struct {
	SocketPath string
	MachineID  string
	Root       string
	mu         sync.Mutex
	containers map[string]string
}

func New(socket, machineID, root string) *Server {
	return &Server{SocketPath: socket, MachineID: machineID, Root: root, containers: map[string]string{}}
}

func (s *Server) ListenAndServe() error {
	_ = os.Remove(s.SocketPath)
	if err := os.MkdirAll(filepath.Dir(s.SocketPath), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(s.Root, 0o755); err != nil {
		return err
	}
	ln, err := net.Listen("unix", s.SocketPath)
	if err != nil {
		return err
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			return err
		}
		go s.handleConn(c)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			return
		}
		var req protocol.Request
		if err := json.Unmarshal(line, &req); err != nil {
			return
		}
		resp := s.dispatch(req)
		b, _ := json.Marshal(resp)
		_, _ = w.Write(append(b, '\n'))
		_ = w.Flush()
	}
}

func (s *Server) dispatch(req protocol.Request) protocol.Response {
	ok := func(v any) protocol.Response {
		b, _ := json.Marshal(v)
		return protocol.Response{ID: req.ID, OK: true, Payload: b}
	}
	bad := func(err error) protocol.Response { return protocol.Response{ID: req.ID, OK: false, Error: err.Error()} }
	switch req.Method {
	case "ExecCommand":
		var in struct {
			Command []string `json:"command"`
		}
		_ = json.Unmarshal(req.Payload, &in)
		if len(in.Command) == 0 {
			return bad(fmt.Errorf("empty command"))
		}
		cmd := exec.Command(in.Command[0], in.Command[1:]...)
		cmd.Dir = s.Root
		out, err := cmd.CombinedOutput()
		if err != nil {
			return ok(map[string]string{"output": string(out) + "\n" + err.Error()})
		}
		return ok(map[string]string{"output": string(out)})
	case "OpenPTY":
		return ok(map[string]string{"status": "ready"})
	case "SendInput", "ClosePTY":
		return ok(map[string]string{"status": "ok"})
	case "ReadFile":
		var in struct {
			Path string `json:"path"`
		}
		_ = json.Unmarshal(req.Payload, &in)
		b, err := os.ReadFile(filepath.Join(s.Root, strings.TrimPrefix(in.Path, "/")))
		if err != nil {
			return bad(err)
		}
		return ok(map[string]string{"content_b64": base64.StdEncoding.EncodeToString(b)})
	case "WriteFile", "UploadFile":
		var in struct{ Path, ContentB64 string }
		_ = json.Unmarshal(req.Payload, &in)
		target := filepath.Join(s.Root, strings.TrimPrefix(in.Path, "/"))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return bad(err)
		}
		b, err := base64.StdEncoding.DecodeString(in.ContentB64)
		if err != nil {
			return bad(err)
		}
		if err := os.WriteFile(target, b, 0o644); err != nil {
			return bad(err)
		}
		return ok(map[string]string{"status": "written"})
	case "DownloadFile":
		var in struct {
			Path      string `json:"path"`
			Recursive bool   `json:"recursive"`
		}
		_ = json.Unmarshal(req.Payload, &in)
		base := filepath.Join(s.Root, strings.TrimPrefix(in.Path, "/"))
		if !in.Recursive {
			b, err := os.ReadFile(base)
			if err != nil {
				return bad(err)
			}
			return ok(map[string]string{"content_b64": base64.StdEncoding.EncodeToString(b)})
		}
		files := map[string]string{}
		_ = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			rel, _ := filepath.Rel(s.Root, path)
			b, _ := os.ReadFile(path)
			files["/"+filepath.ToSlash(rel)] = base64.StdEncoding.EncodeToString(b)
			return nil
		})
		return ok(map[string]any{"files": files})
	case "ListProcesses":
		out, _ := exec.Command("ps", "-eo", "pid,comm").Output()
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) > 20 {
			lines = lines[:20]
		}
		return ok(map[string]any{"processes": lines})
	case "GetMachineInfo":
		return ok(map[string]any{"machine_id": s.MachineID, "time": time.Now().UTC()})
	case "StartContainer":
		var in struct{ Name, Image string }
		_ = json.Unmarshal(req.Payload, &in)
		s.mu.Lock()
		s.containers[in.Name] = in.Image
		s.mu.Unlock()
		return ok(map[string]string{"id": in.Name, "status": "running"})
	case "StopContainer":
		var in struct{ Name string }
		_ = json.Unmarshal(req.Payload, &in)
		s.mu.Lock()
		delete(s.containers, in.Name)
		s.mu.Unlock()
		return ok(map[string]string{"status": "stopped"})
	case "ListContainers":
		s.mu.Lock()
		defer s.mu.Unlock()
		out := []string{}
		for n, i := range s.containers {
			out = append(out, n+":"+i)
		}
		return ok(map[string]any{"containers": out})
	case "FetchLogs", "ContainerLogs":
		return ok(map[string]string{"logs": "agent logs not persisted yet"})
	default:
		return bad(fmt.Errorf("unsupported method %s", req.Method))
	}
}
