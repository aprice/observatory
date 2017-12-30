package server

import (
	"net/http"
	"strings"

	"github.com/aprice/observatory"
	"github.com/aprice/observatory/collections"
	"github.com/aprice/observatory/server/config"
)

// ObservatoryMux handles routing Observatory REST API requests using
// compile-time static routing.
type ObservatoryMux struct {
	Conf            *config.Configuration
	acceptedMethods collections.StringSet

	subjectsCrudHandler http.Handler
	checksCrudHandler   http.Handler
	alertsCrudHandler   http.Handler
	periodsCrudHandler  http.Handler
}

// NewObservatoryMux constructs a new ObservatoryMux from the given configuration,
// including static UI serving.
func NewObservatoryMux(conf *config.Configuration) *ObservatoryMux {
	return &ObservatoryMux{
		Conf:                conf,
		subjectsCrudHandler: &crudRouter{subjectsCrud{conf}},
		checksCrudHandler:   &crudRouter{checksCrud{conf}},
		alertsCrudHandler:   &crudRouter{alertsCrud{conf}},
		periodsCrudHandler:  &crudRouter{periodsCrud{conf}},
		acceptedMethods:     collections.NewStringSet("GET", "POST", "PUT", "DELETE", "OPTIONS"),
	}
}

func (m ObservatoryMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.SetHeaders(w, r)
	if !m.acceptedMethods.Contains(r.Method) {
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
		return
	}
	accept := r.Header.Get("Accept")
	if accept != "" && !strings.Contains(accept, "application/json") && !strings.Contains(accept, "*/*") && !strings.Contains(accept, "application/*") && !strings.Contains(accept, "*/json") {
		http.Error(w, "Not Acceptable", http.StatusNotAcceptable)
		return
	}
	parts := strings.Split(r.URL.Path[1:], "/")
	// If there were more routes, a map might be faster than a select.
	// If we need different handlers for sub-paths, just use a nested select
	// or put the logc in the handler. For different HTTP methods, put the logic
	// in the handler.
	switch parts[0] {
	case "up":
		handleUp(w, r, *m.Conf)
	case "info":
		handleInfo(w, r, *m.Conf)
	case "configuration":
		handleConfiguration(w, r, *m.Conf)
	case "peers":
		handlePeers(w, r, *m.Conf)
	case "subjects":
		m.subjectsCrudHandler.ServeHTTP(w, r)
	case "checks":
		m.checksCrudHandler.ServeHTTP(w, r)
	case "alerts":
		m.alertsCrudHandler.ServeHTTP(w, r)
	case "periods":
		m.periodsCrudHandler.ServeHTTP(w, r)
	case "checkresults":
		handleCheckResults(w, r, *m.Conf)
	case "checkstates":
		handleCheckStates(w, r, *m.Conf)
	case "roles":
		handleRoles(w, r, *m.Conf)
	case "tags":
		handleTags(w, r, *m.Conf)
	case "debug":
		handleDebug(w, r)
	default:
		NotFoundResponse(w)
	}
}

// SetHeaders for general-purpose settings
func (m *ObservatoryMux) SetHeaders(w http.ResponseWriter, r *http.Request) {
	// General
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Coordinator-Version", observatory.Version)
	w.Header().Set("X-Coordinator-API-Version", string(observatory.APIVersion))

	// CORS
	if m.Conf.AllowCors != "" {
		w.Header().Set("Access-Control-Allow-Origin", m.Conf.AllowCors)
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	}
}
