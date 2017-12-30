package mongo

import (
	"time"

	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/aprice/observatory/model"
	"github.com/aprice/observatory/utils"
)

// InitConnection sets up the MongoDB database context factory.
func InitConnection(host, db, user, pass string) (model.AppContextFactory, error) {
	di := mgo.DialInfo{
		Addrs:    []string{host},
		Database: db,
	}
	if user != "" {
		di.Username = user
		di.Password = pass
	}
	session, err := mgo.DialWithInfo(&di)
	if err != nil {
		return nil, err
	}
	return AppContextFactory{session, db}, nil
}

type AppContextFactory struct {
	session *mgo.Session
	db      string
}

func (f AppContextFactory) Get() (model.AppContext, error) {
	if err := f.session.Ping(); err != nil {
		return nil, err
	}
	sess := f.session.Copy()
	return &AppContext{DB: sess.DB(f.db)}, nil
}

func (f AppContextFactory) Close() error {
	f.session.Close()
	return nil
}

// AppContext serves as a repository factory for the current session.
type AppContext struct {
	DB *mgo.Database

	subjectRepo     *SubjectRepo
	checkRepo       *CheckRepo
	checkResultRepo *CheckResultRepo
	checkStateRepo  *CheckStateRepo
	alertRepo       *AlertRepo
	periodRepo      *PeriodRepo
}

// SubjectRepo returns a pointer to a SubjectRepo in the current context.
func (c *AppContext) SubjectRepo() model.SubjectRepo {
	if c.subjectRepo == nil {
		c.subjectRepo = &SubjectRepo{c.DB.C("Subjects")}
	}
	return c.subjectRepo
}

// CheckRepo returns a pointer to a CheckRepo in the current context.
func (c *AppContext) CheckRepo() model.CheckRepo {
	if c.checkRepo == nil {
		c.checkRepo = &CheckRepo{c.DB.C("Checks")}
	}
	return c.checkRepo
}

// CheckResultRepo returns a pointer to a CheckResultRepo in the current context.
func (c *AppContext) CheckResultRepo() model.CheckResultRepo {
	if c.checkResultRepo == nil {
		c.checkResultRepo = &CheckResultRepo{c.DB.C("CheckResults")}
	}
	return c.checkResultRepo
}

// CheckStateRepo returns a pointer to a CheckStateRepo in the current context.
func (c *AppContext) CheckStateRepo() model.CheckStateRepo {
	if c.checkStateRepo == nil {
		c.checkStateRepo = &CheckStateRepo{c.DB.C("CheckStates")}
	}
	return c.checkStateRepo
}

// AlertRepo returns a pointer to an AlertRepo in the current context.
func (c *AppContext) AlertRepo() model.AlertRepo {
	if c.alertRepo == nil {
		c.alertRepo = &AlertRepo{c.DB.C("Alerts")}
	}
	return c.alertRepo
}

// RoleRepo returns a pointer to a RoleRepo in the current context.
func (c *AppContext) RoleRepo() model.RoleRepo {
	if c.subjectRepo == nil {
		c.subjectRepo = &SubjectRepo{c.DB.C("Subjects")}
	}
	return c.subjectRepo
}

// TagRepo returns a pointer to a TagRepo in the current context.
func (c *AppContext) TagRepo() model.TagRepo {
	if c.checkRepo == nil {
		c.checkRepo = &CheckRepo{c.DB.C("Checks")}
	}
	return c.checkRepo
}

// PeriodRepo returns a pointer to a PeriodRepo in the current context.
func (c *AppContext) PeriodRepo() model.PeriodRepo {
	if c.periodRepo == nil {
		c.periodRepo = &PeriodRepo{c.DB.C("Periods")}
	}
	return c.periodRepo
}

// CheckConnection with the database server.
func (c *AppContext) CheckConnection() error {
	return c.DB.Session.Ping()
}

// Close the DB session associated with this context.
func (c *AppContext) Close() error {
	c.DB.Session.Close()
	return nil
}

// DropDatabase drops the current Mongo database.
func (c *AppContext) DropDatabase() error {
	return c.DB.DropDatabase()
}

// SubjectRepo acts as a repository of Subjects in the database.
type SubjectRepo struct {
	c *mgo.Collection
}

func (r *SubjectRepo) Count() (int, error) {
	return r.c.Count()
}

// Find a Subject by its ID.
func (r *SubjectRepo) Find(id uuid.UUID) (model.Subject, error) {
	var result model.Subject
	err := r.c.FindId(id).One(&result)
	return result, convertError(err)
}

// ByName a Subject retrieve.
func (r *SubjectRepo) Named(name string) (model.Subject, error) {
	var result model.Subject
	err := r.c.Find(bson.M{"name": name}).One(&result)
	return result, convertError(err)
}

// Search the Subjects in the repo by name regular expression and role (combined with AND).
func (r *SubjectRepo) Search(name, role string) ([]model.Subject, error) {
	result := []model.Subject{}
	query := bson.M{}
	if name != "" {
		query["name"] = bson.M{"$regex": name, "$options": "i"}
	}
	if role != "" {
		query["roles"] = role
	}
	err := r.c.Find(query).All(&result)
	return result, convertError(err)
}

// ByRoles looks up all subjects with any of the given roles.
func (r *SubjectRepo) ByRoles(roles []string) ([]model.Subject, error) {
	result := []model.Subject{}
	err := r.c.Find(bson.M{"roles": bson.M{"$in": roles}}).All(&result)
	return result, convertError(err)
}

// AllRoles returns all distinct roles used across all Subjects.
func (r *SubjectRepo) AllRoles() ([]string, error) {
	result := []string{}
	err := r.c.Find(nil).Distinct("roles", &result)
	return result, convertError(err)
}

// SharedRoles returns all distinct roles used across all Subjects which also
// have the given role.
func (r *SubjectRepo) SharedRoles(role string) ([]string, error) {
	result := []string{}
	err := r.c.Find(bson.M{"roles": role}).Distinct("roles", &result)
	return result, convertError(err)
}

func (r *SubjectRepo) CountRoles() (int, error) {
	result := map[string]int{}
	err := r.c.Pipe([]bson.M{
		bson.M{
			"$match": bson.M{
				"roles": bson.M{"$not": bson.M{"$size": 0}},
			},
		},
		bson.M{"$unwind": "$roles"},
		bson.M{
			"$group": bson.M{
				"_id":   "$roles",
				"count": bson.M{"$sum": 1},
			},
		},
		bson.M{
			"$group": bson.M{
				"_id":   "count",
				"count": bson.M{"$sum": 1},
			},
		},
	}).One(result)
	if err != nil {
		return 0, convertError(err)
	}

	return result["count"], nil
}

// Create a new Subject in the repo.
func (r *SubjectRepo) Create(subject *model.Subject) error {
	id := utils.NewTimeUUID()
	subject.ID = id
	_, err := r.c.UpsertId(id, subject)
	if err != nil {
		return convertError(err)
	}

	subject.ID = id
	return nil
}

// Update a Subject in the repo.
func (r *SubjectRepo) Update(subject model.Subject) error {
	_, err := r.c.UpsertId(subject.ID, subject)
	return convertError(err)
}

// Delete a Subject from the repo.
func (r *SubjectRepo) Delete(subjectID uuid.UUID) error {
	err := r.c.Remove(bson.M{"_id": subjectID})
	return convertError(err)
}

// CheckRepo acts as a repository of Checks in the database.
type CheckRepo struct {
	c *mgo.Collection
}

func (r *CheckRepo) Count() (int, error) {
	return r.c.Count()
}

// Find a Check by its ID.
func (r *CheckRepo) Find(id uuid.UUID) (model.Check, error) {
	var result model.Check
	err := r.c.FindId(id).One(&result)
	return result, convertError(err)
}

// Search the Checks in the repo by name regular expression and role/tag
// (combined with AND). Any parameter left blank will be ignored.
func (r *CheckRepo) Search(name, role, tag string) ([]model.Check, error) {
	result := []model.Check{}
	query := bson.M{}
	if name != "" {
		query["name"] = bson.M{"$regex": name, "$options": "i"}
	}
	if role != "" {
		query["roles"] = role
	}
	if tag != "" {
		query["tags"] = tag
	}
	err := r.c.Find(query).All(&result)
	return result, convertError(err)
}

// Create a new Check in the repo.
func (r *CheckRepo) Create(check *model.Check) error {
	id := utils.NewTimeUUID()
	check.ID = id
	_, err := r.c.UpsertId(id, check)
	if err != nil {
		return convertError(err)
	}
	return nil
}

// Update a Check in the repo.
func (r *CheckRepo) Update(check model.Check) error {
	_, err := r.c.UpsertId(check.ID, check)
	return convertError(err)
}

// Delete a Check from the repo.
func (r *CheckRepo) Delete(checkID uuid.UUID) error {
	err := r.c.Remove(bson.M{"_id": checkID})
	return convertError(err)
}

// ForRoles returns all Checks for the given Roles.
func (r *CheckRepo) ForRoles(roles []string) ([]model.Check, error) {
	result := []model.Check{}
	q := r.c.Find(bson.M{"roles": bson.M{"$in": roles}})
	count, err := q.Count()
	if err != nil {
		return []model.Check{}, convertError(err)
	}
	if count == 0 {
		return []model.Check{}, model.ErrNotFound
	}
	err = q.All(&result)
	return result, convertError(err)
}

// OfTypes returns all Checks for the given Types.
func (r *CheckRepo) OfTypes(types []model.CheckType) ([]model.Check, error) {
	result := []model.Check{}
	q := r.c.Find(bson.M{"type": bson.M{"$in": types}})
	count, err := q.Count()
	if err != nil {
		return []model.Check{}, convertError(err)
	}
	if count == 0 {
		return []model.Check{}, model.ErrNotFound
	}
	err = q.All(&result)
	return result, convertError(err)
}

// OfTypesForRoles returns all checks for the given Roles of the given Types.
func (r *CheckRepo) OfTypesForRoles(types []model.CheckType, roles []string) ([]model.Check, error) {
	result := []model.Check{}
	q := r.c.Find(bson.M{"type": bson.M{"$in": types}, "roles": bson.M{"$in": roles}})
	count, err := q.Count()
	if err != nil {
		return []model.Check{}, convertError(err)
	}
	if count == 0 {
		return []model.Check{}, model.ErrNotFound
	}
	err = q.All(&result)
	return result, convertError(err)
}

// AllTags returns all distinct tags used across all Checks.
func (r *CheckRepo) AllTags() ([]string, error) {
	result := []string{}
	err := r.c.Find(nil).Distinct("tags", &result)
	return result, convertError(err)
}

func (r *CheckRepo) CountTags() (int, error) {
	result := map[string]int{}
	err := r.c.Pipe([]bson.M{
		bson.M{
			"$match": bson.M{
				"tags": bson.M{"$not": bson.M{"$size": 0}},
			},
		},
		bson.M{"$unwind": "$tags"},
		bson.M{
			"$group": bson.M{
				"_id":   "$tags",
				"count": bson.M{"$sum": 1},
			},
		},
		bson.M{
			"$group": bson.M{
				"_id":   "count",
				"count": bson.M{"$sum": 1},
			},
		},
	}).One(result)
	if err != nil {
		return 0, convertError(err)
	}

	return result["count"], nil
}

// CheckResultRepo acts as a repository of CheckResults in the database.
type CheckResultRepo struct {
	c *mgo.Collection
}

func (r *CheckResultRepo) Count() (int, error) {
	return r.c.Count()
}

// Create a new CheckResult in the repo.
func (r *CheckResultRepo) Create(check *model.CheckResult) error {
	id := utils.NewTimeUUID()
	check.ID = id
	_, err := r.c.UpsertId(id, check)
	return convertError(err)
}

// DeleteBySubject deletes all check results for a Subject. Used for cleanup
// after deleting a Subject.
func (r *CheckResultRepo) DeleteBySubject(subjectID uuid.UUID) error {
	err := r.c.Remove(bson.M{"subjectId": subjectID})
	return convertError(err)
}

// DeleteByCheck deletes all check results for a Check. Used for cleanup
// after deleting a Check.
func (r *CheckResultRepo) DeleteByCheck(checkID uuid.UUID) error {
	err := r.c.Remove(bson.M{"checkID": checkID})
	return convertError(err)
}

// DeleteBySubjectCheck deletes all check results for a given subject and check.
// Used for cleanup after removing a role from a subject.
func (r *CheckResultRepo) DeleteBySubjectCheck(id model.SubjectCheckID) error {
	err := r.c.Remove(bson.M{"subjectid": id.SubjectID, "checkid": id.CheckID})
	return convertError(err)
}

// CheckStateRepo acts as a repository of CheckStates in the database.
type CheckStateRepo struct {
	c *mgo.Collection
}

func (r *CheckStateRepo) Count() (int, error) {
	return r.c.Count()
}

// Find a CheckState by its ID.
func (r *CheckStateRepo) Find(id model.SubjectCheckID) (model.CheckState, error) {
	var result model.CheckState
	err := r.c.FindId(id).One(&result)
	return result, convertError(err)
}

// ForOwner returns all CheckStates owned by a given coordinator.
func (r *CheckStateRepo) ForOwner(owner uuid.UUID) ([]model.CheckState, error) {
	result := []model.CheckState{}
	err := r.c.Find(bson.M{"owner": owner}).All(&result)
	return result, convertError(err)
}

// ForTypes returns all CheckStates of a given .
func (r *CheckStateRepo) ForTypes(types []model.CheckType) ([]model.CheckState, error) {
	result := []model.CheckState{}
	err := r.c.Find(bson.M{"type": bson.M{"$in": types}}).All(&result)
	return result, convertError(err)
}

type coordinatorLoadRaw struct {
	Coordinator uuid.UUID `bson:"_id"`
	Load        int
}

// CoordinatorWorkload gets the number of checks owned by each coordinator.
func (r *CheckStateRepo) CoordinatorWorkload() (model.CoordinatorLoad, error) {
	result := model.CoordinatorLoad{}
	res := []coordinatorLoadRaw{}
	err := r.c.Pipe([]bson.M{bson.M{"$match": bson.M{"coordinator": bson.M{"$ne": ""}}},
		{"$group": bson.M{"_id": "$owner", "count": bson.M{"$sum": 1}}}}).All(&res)
	if err != nil {
		return result, convertError(err)
	}
	for _, item := range res {
		if item.Coordinator != uuid.Nil {
			result[item.Coordinator] = item.Load
		}
	}
	return result, convertError(err)
}

// InStatus returns all CheckStates for the given statuses and roles.
func (r *CheckStateRepo) InStatusRoles(statuses []model.CheckStatus, roles []string) ([]model.CheckState, error) {
	result := []model.CheckState{}
	query := bson.M{"status": bson.M{"$in": statuses}}
	if len(roles) > 0 {
		query["roles"] = bson.M{"$all": roles}
	}
	q := r.c.Find(query)
	count, err := q.Count()
	if err != nil {
		return result, convertError(err)
	}
	if count == 0 {
		return []model.CheckState{}, model.ErrNotFound
	}
	err = q.All(&result)
	return result, convertError(err)
}

type roleStatus struct {
	Status model.CheckStatus `bson:"_id"`
	Count  int
}

// CountInRolesByStatus returns the count of distinct Subjects for the given role, by status.
func (r *CheckStateRepo) CountInRolesByStatus(roles []string) (model.StatusSummary, error) {
	var out model.StatusSummary
	res := []roleStatus{}
	err := r.c.Pipe([]bson.M{bson.M{"$match": bson.M{"roles": bson.M{"$all": roles}}},
		{"$group": bson.M{"_id": "$_id.subjectid", "maxstatus": bson.M{"$max": "$status"}}},
		{"$group": bson.M{"_id": "$maxstatus", "count": bson.M{"$sum": 1}}}}).All(&res)
	if err != nil {
		return out, convertError(err)
	}
	for _, item := range res {
		switch item.Status {
		case model.StatusOK:
			out.Ok = item.Count
		case model.StatusWarning:
			out.Warning = item.Count
		case model.StatusCritical:
			out.Critical = item.Count
		}
	}
	return out, convertError(err)
}

// Upsert a CheckState in the repo.
func (r *CheckStateRepo) Upsert(check model.CheckState) error {
	_, err := r.c.UpsertId(check.ID, check)
	return convertError(err)
}

// DeleteBySubject deletes all check states for a Subject. Used for cleanup
// after deleting a Subject.
func (r *CheckStateRepo) DeleteBySubject(subjectID uuid.UUID) error {
	err := r.c.Remove(bson.M{"_id.subjectId": subjectID})
	return convertError(err)
}

// DeleteByCheck deletes all check states for a Check. Used for cleanup after
// deleting a Check.
func (r *CheckStateRepo) DeleteByCheck(checkID uuid.UUID) error {
	err := r.c.Remove(bson.M{"_id.checkID": checkID})
	return convertError(err)
}

// DeleteBySubjectCheck deletes the check state for a given subject and check.
// Used for cleanup after removing a role from a subject.
func (r *CheckStateRepo) DeleteBySubjectCheck(id model.SubjectCheckID) error {
	err := r.c.Remove(bson.M{"_id": id})
	return convertError(err)
}

// AlertRepo acts as a repository of Alerts in the database.
type AlertRepo struct {
	c *mgo.Collection
}

func (r *AlertRepo) Count() (int, error) {
	return r.c.Count()
}

// Create a new Alert in the repo.
func (r *AlertRepo) Create(alert *model.Alert) error {
	id := utils.NewTimeUUID()
	alert.ID = id
	_, err := r.c.UpsertId(id, alert)
	if err != nil {
		return convertError(err)
	}

	alert.ID = id
	return nil
}

// Update an Alert in the repo.
func (r *AlertRepo) Update(alert model.Alert) error {
	_, err := r.c.UpsertId(alert.ID, alert)
	return convertError(err)
}

// Delete an Alert from the repo.
func (r *AlertRepo) Delete(alertID uuid.UUID) error {
	err := r.c.Remove(bson.M{"_id": alertID})
	return convertError(err)
}

// Find an Alert by its ID.
func (r *AlertRepo) Find(id uuid.UUID) (model.Alert, error) {
	var result model.Alert
	err := r.c.FindId(id).One(&result)
	return result, convertError(err)
}

// Search the Alerts in the repo by name regular expression and role/tag
// (combined with AND). Any parameter left blank will be ignored.
func (r *AlertRepo) Search(name, role, tag string) ([]model.Alert, error) {
	result := []model.Alert{}
	query := bson.M{}
	if name != "" {
		query["name"] = bson.M{"$regex": name, "$options": "i"}
	}
	if role != "" {
		query["roles"] = role
	}
	if tag != "" {
		query["tags"] = tag
	}
	err := r.c.Find(query).All(&result)
	return result, convertError(err)
}

// FindByFilter returns alerts that match the given roles and tags.
func (r *AlertRepo) FindByFilter(roles, tags []string) ([]model.Alert, error) {
	result := []model.Alert{}
	q := r.c.Find(bson.M{"roles": bson.M{"$in": roles}, "tags": bson.M{"$in": tags}})
	count, err := q.Count()
	if err != nil {
		return result, convertError(err)
	}
	if count == 0 {
		return []model.Alert{}, model.ErrNotFound
	}
	err = q.All(&result)
	return result, convertError(err)
}

// PeriodRepo acts as a repository of Periods in the database.
type PeriodRepo struct {
	c *mgo.Collection
}

func (r *PeriodRepo) Count() (int, error) {
	return r.c.Count()
}

// Create a new Period in the repo.
func (r *PeriodRepo) Create(period *model.Period) error {
	id := utils.NewTimeUUID()
	period.ID = id
	_, err := r.c.UpsertId(id, period)
	if err != nil {
		return convertError(err)
	}
	return nil
}

// Update an Period in the repo.
func (r *PeriodRepo) Update(period model.Period) error {
	_, err := r.c.UpsertId(period.ID, period)
	return convertError(err)
}

// Delete an Period from the repo.
func (r *PeriodRepo) Delete(periodID uuid.UUID) error {
	err := r.c.Remove(bson.M{"_id": periodID})
	return convertError(err)
}

// Find an Period by its ID.
func (r *PeriodRepo) Find(id uuid.UUID) (model.Period, error) {
	var result model.Period
	err := r.c.FindId(id).One(&result)
	return result, convertError(err)
}

// Search the Periods in the repo by name regular expression and role/tag
// (combined with AND). Any parameter left blank will be ignored.
func (r *PeriodRepo) Search(name, role, tag string) ([]model.Period, error) {
	result := []model.Period{}
	query := bson.M{}
	if name != "" {
		query["name"] = bson.M{"$regex": name, "$options": "i"}
	}
	if role != "" {
		query["roles"] = role
	}
	if tag != "" {
		query["tags"] = tag
	}
	err := r.c.Find(query).All(&result)
	return result, convertError(err)
}

// FindForSubject returns all active Periods that affect a subject by roles or ID for the given types (or all types if no types given).
func (r *PeriodRepo) FindForSubject(subject model.Subject, types []model.PeriodType) ([]model.Period, error) {
	result := []model.Period{}
	query := bson.M{
		"$or": []bson.M{
			bson.M{"roles": bson.M{"$in": subject.Roles}},
			bson.M{"subjects": subject.ID},
		},
		"tags":  bson.M{"$eq": [0]string{}},
		"start": bson.M{"$lte": time.Now()},
		"end":   bson.M{"$gte": time.Now()},
	}
	if len(types) > 0 {
		query["type"] = bson.M{"$in": types}
	}
	q := r.c.Find(query)
	count, err := q.Count()
	if err != nil {
		return result, convertError(err)
	}
	if count == 0 {
		return []model.Period{}, model.ErrNotFound
	}
	err = q.All(&result)
	return result, convertError(err)
}

// FindForSubjectChecks returns all active Periods that affect a subject by roles or ID for the given tags and types (or all types if no types given).
func (r *PeriodRepo) FindForSubjectChecks(subject model.Subject, tags []string, types []model.PeriodType) ([]model.Period, error) {
	result := []model.Period{}
	query := bson.M{
		"$or": []bson.M{
			bson.M{"roles": bson.M{"$in": subject.Roles}},
			bson.M{"subjects": subject.ID},
			bson.M{"$and": []bson.M{
				bson.M{"$or": []bson.M{bson.M{"roles": bson.M{"$size": 0}}, bson.M{"roles": nil}}},
				bson.M{"$or": []bson.M{bson.M{"subjects": bson.M{"$size": 0}}, bson.M{"subjects": nil}}},
			}},
		},
		"tags":  bson.M{"$in": tags},
		"start": bson.M{"$lte": time.Now()},
		"end":   bson.M{"$gte": time.Now()},
	}
	if len(types) > 0 {
		query["type"] = bson.M{"$in": types}
	}
	q := r.c.Find(query)
	count, err := q.Count()
	if err != nil {
		return result, convertError(err)
	}
	if count == 0 {
		return []model.Period{}, model.ErrNotFound
	}
	err = q.All(&result)
	return result, convertError(err)
}

// FindByType returns all active periods for the given types (or all types if no types given).
func (r *PeriodRepo) FindByType(types []model.PeriodType) ([]model.Period, error) {
	result := []model.Period{}
	query := bson.M{
		"start": bson.M{"$lte": time.Now()},
		"end":   bson.M{"$gte": time.Now()},
	}
	if len(types) > 0 {
		query["type"] = bson.M{"$in": types}
	}
	q := r.c.Find(query)
	count, err := q.Count()
	if err != nil {
		return result, convertError(err)
	}
	if count == 0 {
		return []model.Period{}, model.ErrNotFound
	}
	err = q.All(&result)
	return result, convertError(err)
}

func convertError(err error) error {
	if err == mgo.ErrNotFound {
		return model.ErrNotFound
	}
	return err
}
