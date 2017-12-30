package model

import (
	"errors"

	"github.com/satori/go.uuid"
)

// ErrNotFound is a DB-agnostic object not found error
var ErrNotFound = errors.New("The requested object could not be found")

type AppContextFactory interface {
	Get() (AppContext, error)
	Close() error
}

type AppContext interface {
	SubjectRepo() SubjectRepo
	CheckRepo() CheckRepo
	CheckResultRepo() CheckResultRepo
	CheckStateRepo() CheckStateRepo
	AlertRepo() AlertRepo
	TagRepo() TagRepo
	RoleRepo() RoleRepo
	PeriodRepo() PeriodRepo
	CheckConnection() error
	Close() error
}

type SubjectRepo interface {
	Find(id uuid.UUID) (Subject, error)
	Create(subject *Subject) error
	Update(subject Subject) error
	Delete(subjectID uuid.UUID) error
	Count() (int, error)
	Named(name string) (Subject, error)
	Search(name, role string) ([]Subject, error)
	ByRoles(roles []string) ([]Subject, error)
}

type RoleRepo interface {
	AllRoles() ([]string, error)
	SharedRoles(role string) ([]string, error)
	CountRoles() (int, error)
}

type CheckRepo interface {
	Find(id uuid.UUID) (Check, error)
	Create(check *Check) error
	Update(check Check) error
	Delete(checkID uuid.UUID) error
	Count() (int, error)
	Search(name, role, tag string) ([]Check, error)
	ForRoles(roles []string) ([]Check, error)
	OfTypes(types []CheckType) ([]Check, error)
	OfTypesForRoles(types []CheckType, roles []string) ([]Check, error)
}

type TagRepo interface {
	AllTags() ([]string, error)
	CountTags() (int, error)
}

type CheckResultRepo interface {
	Create(check *CheckResult) error
	Count() (int, error)
	DeleteBySubject(subjectID uuid.UUID) error
	DeleteByCheck(checkID uuid.UUID) error
	DeleteBySubjectCheck(id SubjectCheckID) error
}

type CheckStateRepo interface {
	Find(id SubjectCheckID) (CheckState, error)
	Upsert(check CheckState) error
	DeleteBySubject(subjectID uuid.UUID) error
	DeleteByCheck(checkID uuid.UUID) error
	DeleteBySubjectCheck(id SubjectCheckID) error
	Count() (int, error)
	ForOwner(owner uuid.UUID) ([]CheckState, error)
	ForTypes(types []CheckType) ([]CheckState, error)
	CoordinatorWorkload() (CoordinatorLoad, error)
	InStatusRoles(statuses []CheckStatus, role []string) ([]CheckState, error)
	CountInRolesByStatus(role []string) (StatusSummary, error)
}

type AlertRepo interface {
	Find(id uuid.UUID) (Alert, error)
	Create(alert *Alert) error
	Update(alert Alert) error
	Delete(alertID uuid.UUID) error
	Count() (int, error)
	Search(name, role, tag string) ([]Alert, error)
	FindByFilter(roles, tags []string) ([]Alert, error)
}

type PeriodRepo interface {
	Find(id uuid.UUID) (Period, error)
	Create(alert *Period) error
	Update(alert Period) error
	Delete(alertID uuid.UUID) error
	Count() (int, error)
	Search(name, role, tag string) ([]Period, error)
	FindForSubject(subject Subject, types []PeriodType) ([]Period, error)
	FindForSubjectChecks(subject Subject, tags []string, types []PeriodType) ([]Period, error)
	FindByType(types []PeriodType) ([]Period, error)
}
