package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"virtualpc/cmd/vpc-agent/server"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "server" {
		fs := flag.NewFlagSet("server", flag.ExitOnError)
		socket := fs.String("socket", "/tmp/vpc-agent.sock", "unix socket (vsock bridge)")
		machineID := fs.String("machine-id", "unknown", "machine id")
		root := fs.String("root", "/tmp/vpc-agent-root", "guest fs root")
		_ = fs.Parse(os.Args[2:])
		s := server.New(*socket, *machineID, *root)
		log.Fatal(s.ListenAndServe())
		return
	}
	fmt.Println("usage: vpc-agent server --socket <path> --machine-id <id> --root <dir>")
}
