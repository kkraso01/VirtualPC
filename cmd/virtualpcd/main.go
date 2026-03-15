package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"virtualpc/internal/api"
	"virtualpc/internal/config"
	"virtualpc/internal/daemon"
)

func main() {
	cfg := config.Load()
	svc, err := daemon.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	_ = os.Remove(cfg.UnixSocket)
	l, err := net.Listen("unix", cfg.UnixSocket)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.Chmod(cfg.UnixSocket, 0o666); err != nil {
		log.Fatal(err)
	}
	log.Printf("virtualpcd listening on unix://%s", cfg.UnixSocket)
	if err := http.Serve(l, api.New(svc).Handler()); err != nil {
		log.Fatal(err)
	}
}
