package models

import "time"

type MachineState string

const (
	MachinePending  MachineState = "pending"
	MachineBooting  MachineState = "booting"
	MachineRunning  MachineState = "running"
	MachineStopping MachineState = "stopping"
	MachineStopped  MachineState = "stopped"
	MachineFailed   MachineState = "failed"
	MachineDeleted  MachineState = "deleted"
)

type NetworkMode string

const (
	NetworkOffline   NetworkMode = "offline"
	NetworkNAT       NetworkMode = "nat"
	NetworkAllowlist NetworkMode = "allowlist"
)

type MachineProfile struct {
	Name              string      `json:"name"`
	BaseImage         string      `json:"base_image"`
	VCPU              int         `json:"vcpu"`
	MemoryMB          int         `json:"memory_mb"`
	DiskGB            int         `json:"disk_gb"`
	BrowserEnabled    bool        `json:"browser_enabled"`
	ContainerdEnabled bool        `json:"containerd_enabled"`
	NetworkMode       NetworkMode `json:"network_mode"`
	PolicyClass       string      `json:"policy_class"`
	Allowlist         []string    `json:"allowlist,omitempty"`
}

type Machine struct {
	ID          string       `json:"id"`
	ProfileName string       `json:"profile_name"`
	ProjectID   string       `json:"project_id,omitempty"`
	State       MachineState `json:"state"`
	RuntimeID   string       `json:"runtime_id,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Snapshot struct {
	ID        string    `json:"id"`
	MachineID string    `json:"machine_id"`
	ParentID  string    `json:"parent_id,omitempty"`
	DiskRef   string    `json:"disk_ref"`
	Metadata  string    `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}

type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Service struct {
	ID        string    `json:"id"`
	MachineID string    `json:"machine_id"`
	Name      string    `json:"name"`
	Image     string    `json:"image"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Task struct {
	ID        string    `json:"id"`
	MachineID string    `json:"machine_id"`
	Goal      string    `json:"goal"`
	Status    string    `json:"status"`
	Logs      []string  `json:"logs"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuditEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	SubjectID string    `json:"subject_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type DaemonState struct {
	Machines    map[string]Machine        `json:"machines"`
	Profiles    map[string]MachineProfile `json:"profiles"`
	Snapshots   map[string]Snapshot       `json:"snapshots"`
	Projects    map[string]Project        `json:"projects"`
	Services    map[string]Service        `json:"services"`
	Tasks       map[string]Task           `json:"tasks"`
	AuditEvents []AuditEvent              `json:"audit_events"`
	StartedAt   time.Time                 `json:"started_at"`
}
