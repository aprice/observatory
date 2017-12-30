package remotecheck

import (
	"log"
	"sync"
	"time"

	"github.com/aprice/observatory/collections"
	"github.com/aprice/observatory/model"
	"github.com/aprice/observatory/server/config"
	"github.com/aprice/observatory/utils"
)

const remoteCheckUpdateInterval = time.Duration(30) * time.Second
const remoteCheckAssignInterval = time.Duration(5) * time.Minute

// UpdateRemoteChecks starts a loop continuously updating remote check assignments.
func UpdateRemoteChecks(conf config.Configuration, quit utils.SentinelChannel) {
	if err := UpdateCheckAssignments(conf); err != nil {
		log.Printf("Initial remote check update failed: %s", err.Error())
	}
	go func() {
		if checks, err := GetAssignedChecks(conf); err != nil {
			log.Printf("Updating remote check assignments failed: %s", err.Error())
		} else {
			updateChecks(checks, conf)
		}
	}()
	updateTicker := time.NewTicker(remoteCheckUpdateInterval)
	defer updateTicker.Stop()
	assignTicker := time.NewTicker(remoteCheckAssignInterval)
	defer assignTicker.Stop()
	for {
		select {
		case <-quit:
			stopAllChecks()
			return
		case <-updateTicker.C:
			go func() {
				if checks, err := GetAssignedChecks(conf); err != nil {
					log.Printf("Getting remote check assignments failed: %s", err.Error())
				} else {
					updateChecks(checks, conf)
				}
				log.Println("Finished getting remoting check assignments")
			}()
		case <-assignTicker.C:
			go func() {
				// Follow the leader
				if !conf.IsLeader() {
					log.Printf("Not leader, skipping remote check balancing.")
					return
				}
				log.Printf("%s is leader for remote check balancing.", conf.ID.String())
				if err := UpdateCheckAssignments(conf); err != nil {
					log.Printf("Updating remote check assignments failed: %s", err.Error())
				}
				log.Println("Finished updating remoting check assignments")
			}()
		}
	}
}

// runningCheck describes a currently looping check.
type runningCheck struct {
	Control chan model.CheckStateDetail
}

// Update check with new configuration.
func (r *runningCheck) update(check model.CheckStateDetail) {
	log.Printf("Updating running check config %s", check.ID)
	r.Control <- check
}

// Stop check from looping.
func (r *runningCheck) stop() {
	r.Control <- model.CheckStateDetail{}
}

var noCheckID = model.SubjectCheckID{}.String()
var runningChecks = map[string]runningCheck{}
var rcsLock = sync.Mutex{}

// stopAllChecks from looping.
func stopAllChecks() {
	rcsLock.Lock()
	defer rcsLock.Unlock()
	for _, rc := range runningChecks {
		rc.stop()
	}
}

// updateChecks configurations.
func updateChecks(ownedChecks []model.CheckStateDetail, conf config.Configuration) {
	allChecks := make(map[string]model.CheckStateDetail, len(ownedChecks))
	seenChecks := make(collections.StringSet, len(ownedChecks))
	for _, csd := range ownedChecks {
		allChecks[csd.ID.String()] = csd
	}

	rcsLock.Lock()
	defer rcsLock.Unlock()
	for id, rc := range runningChecks {
		if check, ok := allChecks[id]; ok {
			rc.update(check)
			seenChecks.Add(id)
		} else {
			rc.stop()
		}
	}

	for id, c := range allChecks {
		if !seenChecks.Contains(id) {
			addCheck(c, conf)
		}
	}
}

// addCheck and start it looping. NOT thread-safe!
func addCheck(check model.CheckStateDetail, conf config.Configuration) {
	log.Println("Adding new running check.")
	if oldCheck, ok := runningChecks[check.ID.String()]; ok {
		log.Printf("Attempt to add existing check %s", check.ID)
		oldCheck.update(check)
		return
	}
	newCheck := runningCheck{make(chan model.CheckStateDetail)}
	runningChecks[check.ID.String()] = newCheck
	go repeatCheck(check, conf, newCheck.Control)
}

// repeatCheck every check.Interval seconds.
func repeatCheck(initialConfig model.CheckStateDetail, conf config.Configuration, control chan model.CheckStateDetail) {
	csd := initialConfig
	log.Printf("Executing check %s every %v", csd.Check.Name, csd.Check.IntervalDuration())
	executeCheck(csd, conf)
	repeater := utils.StartRepeater(csd.Check.IntervalDuration())
	defer repeater.Stop()
	for {
		select {
		case csd = <-control:
			if csd.ID.String() == noCheckID {
				return
			}
			repeater.UpdateInterval(csd.Check.IntervalDuration())
		case <-repeater.C:
			executeCheck(csd, conf)
		}
	}
}
