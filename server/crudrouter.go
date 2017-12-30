package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aprice/observatory/model"
	"github.com/aprice/observatory/utils"
	uuid "github.com/satori/go.uuid"
)

type crudHandler interface {
	entity() interface{}
	search(w http.ResponseWriter, r *http.Request) (interface{}, error)
	create(w http.ResponseWriter, r *http.Request, entity interface{}) (string, error)
	retrieve(w http.ResponseWriter, r *http.Request, id uuid.UUID) (interface{}, error)
	update(w http.ResponseWriter, r *http.Request, id uuid.UUID, entity interface{}) error
	delete(w http.ResponseWriter, r *http.Request, id uuid.UUID) error
}

type crudRouter struct {
	handler crudHandler
}

func (cr crudRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		entity interface{}
		l      string
	)
	parts := strings.Split(r.URL.Path[1:], "/")
	numParts := len(parts)
	if numParts > 2 {
		NotFoundResponse(w)
		return
	}
	var id uuid.UUID
	if numParts == 2 {
		id = uuid.FromStringOrNil(parts[1])
		if id == uuid.Nil {
			BadRequestResponse(w, fmt.Errorf("Bad UUID: %s", parts[1]))
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		if len(parts) == 1 {
			entity, err = cr.handler.search(w, r)
			if err == model.ErrNotFound {
				entity = []utils.Sentinel{}
			} else if err != nil {
				ErrorResponse(w, err)
				return
			}
			OkResponse(w, r, entity, defaultLifetime)
			return
		}

		entity, err = cr.handler.retrieve(w, r, id)
		if err != nil {
			ErrorResponse(w, err)
			return
		}
		OkResponse(w, r, entity, defaultLifetime)

	case http.MethodPost:
		if len(parts) > 1 {
			NotAllowedResponse(w, []string{"GET", "PUT", "DELETE"})
			return
		}
		entity = cr.handler.entity()
		err = json.NewDecoder(r.Body).Decode(&entity)
		if err != nil {
			BadRequestResponse(w, err)
			return
		}
		l, err = cr.handler.create(w, r, entity)
		if err != nil {
			ErrorResponse(w, err)
			return
		}
		CreatedResponse(w, r, entity, l)

	case http.MethodPut:
		if len(parts) == 1 {
			NotAllowedResponse(w, []string{"GET", "POST"})
			return
		}
		entity = cr.handler.entity()
		err = json.NewDecoder(r.Body).Decode(&entity)
		if err != nil {
			BadRequestResponse(w, err)
			return
		}
		err = cr.handler.update(w, r, id, entity)
		if err != nil {
			ErrorResponse(w, err)
			return
		}
		NoContentResponse(w)

	case http.MethodDelete:
		if len(parts) == 1 {
			NotAllowedResponse(w, []string{"GET", "POST"})
			return
		}
		err = cr.handler.delete(w, r, id)
		if err != nil {
			ErrorResponse(w, err)
			return
		}
		NoContentResponse(w)

	case http.MethodOptions:
		if len(parts) == 1 {
			OptionsResponse(w, r, []string{"GET", "POST"}, utils.Nothing)
		} else {
			OptionsResponse(w, r, []string{"GET", "PUT", "DELETE"}, utils.Nothing)
		}
	default:
		if len(parts) == 1 {
			NotAllowedResponse(w, []string{"GET", "POST"})
		} else {
			NotAllowedResponse(w, []string{"GET", "PUT", "DELETE"})
		}
	}
}
