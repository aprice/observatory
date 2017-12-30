// +build !release

package server

import (
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"
	"time"
)

var goroutineCounter *expvar.Int
var uptime *expvar.Int
var launchTime time.Time

func init() {
	goroutineCounter = expvar.NewInt("Goroutines")
	uptime = expvar.NewInt("Uptime")
	launchTime = time.Now()
}

func handleDebug(w http.ResponseWriter, r *http.Request) {
	switch pathPart(r, 1) {
	case "vars":
		handleDebugVars(w, r)
	case "pprof":
		handlePprof(w, r)
	default:
		NotFoundResponse(w)
	}
}

func handleDebugVars(w http.ResponseWriter, r *http.Request) {
	goroutineCounter.Set(int64(runtime.NumGoroutine()))
	uptime.Set(int64(time.Now().Sub(launchTime)))
	fmt.Fprint(w, "{")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprint(w, ",")
		}
		first = false
		fmt.Fprintf(w, "%q:%s", kv.Key, kv.Value)
	})
	fmt.Fprint(w, "}")
}

func handlePprof(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	switch pathPart(r, 2) {
	case "cmdline":
		pprof.Cmdline(w, r)
	case "profile":
		pprof.Profile(w, r)
	case "symbol":
		pprof.Symbol(w, r)
	case "trace":
		pprof.Trace(w, r)
	default:
		pprof.Index(w, r)
	}
}
