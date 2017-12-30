package server

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aprice/observatory"
	"github.com/aprice/observatory/server/config"
	"github.com/aprice/observatory/utils"
)

// MuxWrapper wraps an http.Handler, typically a mux, and adds CORS support,
// GZIP compression, logging, health checking, and basic headers.
type MuxWrapper struct {
	Mux         http.Handler
	Conf        *config.Configuration
	HealthCheck func() bool
}

var serverHeader = "Observatory Coordinator " + observatory.Version

func (mw MuxWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t1 := time.Now()

	ww := &ResponseWriterWrapper{Writer: w, ResponseWriter: w}
	defer ww.GzipWrap(r.Header.Get("Accept-Encoding"), r.URL.Path)()

	ww.Header().Set("Server", serverHeader)

	if mw.Conf.Up && mw.HealthCheck() {
		mw.Mux.ServeHTTP(ww, r)
	} else {
		w.Header().Set("Content-Type", "text/plain")
		http.Error(w, "The server is down.", http.StatusServiceUnavailable)
	}

	t2 := time.Now()
	log.Printf("%s %q: %d %s %s;%s %v",
		r.Method,
		r.URL.String(),
		ww.Status,
		utils.HumanReadableBytesSI(int64(ww.Length), 2),
		ww.Header().Get("Content-Encoding"),
		ww.Header().Get("Content-Type"),
		t2.Sub(t1))
}

// ResponseWriterWrapper wraps an http.ResponseWriter and provides writer
// wrapping and recording of response status and length.
type ResponseWriterWrapper struct {
	io.Writer
	http.ResponseWriter

	Status int
	Length int
}

// WriteHeader implements http.ResponseWriter.WriteHeader.
func (rw *ResponseWriterWrapper) WriteHeader(status int) {
	rw.Status = status
	rw.ResponseWriter.WriteHeader(status)
}

// Write implements io.Writer.Write.
func (rw *ResponseWriterWrapper) Write(b []byte) (int, error) {
	written, err := rw.Writer.Write(b)
	rw.Length += written
	return written, err
}

// GzipWrap a response writer if the client supports it. Returned function is
// the Gzip stream closer.
func (rw *ResponseWriterWrapper) GzipWrap(acceptEncoding, path string) func() {
	if strings.Contains(acceptEncoding, "gzip") && !strings.HasSuffix(path, ".jpg") {
		rw.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(rw.ResponseWriter)
		rw.Writer = gz
		return func() { gz.Close() }
	}
	return func() {}
}
