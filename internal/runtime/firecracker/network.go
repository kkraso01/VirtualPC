package firecracker

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type NetworkMode string

const (
	NetworkModeOffline   NetworkMode = "offline"
	NetworkModeNAT       NetworkMode = "nat"
	NetworkModeAllowlist NetworkMode = "allowlist"
)

type NetworkConfig struct {
	MachineID string   `json:"machine_id"`
	TapName   string   `json:"tap_name"`
	Mode      string   `json:"mode"`
	Allowlist []string `json:"allowlist,omitempty"`
	NAT       bool     `json:"nat"`
}

type NetworkManager struct{ baseDir string }

func NewNetworkManager(baseDir string) *NetworkManager { return &NetworkManager{baseDir: baseDir} }

func (n *NetworkManager) Setup(machineID string, mode NetworkMode, allowlist []string) error {
	cfg := NetworkConfig{MachineID: machineID, TapName: "tap-" + machineID[:8], Mode: string(mode), Allowlist: allowlist, NAT: mode == NetworkModeNAT}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(filepath.Join(n.baseDir, machineID, "network.json"), b, 0o644)
}

func (n *NetworkManager) Cleanup(machineID string) error {
	_ = os.Remove(filepath.Join(n.baseDir, machineID, "network.json"))
	return nil
}
