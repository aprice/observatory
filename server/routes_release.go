// +build release

package server

import "net/http"

func handleDebug(w http.ResponseWriter, r *http.Request) {
	NotFoundResponse(w)
}
