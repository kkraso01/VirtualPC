package main

import (
	"encoding/json"
	"fmt"
	"os"

	"virtualpc/internal/cli"
)

func main() {
	sock := env("VPC_UNIX_SOCKET", "/tmp/virtualpcd.sock")
	c := cli.New(sock)
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}
	must := func(err error) {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	print := func(v any) { b, _ := json.MarshalIndent(v, "", "  "); fmt.Println(string(b)) }

	switch args[0] {
	case "daemon":
		if len(args) > 1 && args[1] == "status" {
			var out any
			must(c.Do("GET", "/v1/daemon/status", nil, &out))
			print(out)
			return
		}
	case "config":
		if len(args) > 1 && args[1] == "inspect" {
			var out any
			must(c.Do("GET", "/v1/config", nil, &out))
			print(out)
			return
		}
	case "profile":
		if len(args) > 1 && args[1] == "list" {
			var out any
			must(c.Do("GET", "/v1/profiles", nil, &out))
			print(out)
			return
		}
		if len(args) > 2 && args[1] == "inspect" {
			var out any
			must(c.Do("GET", "/v1/profiles/"+args[2], nil, &out))
			print(out)
			return
		}
	case "machine":
		handleMachine(c, args[1:], must, print)
		return
	case "project":
		handleProject(c, args[1:], must, print)
		return
	case "service":
		handleService(c, args[1:], must, print)
		return
	case "snapshot":
		handleSnapshot(c, args[1:], must, print)
		return
	case "task":
		handleTask(c, args[1:], must, print)
		return
	}
	usage()
	os.Exit(1)
}

func handleMachine(c *cli.Client, a []string, must func(error), print func(any)) {
	if len(a) == 0 {
		usage()
		os.Exit(1)
	}
	switch a[0] {
	case "create":
		profile := "minimal-shell"
		for i := 1; i < len(a); i++ {
			if a[i] == "--profile" && i+1 < len(a) {
				profile = a[i+1]
			}
		}
		var out any
		must(c.Do("POST", "/v1/machines", map[string]string{"profile": profile}, &out))
		print(out)
	case "list":
		var out any
		must(c.Do("GET", "/v1/machines", nil, &out))
		print(out)
	case "inspect":
		var out any
		must(c.Do("GET", "/v1/machines/"+a[1], nil, &out))
		print(out)
	case "start":
		var out any
		must(c.Do("POST", "/v1/machines/"+a[1]+"/start", map[string]string{}, &out))
		print(out)
	case "stop":
		var out any
		must(c.Do("POST", "/v1/machines/"+a[1]+"/stop", map[string]string{}, &out))
		print(out)
	case "destroy":
		var out any
		must(c.Do("POST", "/v1/machines/"+a[1]+"/destroy", map[string]string{}, &out))
		print(out)
	case "exec":
		idx := indexOf(a, "--")
		var out any
		must(c.Do("POST", "/v1/machines/"+a[1]+"/exec", map[string]any{"command": a[idx+1:]}, &out))
		print(out)
	case "shell":
		var out any
		must(c.Do("POST", "/v1/machines/"+a[1]+"/shell", map[string]any{}, &out))
		print(out)
	case "logs":
		var out any
		must(c.Do("POST", "/v1/machines/"+a[1]+"/logs", map[string]string{}, &out))
		print(out)
	case "ps":
		var out any
		must(c.Do("POST", "/v1/machines/"+a[1]+"/ps", map[string]string{}, &out))
		print(out)
	case "assign":
		project := ""
		for i := 2; i < len(a); i++ {
			if a[i] == "--project" && i+1 < len(a) {
				project = a[i+1]
			}
		}
		var out any
		must(c.Do("POST", "/v1/assign", map[string]string{"machine_id": a[1], "project_id": project}, &out))
		print(out)
	case "fork":
		var out any
		must(c.Do("POST", "/v1/fork", map[string]string{"snapshot_id": a[1]}, &out))
		print(out)
	case "cp-to":
		if len(a) < 4 {
			usage()
			os.Exit(1)
		}
		var out any
		must(c.Do("POST", "/v1/machines/"+a[1]+"/cp-to", map[string]any{"src": a[2], "dst": a[3], "recursive": hasFlag(a, "-r")}, &out))
		print(out)
	case "cp-from":
		if len(a) < 4 {
			usage()
			os.Exit(1)
		}
		var out any
		must(c.Do("POST", "/v1/machines/"+a[1]+"/cp-from", map[string]any{"src": a[2], "dst": a[3], "recursive": hasFlag(a, "-r")}, &out))
		print(out)
	default:
		usage()
		os.Exit(1)
	}
}

func handleProject(c *cli.Client, a []string, must func(error), print func(any)) {
	if len(a) == 0 {
		usage()
		os.Exit(1)
	}
	switch a[0] {
	case "create":
		var out any
		must(c.Do("POST", "/v1/projects", map[string]string{"name": a[1]}, &out))
		print(out)
	case "list":
		var out any
		must(c.Do("GET", "/v1/projects", nil, &out))
		print(out)
	default:
		usage()
		os.Exit(1)
	}
}
func handleService(c *cli.Client, a []string, must func(error), print func(any)) {
	if len(a) == 0 {
		usage()
		os.Exit(1)
	}
	switch a[0] {
	case "create":
		machine, name, image := "", "", ""
		for i := 1; i < len(a); i++ {
			if a[i] == "--machine" && i+1 < len(a) {
				machine = a[i+1]
			}
			if a[i] == "--name" && i+1 < len(a) {
				name = a[i+1]
			}
			if a[i] == "--image" && i+1 < len(a) {
				image = a[i+1]
			}
		}
		var out any
		must(c.Do("POST", "/v1/services", map[string]string{"MachineID": machine, "Name": name, "Image": image}, &out))
		print(out)
	case "list":
		machine := ""
		for i := 1; i < len(a); i++ {
			if a[i] == "--machine" && i+1 < len(a) {
				machine = a[i+1]
			}
		}
		var out any
		must(c.Do("GET", "/v1/services?machine_id="+machine, nil, &out))
		print(out)
	case "logs", "stop", "destroy":
		if len(a) < 2 {
			usage()
			os.Exit(1)
		}
		var out any
		must(c.Do("POST", "/v1/services/"+a[1]+"/"+a[0], map[string]any{}, &out))
		print(out)
	default:
		usage()
		os.Exit(1)
	}
}
func handleSnapshot(c *cli.Client, a []string, must func(error), print func(any)) {
	if len(a) == 0 {
		usage()
		os.Exit(1)
	}
	switch a[0] {
	case "create":
		var out any
		must(c.Do("POST", "/v1/snapshots", map[string]string{"machine_id": a[1]}, &out))
		print(out)
	case "list":
		var out any
		must(c.Do("GET", "/v1/snapshots", nil, &out))
		print(out)
	case "inspect":
		var out any
		must(c.Do("GET", "/v1/snapshots/"+a[1], nil, &out))
		print(out)
	default:
		usage()
		os.Exit(1)
	}
}
func handleTask(c *cli.Client, a []string, must func(error), print func(any)) {
	if len(a) == 0 {
		usage()
		os.Exit(1)
	}
	switch a[0] {
	case "create":
		machine, goal := "", ""
		for i := 1; i < len(a); i++ {
			if a[i] == "--machine" && i+1 < len(a) {
				machine = a[i+1]
			}
			if a[i] == "--goal" && i+1 < len(a) {
				goal = a[i+1]
			}
		}
		var out any
		must(c.Do("POST", "/v1/tasks", map[string]string{"MachineID": machine, "Goal": goal}, &out))
		print(out)
	case "run":
		var out any
		must(c.Do("POST", "/v1/tasks/run", map[string]string{"task_id": a[1]}, &out))
		print(out)
	case "inspect", "logs":
		var out any
		must(c.Do("GET", "/v1/tasks/"+a[1], nil, &out))
		print(out)
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("vpc daemon status | config inspect | profile list | machine ... | project ... | service ... | snapshot ... | task ...")
}
func env(k, f string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return f
}
func indexOf(a []string, want string) int {
	for i, v := range a {
		if v == want {
			return i
		}
	}
	return len(a) - 1
}

func hasFlag(a []string, flag string) bool {
	for _, v := range a {
		if v == flag {
			return true
		}
	}
	return false
}
