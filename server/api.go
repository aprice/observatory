package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/aprice/observatory/model"
	"github.com/aprice/observatory/server/config"
	"github.com/aprice/observatory/utils"
)

// Modifier describes an object that can report the date it was last modified.
type Modifier interface {
	GetModified() time.Time
}

// Start the REST API server.
func Start(conf *config.Configuration) {
	server := New(conf)
	log.Printf("Coordinating observations at %s.", conf.Endpoint())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), server))
}

// OkResponse writes a 200 OK response, with the given payload as JSON. Lifetime
// is used to set caching headers, and if the payload is a Modifier, the Modified
// header will be set.
func OkResponse(w http.ResponseWriter, r *http.Request, payload interface{}, lifetime time.Duration) {
	if lifetime > 0 {
		w.Header().Add("Expires", time.Now().Add(lifetime).UTC().Format(http.TimeFormat))
	}
	if modified, ok := lastModified(payload); ok {
		w.Header().Add("Modified", modified.UTC().Format(http.TimeFormat))
		if ims := r.Header.Get("If-Modified-Since"); ims != "" {
			if lastKnown, err := time.Parse(http.TimeFormat, ims); err != nil {
				if lastKnown.Equal(modified) || lastKnown.After(modified) {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
		}
	}
	json, err := getJSON(w, r, payload)
	if json != nil {
		defer bp.Give(json)
	}
	if err != nil {
		http.Error(w, errorJSON(err), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		json.WriteTo(w)
	}
}

// NoContentResponse writes a No Content response with headers but no payload.
func NoContentResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// CreatedResponse writes a Created response with the given payload and Location
// header.
func CreatedResponse(w http.ResponseWriter, r *http.Request, payload interface{}, location string) {
	json, err := getJSON(w, r, payload)
	if json != nil {
		defer bp.Give(json)
	}
	if err != nil {
		http.Error(w, errorJSON(err), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusCreated)
		json.WriteTo(w)
	}
}

// OptionsResponse writes a 200 OK response to an OPTIONS request, including the
// Allow header, with the given payload as JSON. The payload should describe
// usage of the route.
func OptionsResponse(w http.ResponseWriter, r *http.Request, methods []string, payload interface{}) {
	w.Header().Set("Allow", strings.Join(methods, ", "))
	json, err := getJSON(w, r, payload)
	if json != nil {
		defer bp.Give(json)
	}
	if err != nil {
		http.Error(w, errorJSON(err), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		json.WriteTo(w)
	}
}

// NotAllowedResponse writes a Method Not Allowed response, including the
// Allow header.
func NotAllowedResponse(w http.ResponseWriter, methods []string) {
	w.Header().Set("Allow", strings.Join(methods, ", "))
	w.Header().Set("Content-Type", "text/plain")
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// ErrorResponse writes an Internal Server Error response, or a Not Found
// response if the error given is model.ErrNotFound.
func ErrorResponse(w http.ResponseWriter, err error) {
	if err == model.ErrNotFound {
		NotFoundResponse(w)
	} else {
		http.Error(w, errorMessageJSON("Internal Server Error: "+err.Error()), http.StatusInternalServerError)
	}
}

// NotFoundResponse writes a Not Found error response.
func NotFoundResponse(w http.ResponseWriter) {
	http.Error(w, errorMessageJSON("Not Found"), http.StatusNotFound)
}

// BadRequestResponse writes a Bad Request response, including error message.
func BadRequestResponse(w http.ResponseWriter, err error) {
	http.Error(w, errorMessageJSON("Bad Request: "+err.Error()), http.StatusBadRequest)
}

// NotImplementedResponse writes a Not Yet Implemented response.
func NotImplementedResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	http.Error(w, "Not Yet Implemented", http.StatusNotImplemented)
}

// TODO: Global, hard-coded settings, arbitrary settings, bad bad bad
var bp = utils.NewBufferPool(2048, 4096, 100)

func getJSON(w http.ResponseWriter, r *http.Request, payload interface{}) (*bytes.Buffer, error) {
	b := bp.Get()
	err := json.NewEncoder(b).Encode(payload)
	if err != nil {
		b.WriteString(err.Error())
		return b, err
	}
	cb := r.URL.Query().Get("callback")
	if cb == "" {
		return b, nil
	}
	bb := bp.Get()
	bb.WriteString(cb)
	bb.WriteByte('(')
	bb.ReadFrom(b)
	bb.Write([]byte{')', ';'})
	bp.Give(b)
	return bb, nil
}

func errorJSON(err error) string {
	return errorMessageJSON(err.Error())
}

func errorMessageJSON(err string) string {
	return fmt.Sprintf("{\"error\": %q}", err)
}

func pathPart(r *http.Request, index int) string {
	parts := strings.Split(r.URL.Path[1:], "/")
	if index < len(parts) {
		return parts[index]
	}
	return ""
}

func countPathParts(r *http.Request) int {
	return strings.Count(r.URL.Path[1:], "/")
}

func lastModified(i interface{}) (time.Time, bool) {
	if m, ok := i.(Modifier); ok {
		return m.GetModified(), true
	}

	val := reflect.ValueOf(i)
	if (val.Kind() == reflect.Array || val.Kind() == reflect.Slice) && val.Len() > 0 {
		if _, ok := val.Index(0).Interface().(Modifier); !ok {
			return time.Now(), false
		}

		modifieds := make([]time.Time, 0, val.Len())
		for i := 0; i < val.Len(); i++ {
			e := val.Index(i).Interface()
			if m, ok := e.(Modifier); ok {
				modifieds = append(modifieds, m.GetModified())
			}
		}
		if len(modifieds) > 0 {
			return utils.LatestDate(modifieds...), true
		}
	}
	return time.Now(), false
}
