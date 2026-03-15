package config

import "os"

type Config struct {
	UnixSocket     string
	DataPath       string
	FirecrackerDir string
	PostgresDSN    string
	NATSURL        string
	MinIOEndpoint  string
	TemporalHost   string
}

func Load() Config {
	cfg := Config{
		UnixSocket:     get("VPCD_UNIX_SOCKET", "/tmp/virtualpcd.sock"),
		DataPath:       get("VPCD_DATA_PATH", "./data/state.json"),
		FirecrackerDir: get("VPCD_FIRECRACKER_DIR", "./data/firecracker"),
		PostgresDSN:    get("VPCD_POSTGRES_DSN", "postgres://virtualpc:virtualpc@127.0.0.1:5432/virtualpc?sslmode=disable"),
		NATSURL:        get("VPCD_NATS_URL", "nats://127.0.0.1:4222"),
		MinIOEndpoint:  get("VPCD_MINIO_ENDPOINT", "127.0.0.1:9000"),
		TemporalHost:   get("VPCD_TEMPORAL_HOSTPORT", "127.0.0.1:7233"),
	}
	return cfg
}

func get(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
