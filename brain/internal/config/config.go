package config

import "os"

type Config struct {
	HTTPAddress    string
	AgentTransport string
	AgentAddress   string
}

func Load() Config {
	httpAddress := os.Getenv("KARI_BRAIN_HTTP_ADDR")
	if httpAddress == "" {
		httpAddress = ":8080"
	}

	agentTransport := os.Getenv("KARI_BRAIN_AGENT_TRANSPORT")
	if agentTransport == "" {
		agentTransport = "inmemory"
	}

	agentAddress := os.Getenv("KARI_BRAIN_AGENT_ADDR")
	if agentAddress == "" {
		agentAddress = "127.0.0.1:9090"
	}

	return Config{
		HTTPAddress:    httpAddress,
		AgentTransport: agentTransport,
		AgentAddress:   agentAddress,
	}
}
