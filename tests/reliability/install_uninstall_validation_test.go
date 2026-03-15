package reliability

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestInstallUninstallScriptsValidation(t *testing.T) {
	root := filepath.Clean("../..")
	tmp := t.TempDir()
	fakebin := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(fakebin, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"systemctl", "firecracker"} {
		p := filepath.Join(fakebin, n)
		if err := os.WriteFile(p, []byte("#!/usr/bin/env bash\nexit 0\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	buildCmd := exec.Command("bash", "scripts/build-binaries.sh")
	buildCmd.Dir = root
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build-binaries failed: %v\n%s", err, string(out))
	}
	assetsDir := filepath.Join(root, "data", "firecracker")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "vmlinux"), []byte("test-kernel"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "rootfs.ext4"), []byte("test-rootfs"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Remove(filepath.Join(assetsDir, "vmlinux"))
		_ = os.Remove(filepath.Join(assetsDir, "rootfs.ext4"))
	})

	installCmd := exec.Command("bash", "scripts/install.sh")
	installCmd.Dir = root
	installCmd.Env = append(os.Environ(),
		"PATH="+fakebin+":"+os.Getenv("PATH"),
		"VPC_INSTALL_SKIP_KVM_CHECK=1",
		"SYSTEMCTL_BIN=systemctl",
		"VPC_INSTALL_PREFIX="+filepath.Join(tmp, "opt"),
		"VPC_ETC_DIR="+filepath.Join(tmp, "etc"),
		"VPC_RUN_DIR="+filepath.Join(tmp, "run"),
		"VPC_DATA_DIR="+filepath.Join(tmp, "data"),
	)
	if out, err := installCmd.CombinedOutput(); err != nil {
		t.Fatalf("install script failed: %v\n%s", err, string(out))
	}

	uninstallCmd := exec.Command("bash", "scripts/uninstall.sh")
	uninstallCmd.Dir = root
	uninstallCmd.Env = append(os.Environ(),
		"PATH="+fakebin+":"+os.Getenv("PATH"),
		"SYSTEMCTL_BIN=systemctl",
		"VPC_INSTALL_PREFIX="+filepath.Join(tmp, "opt"),
		"VPC_ETC_DIR="+filepath.Join(tmp, "etc"),
		"VPC_RUN_DIR="+filepath.Join(tmp, "run"),
		"VPC_DATA_DIR="+filepath.Join(tmp, "data"),
		"VPC_PURGE_STATE=1",
	)
	if out, err := uninstallCmd.CombinedOutput(); err != nil {
		t.Fatalf("uninstall script failed: %v\n%s", err, string(out))
	}
}
