package api

import (
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(address string, siteActivation SiteActivationHandler, serverOnboarding ServerOnboardingHandler) Server {
	return Server{
		httpServer: &http.Server{
			Addr:              address,
			Handler:           newMux(siteActivation, serverOnboarding),
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

func (s Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func newMux(siteActivation SiteActivationHandler, serverOnboarding ServerOnboardingHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/sites/activate", siteActivation.HandleActivateSite)
	mux.HandleFunc("/v1/servers/onboard", serverOnboarding.HandleOnboardServer)
	mux.HandleFunc("/healthz", handleHealthz)
	return mux
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
