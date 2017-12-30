package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/aprice/observatory/actions"
	"github.com/aprice/observatory/model"
	"github.com/aprice/observatory/server/config"
	"github.com/aprice/observatory/utils"
)

var defaultLifetime = time.Duration(5) * time.Second
var shortLifetime = time.Duration(1) * time.Second
var noLifetime = time.Duration(0)

// Simplest
func handleUp(w http.ResponseWriter, r *http.Request, conf config.Configuration) {
	if countPathParts(r) > 1 {
		NotFoundResponse(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		OkResponse(w, r, true, conf.PeerCheckDuration())
	case http.MethodOptions:
		OptionsResponse(w, r, []string{"GET"}, utils.Nothing)
	default:
		NotAllowedResponse(w, []string{"GET"})
	}
}

// With payload & subdirs
func handleInfo(w http.ResponseWriter, r *http.Request, conf config.Configuration) {
	if countPathParts(r) > 2 {
		NotFoundResponse(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		sub := pathPart(r, 1)
		if sub == "" {
			payload := actions.Info(conf)
			OkResponse(w, r, payload, time.Duration(24)*time.Hour)
			return
		}

		if sub == "datastats" {
			payload, err := actions.DataStats(conf)
			if err != nil {
				ErrorResponse(w, err)
				return
			}

			OkResponse(w, r, payload, defaultLifetime)
			return
		}

		NotFoundResponse(w)
	case http.MethodOptions:
		OptionsResponse(w, r, []string{"GET"}, utils.Nothing)
	default:
		NotAllowedResponse(w, []string{"GET"})
	}
}

// With path parameters
func handleConfiguration(w http.ResponseWriter, r *http.Request, conf config.Configuration) {
	if countPathParts(r) > 2 {
		NotFoundResponse(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		name := pathPart(r, 1)
		roles := []string{}
		if r.URL.Query().Get("roles") != "" {
			roles = strings.Split(r.URL.Query().Get("roles"), ",")
		}
		payload, err := actions.ConfigureAgent(conf, name, roles)
		if err != nil {
			ErrorResponse(w, err)
			return
		}
		OkResponse(w, r, payload, conf.AgentUpdateDuration())
	case http.MethodOptions:
		OptionsResponse(w, r, []string{"GET"}, utils.Nothing)
	default:
		NotAllowedResponse(w, []string{"GET"})
	}
}

func handlePeers(w http.ResponseWriter, r *http.Request, conf config.Configuration) {
	if countPathParts(r) > 1 {
		NotFoundResponse(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		iam := uuid.FromStringOrNil(r.URL.Query().Get("iam"))
		ep := r.URL.Query().Get("endpoint")
		actions.AddPeer(conf, iam, ep)
		payload := actions.GetPeers(conf)
		OkResponse(w, r, payload, conf.PeerUpdateDuration())
	case http.MethodOptions:
		OptionsResponse(w, r, []string{"GET"}, utils.Nothing)
	default:
		NotAllowedResponse(w, []string{"GET"})
	}
}

func handleCheckResults(w http.ResponseWriter, r *http.Request, conf config.Configuration) {
	if countPathParts(r) > 1 {
		NotFoundResponse(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		NotImplementedResponse(w)
	case http.MethodPost:
		result := model.CheckResult{}
		err := json.NewDecoder(r.Body).Decode(&result)
		if err != nil {
			BadRequestResponse(w, err)
			return
		}
		ctx, err := conf.ContextFactory.Get()
		if err != nil {
			ErrorResponse(w, err)
			return
		}
		defer ctx.Close()
		err = actions.RecordCheckResult(result, ctx, conf)
		if err != nil {
			ErrorResponse(w, err)
			return
		}
		NoContentResponse(w)
	case http.MethodOptions:
		OptionsResponse(w, r, []string{"GET", "POST"}, utils.Nothing)
	default:
		NotAllowedResponse(w, []string{"GET", "POST"})
	}
}

func handleCheckStates(w http.ResponseWriter, r *http.Request, conf config.Configuration) {
	if countPathParts(r) > 1 {
		NotFoundResponse(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		var (
			statuses    []model.CheckStatus
			roles       []string
			rawStatuses []string
			ok          bool
			err         error
			ctx         model.AppContext
		)
		ctx, err = conf.ContextFactory.Get()
		if err != nil {
			log.Println(err)
		}
		defer ctx.Close()

		if rawStatuses, ok = r.URL.Query()["status"]; ok {
			statuses = make([]model.CheckStatus, len(rawStatuses))
			for i, rawStatus := range rawStatuses {
				//TODO: Also support parsing text representations of statuses
				var status int
				status, err = strconv.Atoi(rawStatus)
				if err != nil {
					ErrorResponse(w, err)
					return
				}
				// TODO: Sanity check
				statuses[i] = model.CheckStatus(status)
			}
		} else {
			statuses = []model.CheckStatus{model.StatusOK, model.StatusWarning, model.StatusCritical}
		}

		if roles, ok = r.URL.Query()["role"]; !ok {
			roles = []string{}
		}

		results, err := ctx.CheckStateRepo().InStatusRoles(statuses, roles)
		if err != nil {
			ErrorResponse(w, err)
			return
		}
		sorted := model.CheckStateByPriority(results)
		sort.Sort(sort.Reverse(sorted))

		if _, ok := r.URL.Query()["detail"]; ok {
			details, err := actions.FillCheckStateDetails(ctx, sorted)
			if err != nil {
				ErrorResponse(w, err)
				return
			}
			OkResponse(w, r, details, shortLifetime)
			return
		}
		OkResponse(w, r, sorted, shortLifetime)
	case http.MethodOptions:
		OptionsResponse(w, r, []string{"GET"}, utils.Nothing)
	default:
		NotAllowedResponse(w, []string{"GET"})
	}
}

func handleRoles(w http.ResponseWriter, r *http.Request, conf config.Configuration) {
	if countPathParts(r) > 3 {
		NotFoundResponse(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		ctx, err := conf.ContextFactory.Get()
		if err != nil {
			log.Println(err)
		}
		defer ctx.Close()

		sharedRole := r.URL.Query().Get("sharedRole")
		sub := pathPart(r, 1)
		if sub == "" {
			var roles []string
			if sharedRole == "" {
				roles, err = ctx.RoleRepo().AllRoles()
			} else {
				roles, err = ctx.RoleRepo().SharedRoles(sharedRole)
			}
			if err != nil && err != model.ErrNotFound {
				ErrorResponse(w, err)
				return
			}
			sort.Strings(roles)
			OkResponse(w, r, roles, shortLifetime)
			return
		}

		roles := []string{sub}
		if sharedRole != "" {
			roles = append(roles, sharedRole)
		}
		payload, err := ctx.CheckStateRepo().CountInRolesByStatus(roles)
		if err != nil && err != model.ErrNotFound {
			ErrorResponse(w, err)
			return
		}
		OkResponse(w, r, payload, defaultLifetime)
	case http.MethodOptions:
		OptionsResponse(w, r, []string{"GET"}, utils.Nothing)
	default:
		NotAllowedResponse(w, []string{"GET"})
	}
}

func handleTags(w http.ResponseWriter, r *http.Request, conf config.Configuration) {
	if countPathParts(r) > 1 {
		NotFoundResponse(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		ctx, err := conf.ContextFactory.Get()
		if err != nil {
			log.Println(err)
		}
		defer ctx.Close()
		tags, err := ctx.TagRepo().AllTags()
		if err == model.ErrNotFound {
			tags = []string{}
		} else if err != nil {
			ErrorResponse(w, err)
			return
		}
		sort.Strings(tags)
		OkResponse(w, r, tags, defaultLifetime)
	case http.MethodOptions:
		OptionsResponse(w, r, []string{"GET"}, utils.Nothing)
	default:
		NotAllowedResponse(w, []string{"GET"})
	}
}
