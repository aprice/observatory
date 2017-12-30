package checks

import (
	"log"
	"sync"
	"time"

	"github.com/satori/go.uuid"

	"github.com/aprice/observatory/client"
	"github.com/aprice/observatory/collections"
	"github.com/aprice/observatory/model"
	"github.com/aprice/observatory/utils"
)

type Client interface {
	GetObject(path string, headers map[string]string, endpoints []string, payload interface{}) (expires time.Time, err error)
	SendObject(method, path string, endpoints []string, payload interface{}) error
}

type checkConfig struct {
	Check        model.Check
	SubjectID    uuid.UUID
	Coordinators []string
	Active       bool
}

// Execute a check and return its exit code.
func (cc checkConfig) Execute() (model.CheckStatus, error) {
	switch cc.Check.Type {
	case model.CheckExec:
		return executeCheck(cc.SubjectID, cc.Check.Parameters)
	case model.CheckHTTP:
		return httpCheck(cc.SubjectID, cc.Check.Parameters)
	case model.CheckPort:
		return portCheck(cc.SubjectID, cc.Check.Parameters)
	case model.CheckMemory:
		return memCheck(cc.SubjectID, cc.Check.Parameters)
	case model.CheckCPU:
		return loadCheck(cc.SubjectID, cc.Check.Parameters)
	case model.CheckDisk:
		return diskCheck(cc.SubjectID, cc.Check.Parameters)
	default:
		return model.StatusNone, nil
	}
}

type runningCheck struct {
	Control chan checkConfig
}

// Update check with new configuration.
func (r *runningCheck) Update(check model.Check, config model.AgentConfig) {
	newConf := checkConfig{check, config.ID, config.Coordinators, true}
	r.Control <- newConf
}

// Stop check from looping.
func (r *runningCheck) Stop() {
	newConf := checkConfig{Active: false}
	r.Control <- newConf
}

var runningChecks = map[string]runningCheck{}
var rcsLock = sync.Mutex{}

// StopAllChecks from looping.
func StopAllChecks() {
	rcsLock.Lock()
	defer rcsLock.Unlock()
	for _, rc := range runningChecks {
		rc.Stop()
	}
}

// UpdateChecks configurations.
func UpdateChecks(agentConfig model.AgentConfig) {
	allChecks := make(map[string]model.Check, len(agentConfig.Checks))
	seenChecks := make(collections.StringSet, len(agentConfig.Checks))
	stoppedChecks := make([]string, 0, len(runningChecks))

	for _, c := range agentConfig.Checks {
		allChecks[c.ID.String()] = c
	}

	rcsLock.Lock()
	defer rcsLock.Unlock()
	for id, rc := range runningChecks {
		if check, ok := allChecks[id]; ok {
			rc.Update(check, agentConfig)
			seenChecks.Add(id)
		} else {
			rc.Stop()
			stoppedChecks = append(stoppedChecks, id)
		}
	}

	for _, id := range stoppedChecks {
		delete(runningChecks, id)
	}

	for id, c := range allChecks {
		if !seenChecks.Contains(id) {
			addCheck(c, agentConfig)
		}
	}
}

func addCheck(check model.Check, config model.AgentConfig) {
	log.Printf("Adding check %s", check.Name)
	newConf := checkConfig{check, config.ID, config.Coordinators, true}
	newCheck := runningCheck{make(chan checkConfig)}
	runningChecks[check.ID.String()] = newCheck
	go repeatCheck(newConf, newCheck.Control)
}

func repeatCheck(initialConfig checkConfig, control chan checkConfig) {
	config := initialConfig
	log.Printf("Executing check %s every %v", config.Check.Name, config.Check.IntervalDuration())
	doCheck(config)
	repeater := utils.StartRepeater(config.Check.IntervalDuration())
	defer repeater.Stop()
	for {
		if !config.Active {
			return
		}

		select {
		case config = <-control:
			if !config.Active {
				log.Printf("Check %s stopped.", config.Check.Name)
				return
			}
			repeater.UpdateInterval(config.Check.IntervalDuration())
		case <-repeater.C:
			doCheck(config)
		}
	}
}

func doCheck(config checkConfig) {
	log.Printf("Executing check %s\n", config.Check.Name)
	status, err := config.Execute()
	if err != nil {
		log.Println(err)
	} else {
		result := model.NewCheckResult(config.SubjectID, config.Check.ID, time.Now(), status)
		// Record check result non-blocking.
		go func() {
			err = client.SendObject("POST", "/checkresults", config.Coordinators, result)
			if err != nil {
				log.Println(err)
			} else {
				log.Printf("Recorded %s result %d", config.Check.Name, status)
			}
		}()
	}
}
