package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"virtualpc/agent/config"
	"virtualpc/agent/controller"
	"virtualpc/agent/providers"
	"virtualpc/agent/tools"
	"virtualpc/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "start":
		start(os.Args[2:])
	case "attach":
		attach(os.Args[2:])
	case "logs":
		logs(os.Args[2:])
	case "stop":
		stop(os.Args[2:])
	case "schemas":
		printSchemas()
	default:
		usage()
		os.Exit(1)
	}
}

func start(args []string) {
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	machineID := fs.String("machine", "", "machine id")
	goal := fs.String("goal", "", "task goal")
	sock := fs.String("socket", env("VPC_UNIX_SOCKET", "/tmp/virtualpcd.sock"), "vpc unix socket")
	cfgPath := fs.String("config", "agent/config/agent_config.yaml", "agent config file")
	providerName := fs.String("provider", "", "openai|anthropic")
	approval := fs.Bool("approval", false, "require operator approval for dangerous operations")
	_ = fs.Parse(args)
	if *machineID == "" || *goal == "" {
		fmt.Fprintln(os.Stderr, "--machine and --goal are required")
		os.Exit(2)
	}
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load config: %v\n", err)
		cfg = config.Default()
	}
	if *providerName != "" {
		cfg.Provider = *providerName
	}
	session := &controller.Session{ID: fmt.Sprintf("s-%d", time.Now().Unix()), MachineID: *machineID, Goal: *goal, Status: "created"}
	if err := session.Save(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	logPath := controller.SessionLogPath(session.ID)
	ctr, err := controller.New(cfg, cli.New(*sock), chooseProvider(cfg), "agent/prompts/system_prompt.txt", *approval, logPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer ctr.Close()
	if err := ctr.Run(session); err != nil {
		fmt.Fprintf(os.Stderr, "session stopped: %v\n", err)
	}
	_ = session.Save()
	printJSON(map[string]any{"session_id": session.ID, "status": session.Status, "log_path": logPath})
}

func chooseProvider(cfg config.Config) providers.Provider {
	switch cfg.Provider {
	case "anthropic":
		return providers.NewAnthropic(cfg.Model)
	case "openai":
		return providers.NewOpenAI(cfg.Model)
	default:
		return nil
	}
}

func attach(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "attach requires <session-id>")
		os.Exit(2)
	}
	s, err := controller.LoadSession(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	printJSON(s)
}

func logs(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "logs requires <session-id>")
		os.Exit(2)
	}
	b, err := os.ReadFile(controller.SessionLogPath(args[0]))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(string(b))
}

func stop(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "stop requires <session-id>")
		os.Exit(2)
	}
	s, err := controller.LoadSession(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	s.Status = "stopped-by-operator"
	s.UpdatedAt = time.Now()
	if err := s.Save(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	printJSON(map[string]string{"session_id": s.ID, "status": s.Status})
}

func printSchemas() {
	printJSON(map[string]any{"tools": tools.Catalog()})
}

func usage() {
	fmt.Println("vpc-agent-controller start --machine <id> --goal <text> [--provider openai|anthropic] [--approval]")
	fmt.Println("vpc-agent-controller attach <session-id>")
	fmt.Println("vpc-agent-controller logs <session-id>")
	fmt.Println("vpc-agent-controller stop <session-id>")
	fmt.Println("vpc-agent-controller schemas")
}

func printJSON(v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func init() {
	_ = os.MkdirAll(filepath.Join(os.Getenv("HOME"), ".virtualpc", "agent", "sessions"), 0o755)
}
