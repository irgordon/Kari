package app

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/kari/brain/api"
	"github.com/kari/brain/internal/config"
	"github.com/kari/brain/internal/ports"
	"github.com/kari/brain/internal/usecase"
	repository "github.com/kari/brain/models/inmemory"
	"github.com/kari/brain/pipelines"
	grpcprovider "github.com/kari/brain/providers/grpc"
	"github.com/kari/brain/providers/inmemory"
)

func Run() error {
	cfg := config.Load()
	server, err := newHTTPServer(cfg)
	if err != nil {
		return err
	}
	err = server.Start()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func newHTTPServer(cfg config.Config) (api.Server, error) {
	agent, systemChecker, err := resolveAgentClients(cfg)
	if err != nil {
		return api.Server{}, err
	}
	dnsProvider := inmemory.NewDNSProvider()
	siteRepository := repository.NewSiteRepository()

	pipeline := pipelines.NewSiteActivationPipeline(agent, dnsProvider, siteRepository)
	siteActivationService := usecase.NewSiteActivationService(pipeline)
	serverOnboardingService := usecase.NewServerOnboardingService(systemChecker)

	siteActivationHandler := api.NewSiteActivationHandler(siteActivationService)
	serverOnboardingHandler := api.NewServerOnboardingHandler(serverOnboardingService)
	return api.NewServer(cfg.HTTPAddress, siteActivationHandler, serverOnboardingHandler), nil
}

func resolveAgentClients(cfg config.Config) (ports.Agent, ports.SystemChecker, error) {
	if cfg.AgentTransport == "grpc" {
		client := grpcprovider.NewClient(cfg.AgentAddress)
		return client, client, nil
	}
	if cfg.AgentTransport == "inmemory" {
		agent := inmemory.NewAgent()
		systemChecker := inmemory.NewSystemChecker()
		return agent, systemChecker, nil
	}
	return nil, nil, fmt.Errorf("unsupported KARI_BRAIN_AGENT_TRANSPORT: %s", cfg.AgentTransport)
}
