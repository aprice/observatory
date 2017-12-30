package actions

import (
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/aprice/observatory/alert"
	"github.com/aprice/observatory/collections"
	"github.com/aprice/observatory/model"
	"github.com/aprice/observatory/server/config"
	"github.com/aprice/observatory/utils"
	uuid "github.com/satori/go.uuid"
)

// UpdatedCheckCleanup handles cleaning up check states and results when a check
// is modified to no longer apply to some subjects.
func UpdatedCheckCleanup(conf config.Configuration, checkID uuid.UUID, oldRoles, newRoles []string) {
	if reflect.DeepEqual(oldRoles, newRoles) {
		return
	}
	ictx, err := conf.ContextFactory.Get()
	if err != nil {
		log.Println(err)
	}
	defer ictx.Close()
	subjectsBefore, err := ictx.SubjectRepo().ByRoles(oldRoles)
	if err != nil {
		log.Println(err)
		return
	}
	subjectsAfter, err := ictx.SubjectRepo().ByRoles(newRoles)
	if err != nil {
		log.Println(err)
		return
	}
	afterSet := make(collections.UUIDSet, len(subjectsAfter))
	for _, subject := range subjectsAfter {
		afterSet[subject.ID] = utils.Nothing
	}
	wg := new(sync.WaitGroup)
	for _, subject := range subjectsBefore {
		if _, ok := afterSet[subject.ID]; !ok {
			wg.Add(1)
			go func(id model.SubjectCheckID) {
				ictx.CheckResultRepo().DeleteBySubjectCheck(id)
				ictx.CheckStateRepo().DeleteBySubjectCheck(id)
				wg.Done()
			}(model.SubjectCheckID{SubjectID: subject.ID, CheckID: checkID})
		}
	}
	wg.Wait()
}

// DeletedCheckCleanup handles removing states & results for a check that's been
// deleted.
func DeletedCheckCleanup(conf config.Configuration, id uuid.UUID) {
	ictx, err := conf.ContextFactory.Get()
	if err != nil {
		log.Println(err)
	}
	defer ictx.Close()

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		err := ictx.CheckResultRepo().DeleteByCheck(id)
		if err != nil {
			log.Println(err)
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		err := ictx.CheckStateRepo().DeleteByCheck(id)
		if err != nil {
			log.Println(err)
		}
		wg.Done()
	}()
	wg.Wait()
}

// UpdatedSubjectCleanup handles cleaning up check states and results when a
// subject is modified to no longer apply to some checks.
func UpdatedSubjectCleanup(conf config.Configuration, subjectID uuid.UUID, oldRoles, newRoles []string) {
	if reflect.DeepEqual(oldRoles, newRoles) {
		return
	}
	ictx, err := conf.ContextFactory.Get()
	if err != nil {
		log.Println(err)
	}
	defer ictx.Close()

	var (
		checksBefore []model.Check
		checksAfter  []model.Check
	)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var err error
		checksBefore, err = ictx.CheckRepo().ForRoles(oldRoles)
		if err != nil {
			log.Println(err)
			return
		}
	}()
	wg.Add(1)
	go func() {
		var err error
		checksAfter, err = ictx.CheckRepo().ForRoles(newRoles)
		if err != nil {
			log.Println(err)
			return
		}
	}()
	wg.Wait()

	afterSet := make(collections.UUIDSet, len(checksAfter))
	for _, check := range checksAfter {
		afterSet[check.ID] = utils.Nothing
	}
	for _, check := range checksBefore {
		if _, ok := afterSet[check.ID]; !ok {
			wg.Add(1)
			go func(id model.SubjectCheckID) {
				ictx.CheckResultRepo().DeleteBySubjectCheck(id)
				ictx.CheckStateRepo().DeleteBySubjectCheck(id)
				wg.Done()
			}(model.SubjectCheckID{SubjectID: subjectID, CheckID: check.ID})
		}
	}
	wg.Wait()
}

// DeletedSubjectCleanup handles removing states & results for a subject that's
// been deleted.
func DeletedSubjectCleanup(conf config.Configuration, id uuid.UUID) {
	ictx, err := conf.ContextFactory.Get()
	if err != nil {
		log.Println(err)
	}
	defer ictx.Close()

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		err := ictx.CheckResultRepo().DeleteBySubject(id)
		if err != nil {
			log.Println(err)
		}
	}()
	wg.Add(1)
	go func() {
		err := ictx.CheckStateRepo().DeleteBySubject(id)
		if err != nil {
			log.Println(err)
		}
		wg.Done()
	}()
	wg.Wait()
}

// RecordCheckResult including updating check state.
func RecordCheckResult(result model.CheckResult, ctx model.AppContext, conf config.Configuration) error {
	var (
		subject model.Subject
		check   model.Check
		err     error
		e1      error
		e2      error
	)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		subject, e1 = ctx.SubjectRepo().Find(result.SubjectID)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		check, e2 = ctx.CheckRepo().Find(result.CheckID)
		wg.Done()
	}()
	wg.Wait()

	if e1 != nil {
		return e1
	}
	if e2 != nil {
		return e2
	}

	subjectRoles := collections.NewStringSet(subject.Roles...)
	if !subjectRoles.ContainsAny(check.Roles...) {
		log.Printf("Not saving check result %s (roles %v) because check is not applicable to subject %s (roles %v)",
			check.Name, check.Roles, subject.Name, subject.Roles)
		return nil
	}
	periods, err := ctx.PeriodRepo().FindForSubjectChecks(subject, check.Tags, []model.PeriodType{model.PeriodBlackout, model.PeriodQuiet})
	var blackout, quiet bool
	for _, period := range periods {
		switch period.Type {
		case model.PeriodBlackout:
			blackout = true
		case model.PeriodQuiet:
			quiet = true
		}
	}
	if blackout {
		return nil
	}
	prevStatus := model.StatusOK
	state, err := ctx.CheckStateRepo().Find(result.SubjectCheckID)
	if err == model.ErrNotFound {
		state = model.CheckState{
			ID:            result.SubjectCheckID,
			StatusChanged: result.Time,
			Updated:       result.Time,
			Status:        result.Status,
			Type:          check.Type,
			Roles:         subject.Roles,
			Tags:          check.Tags,
		}
	} else if err != nil {
		log.Printf("Loading CheckState failed: %s", err.Error())
	} else {
		prevStatus = state.Status
		state.Updated = result.Time
		state.Roles = subject.Roles
		state.Tags = check.Tags
		state.Type = check.Type
		if state.Status != result.Status {
			state.StatusChanged = result.Time
			state.Status = result.Status
		}
		if result.Status == model.StatusOK {
			state.Reminders = map[string]time.Time{}
		}
	}

	wg.Add(1)
	go func() {
		if e1 = ctx.CheckResultRepo().Create(&result); e1 != nil {
			log.Printf("Saving CheckResult failed: %s.", e1.Error())
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		if e2 = ctx.CheckStateRepo().Upsert(state); e2 != nil {
			log.Printf("Saving CheckState failed: %s.", e2.Error())
		}
		wg.Done()
	}()
	wg.Wait()
	crd := model.CheckResultDetail{
		CheckResult: result,
		Subject:     subject,
		Check:       check,
	}
	if !quiet {
		err = alert.ExecuteAlerts(crd, prevStatus, ctx, conf)
		if err != nil {
			log.Printf("Executing alerts failed: %s.", err.Error())
		}
	}
	return nil
}

// FillCheckStateDetails transforms a collection of CheckStates into
// CheckStateDetails by looking up the check & subject for each.
func FillCheckStateDetails(ctx model.AppContext, states []model.CheckState) ([]model.CheckStateDetail, error) {
	csds := make([]model.CheckStateDetail, len(states))
	errs := make([]error, len(states))
	subjects := make(map[uuid.UUID]model.Subject, len(states)/2)
	checks := make(map[uuid.UUID]model.Check, len(states)/2)
	wg := new(sync.WaitGroup)
	for i, state := range states {
		wg.Add(1)
		go func(idx int, state model.CheckState) {
			var (
				subject model.Subject
				check   model.Check
				err     error
				ok      bool
			)
			if subject, ok = subjects[state.ID.SubjectID]; !ok {
				subject, err = ctx.SubjectRepo().Find(state.ID.SubjectID)
				if err != nil {
					errs[idx] = err
					return
				}
			}
			if check, ok = checks[state.ID.CheckID]; !ok {
				check, err = ctx.CheckRepo().Find(state.ID.CheckID)
				if err != nil {
					errs[idx] = err
					return
				}
			}
			csds[idx] = model.CheckStateDetail{CheckState: state, Subject: subject, Check: check}
			wg.Done()
		}(i, state)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return []model.CheckStateDetail{}, err
		}
	}
	return csds, nil
}
