package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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
	case "list":
		listSessions()
	case "schemas":
		printJSON(tools.Catalog())
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
	providerName := fs.String("provider", "", "openai|anthropic|openai_compatible|ollama|vllm")
	approval := fs.Bool("approval", false, "require operator approval for dangerous operations")
	_ = fs.Parse(args)
	if *machineID == "" || *goal == "" {
		fmt.Fprintln(os.Stderr, "--machine and --goal are required")
		os.Exit(2)
	}
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		cfg = config.Default()
	}
	if *providerName != "" {
		cfg.Provider = *providerName
	}
	session := &controller.Session{SessionID: fmt.Sprintf("s-%d", time.Now().Unix()), MachineID: *machineID, Goal: *goal, Status: "created", Provider: cfg.Provider, Model: cfg.Model}
	if err := session.Save(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctr, err := controller.New(cfg, cli.New(*sock), chooseProvider(cfg), "agent/prompts/system_prompt.txt", *approval, controller.SessionLogPath(session.SessionID))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer ctr.Close()
	if err := ctr.Run(session); err != nil {
		fmt.Fprintf(os.Stderr, "session stopped: %v\n", err)
	}
	_ = session.Save()
	printJSON(map[string]any{"session_id": session.SessionID, "status": session.Status, "log_path": controller.SessionLogPath(session.SessionID)})
}

func chooseProvider(cfg config.Config) providers.Provider {
	switch cfg.Provider {
	case "anthropic":
		return providers.NewAnthropic(cfg.Model)
	case "openai", "":
		return providers.NewOpenAI(cfg.Model)
	case "openai_compatible", "ollama", "vllm":
		return providers.NewOpenAICompatible(cfg.Model, cfg.BaseURL, cfg.APIKey, cfg.ProviderCapabilities())
	default:
		return nil
	}
}

func attach(args []string) {
	s, err := controller.LoadSession(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	printJSON(s)
}
func logs(args []string) {
	b, err := os.ReadFile(controller.SessionLogPath(args[0]))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(string(b))
}
func stop(args []string) {
	s, err := controller.LoadSession(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	s.StopRequested = true
	s.Status = "stop_requested"
	_ = s.Save()
	printJSON(map[string]any{"session_id": s.SessionID, "status": s.Status})
}
func listSessions() {
	sessions, err := controller.ListSessions()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	printJSON(sessions)
}
func usage() {
	fmt.Println("vpc-agent-controller start --machine <id> --goal <text> [--provider openai|anthropic|openai_compatible|ollama|vllm] [--config path]")
	fmt.Println("vpc-agent-controller attach <session-id>")
	fmt.Println("vpc-agent-controller logs <session-id>")
	fmt.Println("vpc-agent-controller stop <session-id>")
	fmt.Println("vpc-agent-controller list")
	fmt.Println("vpc-agent-controller schemas")
}
func env(k, f string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return f
}
func printJSON(v any) { b, _ := json.MarshalIndent(v, "", "  "); fmt.Println(string(b)) }
