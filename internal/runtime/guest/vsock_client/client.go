package vsock_client

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
	"virtualpc/internal/runtime/guest/protocol"
)

type Client struct {
	conn net.Conn
	r    *bufio.Reader
	w    *bufio.Writer
}

func New(socketPath string) (*Client, error) {
	c, err := net.DialTimeout("unix", socketPath, 3*time.Second)
	if err != nil {
		return nil, err
	}
	return &Client{conn: c, r: bufio.NewReader(c), w: bufio.NewWriter(c)}, nil
}
func (c *Client) Close() error { return c.conn.Close() }

func (c *Client) call(method string, payload any, out any) error {
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	b, _ := json.Marshal(payload)
	req := protocol.Request{ID: id, Method: method, Payload: b}
	rb, _ := json.Marshal(req)
	if _, err := c.w.Write(append(rb, '\n')); err != nil {
		return err
	}
	if err := c.w.Flush(); err != nil {
		return err
	}
	line, err := c.r.ReadBytes('\n')
	if err != nil {
		return err
	}
	var resp protocol.Response
	if err := json.Unmarshal(line, &resp); err != nil {
		return err
	}
	if !resp.OK {
		return errors.New(resp.Error)
	}
	if out != nil {
		return json.Unmarshal(resp.Payload, out)
	}
	return nil
}

func (c *Client) ExecCommand(command []string) (string, error) {
	var o struct {
		Output string `json:"output"`
	}
	err := c.call("ExecCommand", map[string]any{"command": command}, &o)
	return o.Output, err
}
func (c *Client) OpenPTY() error { return c.call("OpenPTY", map[string]any{}, nil) }
func (c *Client) ListProcesses() ([]string, error) {
	var o struct {
		Processes []string `json:"processes"`
	}
	err := c.call("ListProcesses", map[string]any{}, &o)
	return o.Processes, err
}
func (c *Client) FetchLogs(lines int) (string, error) {
	var o struct {
		Logs string `json:"logs"`
	}
	err := c.call("FetchLogs", map[string]int{"lines": lines}, &o)
	return o.Logs, err
}
func (c *Client) StartContainer(name, image string) (string, error) {
	var o struct {
		ID string `json:"id"`
	}
	err := c.call("StartContainer", map[string]string{"name": name, "image": image}, &o)
	return o.ID, err
}
func (c *Client) StopContainer(name string) error {
	return c.call("StopContainer", map[string]string{"name": name}, nil)
}
func (c *Client) ListContainers() ([]string, error) {
	var o struct {
		Containers []string `json:"containers"`
	}
	err := c.call("ListContainers", map[string]any{}, &o)
	return o.Containers, err
}
func (c *Client) ContainerLogs(name string) (string, error) {
	var o struct {
		Logs string `json:"logs"`
	}
	err := c.call("ContainerLogs", map[string]string{"name": name}, &o)
	return o.Logs, err
}

func (c *Client) Upload(src, dst string, recursive bool) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		if !recursive {
			return errors.New("source is directory; use recursive")
		}
		return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(src, path)
			return c.uploadFile(path, filepath.Join(dst, rel))
		})
	}
	return c.uploadFile(src, dst)
}
func (c *Client) uploadFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return c.call("UploadFile", map[string]string{"path": filepath.ToSlash(dst), "content_b64": base64.StdEncoding.EncodeToString(b)}, nil)
}

func (c *Client) Download(src, dst string, recursive bool) error {
	if recursive {
		var o struct {
			Files map[string]string `json:"files"`
		}
		if err := c.call("DownloadFile", map[string]any{"path": src, "recursive": true}, &o); err != nil {
			return err
		}
		for path, b64 := range o.Files {
			target := filepath.Join(dst, strings.TrimPrefix(path, "/"))
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			b, _ := base64.StdEncoding.DecodeString(b64)
			if err := os.WriteFile(target, b, 0o644); err != nil {
				return err
			}
		}
		return nil
	}
	var o struct {
		ContentB64 string `json:"content_b64"`
	}
	if err := c.call("DownloadFile", map[string]any{"path": src}, &o); err != nil {
		return err
	}
	b, _ := base64.StdEncoding.DecodeString(o.ContentB64)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}
