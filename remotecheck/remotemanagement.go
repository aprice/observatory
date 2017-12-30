package remotecheck

import (
	"log"
	"math/rand"
	"time"

	"github.com/satori/go.uuid"

	"github.com/aprice/observatory/collections"
	"github.com/aprice/observatory/model"
	"github.com/aprice/observatory/server/config"
	"github.com/aprice/observatory/utils"
)

type blackoutFilter struct {
	Roles    collections.StringSet
	Tags     collections.StringSet
	Subjects collections.StringSet
}

func newBlackoutFilter(period model.Period) blackoutFilter {
	bf := blackoutFilter{
		Roles:    collections.NewStringSet(period.Roles...),
		Tags:     collections.NewStringSet(period.Tags...),
		Subjects: make(collections.StringSet, len(period.Subjects)),
	}
	for _, id := range period.Subjects {
		bf.Subjects.Add(id.String())
	}
	return bf
}

func (f blackoutFilter) MatchSubjectCheck(subject model.Subject, check model.Check) bool {
	return (len(f.Tags) == 0 || f.Tags.ContainsAny(check.Tags...)) &&
		(len(f.Roles) == 0 || f.Roles.ContainsAny(subject.Roles...)) &&
		(len(f.Subjects) == 0 || f.Subjects.Contains(subject.ID.String()))
}

type blackoutFilterSet []blackoutFilter

func newBlackoutFilterSet(periods []model.Period) blackoutFilterSet {
	bfs := make(blackoutFilterSet, len(periods))
	for _, period := range periods {
		bfs = append(bfs, newBlackoutFilter(period))
	}
	return bfs
}

func (bfs blackoutFilterSet) MatchSubjectCheck(subject model.Subject, check model.Check) bool {
	for _, bf := range bfs {
		if bf.MatchSubjectCheck(subject, check) {
			return true
		}
	}
	return false
}

type remoteCheckDetail struct {
	ID           model.SubjectCheckID
	State        model.CheckState
	Subject      model.Subject
	Check        model.Check
	SubjectRoles collections.StringSet
	CheckTags    collections.StringSet
	Blackout     bool
}

type remoteCheckSet map[model.SubjectCheckID]*remoteCheckDetail

func (rcs remoteCheckSet) AddIfApplicable(subject model.Subject, check model.Check) {
	id := model.SubjectCheckID{SubjectID: subject.ID, CheckID: check.ID}
	rcd := remoteCheckDetail{
		ID:           id,
		Subject:      subject,
		Check:        check,
		SubjectRoles: collections.NewStringSet(subject.Roles...),
		CheckTags:    collections.NewStringSet(check.Tags...),
	}
	log.Println(subject.Roles)
	if !rcd.SubjectRoles.ContainsAny(check.Tags...) {
		// Not applicable
		return
	}
	rcs[id] = &rcd
}

// UpdateCheckAssignments assigns a Coordinator as the owner of every applicable
// combination of remote check and subject. It ensures a CheckState exists for
// each applicable pairing, and assigns them while attempting to balance load
// accross coordinators. Some assigned checks may be reassigned if the current
// owner is known to be down; if the current owner has not executed the check in
// more than twice its interval; or if there is a significant imbalance, for
// example when a new coordinator comes online.
func UpdateCheckAssignments(conf config.Configuration) error {
	var (
		err            error
		checks         []model.Check
		subjects       []model.Subject
		leastLoaded    uuid.UUID
		leastLoadLevel int
		imbalance      float32
	)

	rcs := remoteCheckSet{}

	ctx, err := conf.ContextFactory.Get()
	if err != nil {
		return err
	}
	defer ctx.Close()

	log.Println("Updating remote check assignments.")

	//*** Get Data ***//
	// Get a list of remote checks
	checks, err = ctx.CheckRepo().OfTypes(model.RemoteCheckTypes)
	if err == model.ErrNotFound {
		log.Println("No remote checks, nothing to do.")
		return nil
	}
	if err != nil {
		return err
	}

	// Get the set of distinct roles used in remote checks
	allCheckRoles := make(collections.StringSet, len(checks))
	for _, check := range checks {
		allCheckRoles.Add(check.Roles...)
	}

	// Get the list of subjects that have any of those roles
	subjects, err = ctx.SubjectRepo().ByRoles(allCheckRoles.ToArray())
	if err == model.ErrNotFound {
		log.Println("No subjects of remote checks, nothing to do.")
		return nil
	}
	if err != nil {
		return err
	}

	// Map out what checks apply to which subjects
	for _, subject := range subjects {
		for _, check := range checks {
			rcs.AddIfApplicable(subject, check)
		}
	}

	// Get a list of all CheckStates for remote checks
	allCheckStates, err := ctx.CheckStateRepo().ForTypes(model.RemoteCheckTypes)
	if err != nil && err != model.ErrNotFound {
		return err
	}
	for _, state := range allCheckStates {
		rcs[state.ID].State = state
	}

	//*** Create missing CheckStates ***//
	for id, rcd := range rcs {
		if rcd.State.Type == model.CheckNone {
			state := model.CheckState{
				ID:    id,
				Roles: rcd.Subject.Roles,
				Type:  rcd.Check.Type,
			}
			err = ctx.CheckStateRepo().Upsert(state)
			if err != nil {
				return err
			}
			rcd.State = state
		}
	}

	//*** Prepare For Load Balancing ***//
	// Get # checks owned by each coordinator
	load, err := ctx.CheckStateRepo().CoordinatorWorkload()
	if err != nil && err != model.ErrNotFound {
		return err
	}
	livePeers := conf.Peers.AlivePeerSet()
	livePeers[conf.ID] = conf.Endpoint()
	minLoad, maxLoad := -1, 0
	for id := range livePeers {
		if l, ok := load[id]; !ok {
			load[id] = 0
			minLoad = 0
		} else {
			if l > maxLoad {
				maxLoad = l
			}
			if minLoad == -1 || l < minLoad {
				minLoad = l
			}
		}
	}
	// If there is an imbalance, we'll rebalance some checks at random.
	if maxLoad > 0 {
		imbalance = 1.0 - (float32(minLoad) / float32(maxLoad))
	}
	maxRebalance := float32(maxLoad) / float32(len(rcs))
	if imbalance < 0.05 || (maxLoad-minLoad) <= 1 {
		// If very nearly perfectly balanced, don't randomly reassign any checks.
		imbalance = 0
	} else if imbalance > maxRebalance {
		// No matter how imbalanced, don't randomly reassign too many checks at once.
		imbalance = maxRebalance
	}

	// Determine known down peers
	downPeerSet := collections.UUIDSet{}
	knownPeers := conf.Peers.KnownPeerSet()
	for id := range knownPeers {
		if _, ok := livePeers[id]; !ok {
			downPeerSet[id] = utils.Nothing
			delete(load, id)
		}
	}

	//*** Assign Checks ***//
	toAssign := []model.CheckState{}
	for _, rcd := range rcs {
		if rcd.State.Owner == uuid.Nil {
			toAssign = append(toAssign, rcd.State)
		} else if _, ok := downPeerSet[rcd.State.Owner]; ok {
			toAssign = append(toAssign, rcd.State)
		} else if time.Now().After(rcd.State.Updated.Add(rcd.Check.IntervalDuration() * 2)) {
			toAssign = append(toAssign, rcd.State)
			load[rcd.State.Owner]--
		} else if imbalance > 0 && imbalance > rand.Float32() {
			toAssign = append(toAssign, rcd.State)
			load[rcd.State.Owner]--
		}
	}

	log.Printf("Updating %d remote check assignments.", len(toAssign))
	for _, state := range toAssign {
		leastLoaded = uuid.Nil
		leastLoadLevel = -1
		for id, count := range load {
			if _, ok := livePeers[id]; !ok {
				continue
			}
			if leastLoadLevel == -1 || count < leastLoadLevel {
				leastLoadLevel = count
				leastLoaded = id
			}
		}
		state.Owner = leastLoaded
		load[leastLoaded]++
		ctx.CheckStateRepo().Upsert(state)
	}
	return nil
}

// GetAssignedChecks gathers the list of checks currently assigned to this
// coordinator, excluding any covered by an active blackout period.
func GetAssignedChecks(conf config.Configuration) ([]model.CheckStateDetail, error) {
	ctx, err := conf.ContextFactory.Get()
	if err != nil {
		return []model.CheckStateDetail{}, err
	}
	defer ctx.Close()
	ownChecks, err := ctx.CheckStateRepo().ForOwner(conf.ID)
	if err == model.ErrNotFound {
		return []model.CheckStateDetail{}, nil
	}
	if err != nil && err != model.ErrNotFound {
		return []model.CheckStateDetail{}, err
	}
	periods, err := ctx.PeriodRepo().FindByType([]model.PeriodType{model.PeriodBlackout})
	if err != nil && err != model.ErrNotFound {
		return []model.CheckStateDetail{}, err
	}
	bfs := newBlackoutFilterSet(periods)
	subjects := make(map[uuid.UUID]model.Subject, len(ownChecks)/2)
	checks := make(map[uuid.UUID]model.Check, len(ownChecks)/2)
	result := make([]model.CheckStateDetail, 0, len(ownChecks))
	for _, state := range ownChecks {
		var (
			subject model.Subject
			check   model.Check
			ok      bool
		)
		if subject, ok = subjects[state.ID.SubjectID]; !ok {
			subject, err = ctx.SubjectRepo().Find(state.ID.SubjectID)
			if err != nil {
				return []model.CheckStateDetail{}, err
			}
			subjects[subject.ID] = subject
		}
		if check, ok = checks[state.ID.CheckID]; !ok {
			check, err = ctx.CheckRepo().Find(state.ID.CheckID)
			if err != nil {
				return []model.CheckStateDetail{}, err
			}
			checks[check.ID] = check
		}
		if bfs.MatchSubjectCheck(subject, check) {
			continue
		}
		csd := model.CheckStateDetail{
			CheckState: state,
			Subject:    subject,
			Check:      check,
		}
		result = append(result, csd)
	}
	return result, nil
}
