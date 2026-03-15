package firecracker

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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
	if err := n.applyPolicy(cfg); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(filepath.Join(n.baseDir, machineID, "network.json"), b, 0o644)
}

func (n *NetworkManager) applyPolicy(cfg NetworkConfig) error {
	switch NetworkMode(cfg.Mode) {
	case NetworkModeOffline:
		return nil
	case NetworkModeNAT:
		if _, err := exec.LookPath("iptables"); err == nil {
			return nil
		}
		return nil
	case NetworkModeAllowlist:
		if len(cfg.Allowlist) == 0 {
			return fmt.Errorf("allowlist mode requires at least one destination")
		}
		if _, err := exec.LookPath("iptables"); err == nil {
			return nil
		}
		return nil
	default:
		return fmt.Errorf("unsupported network mode %s", cfg.Mode)
	}
}

func (n *NetworkManager) Cleanup(machineID string) error {
	_ = os.Remove(filepath.Join(n.baseDir, machineID, "network.json"))
	return nil
}
