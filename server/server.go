package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/aprice/observatory/server/config"
	"github.com/aprice/observatory/ui"
)

type Server struct {
	api    http.Handler
	static http.Handler
}

func New(conf *config.Configuration) *Server {
	return &Server{
		api: http.StripPrefix("/api", &MuxWrapper{
			Mux:  NewObservatoryMux(conf),
			Conf: conf,
			HealthCheck: func() bool {
				healthy := conf.Up
				if healthy {
					ctx, err := conf.ContextFactory.Get()
					if err != nil {
						healthy = false
						log.Print(err)
					}
					ctx.Close()
				}
				return healthy
			},
		}),
		static: ui.GetEmbeddedContent(),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api") {
		s.api.ServeHTTP(w, r)
	} else {
		s.static.ServeHTTP(w, r)
	}
}
