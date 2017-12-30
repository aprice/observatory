package remotecheck

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/aprice/observatory"
	"github.com/aprice/observatory/actions"
	"github.com/aprice/observatory/model"
	"github.com/aprice/observatory/server/config"
	"github.com/aprice/observatory/utils"
	uuid "github.com/satori/go.uuid"
)

func executeCheck(csd model.CheckStateDetail, conf config.Configuration) {
	log.Printf("Executing check %s", csd.Check.Name)
	var (
		status model.CheckStatus
		err    error
	)
	ctx, err := conf.ContextFactory.Get()
	if err != nil {
		log.Println(err)
		return
	}
	defer ctx.Close()
	switch csd.Check.Type {
	case model.CheckAgentDown:
		status, err = executeAgentDownCheck(csd.Subject.ID, csd.Check.Parameters, ctx)
	case model.CheckVersion:
		status, err = executeVersionCheck(csd.Subject.ID, csd.Check.Parameters, ctx)
	default:
		status = model.StatusNone
	}
	if err != nil {
		log.Println(err)
	} else {
		result := model.NewCheckResult(csd.Subject.ID, csd.Check.ID, time.Now(), status)
		actions.RecordCheckResult(result, ctx, conf)
	}
}

func executeAgentDownCheck(subjectID uuid.UUID, params map[string]string, ctx model.AppContext) (model.CheckStatus, error) {
	log.Printf("Checking if agent %v has checked in.", subjectID)
	var status model.CheckStatus
	var err error
	var warnThreshold, critThreshold time.Duration
	if warnThresholdRaw, ok := params["warning"]; ok {
		if warnThreshold, err = time.ParseDuration(warnThresholdRaw); err != nil {
			return model.StatusNone, err
		}
	}
	if critThresholdRaw, ok := params["critical"]; ok {
		if critThreshold, err = time.ParseDuration(critThresholdRaw); err != nil {
			return model.StatusNone, err
		}
	}
	subject, err := ctx.SubjectRepo().Find(subjectID)
	if err != nil {
		return model.StatusFailed, err
	}
	if warnThreshold != time.Duration(0) && subject.LastCheckIn.Add(warnThreshold).Before(time.Now()) {
		status = model.StatusWarning
	} else if critThreshold != time.Duration(0) && subject.LastCheckIn.Add(critThreshold).Before(time.Now()) {
		status = model.StatusCritical
	} else {
		status = model.StatusOK
	}
	return status, nil
}

var checkClient = new(http.Client)

func executeVersionCheck(subjectID uuid.UUID, params map[string]string, ctx model.AppContext) (model.CheckStatus, error) {
	req, err := http.NewRequest("GET", "http://observatory-w3files.s3-website-us-east-1.amazonaws.com/version.json", nil)
	if err != nil {
		return model.StatusFailed, err
	}
	resp, err := checkClient.Do(req)
	if resp != nil {
		defer func() {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}()
	}
	if err != nil || resp.StatusCode >= 400 {
		return model.StatusFailed, err
	}
	vi := map[string]string{}
	err = json.NewDecoder(resp.Body).Decode(&vi)
	if err != nil {
		return model.StatusFailed, err
	}

	latest := ""
	if v, ok := params["type"]; ok {
		if vv, ok := vi[v]; ok {
			latest = vv
		}
	}
	if latest == "" {
		latest = vi["stable"]
	}
	comp, err := utils.CompareSemVer(observatory.Version, latest)
	if err != nil {
		return model.StatusFailed, err
	} else if comp < 0 {
		log.Printf("Found update: currently %s, latest %s", observatory.Version, latest)
		return model.StatusWarning, nil
	} else {
		return model.StatusOK, nil
	}
}
