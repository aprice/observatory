package model

import (
	"fmt"
	"time"

	"github.com/aprice/observatory/utils"

	"github.com/satori/go.uuid"
)

// Subject describes a thing being monitored.
type Subject struct {
	ID          uuid.UUID `bson:"_id,omitempty"`
	Name        string
	Roles       []string
	Modified    time.Time
	LastCheckIn time.Time
}

// GetModified returns the last modified date of the Subject.
func (s Subject) GetModified() time.Time {
	return s.Modified
}

// CoordinatorInfo describes a coordinator node.
type CoordinatorInfo struct {
	ID         uuid.UUID
	Leader     bool
	Version    string
	Build      string
	APIVersion int
}

// AgentConfig describes a specific subject's agent configuration.
type AgentConfig struct {
	ID   uuid.UUID
	Name string

	// Coordinator node addresses the agent may use
	Coordinators []string

	// Checks the agent is to perform
	Checks []Check

	Modified time.Time
}

// NewAgentConfig returns a new AgentConfig object with the given settings.
func NewAgentConfig(subject Subject, checks []Check, periods []Period, coordinators []string) AgentConfig {
	lastMod := subject.Modified
	for _, check := range checks {
		if lastMod.Before(check.Modified) {
			lastMod = check.Modified
		}
	}
	for _, period := range periods {
		effMod := period.EffectiveModified()
		if lastMod.Before(effMod) {
			lastMod = effMod
		}
	}
	return AgentConfig{subject.ID, subject.Name, coordinators, checks, lastMod}
}

// GetModified returns the last modified date of the agent config.
func (ac AgentConfig) GetModified() time.Time {
	return ac.Modified
}

// CheckType is an enumeration of check types
type CheckType int

const (
	CheckNone CheckType = iota
	// CheckExec type is an external executable returning a 0 exit code if OK,
	// 1 if notice, 2 if warning, 3 if critical. Parameter "command" defines
	// command to execute.
	CheckExec
	// CheckHTTP type is an HTTP check.
	CheckHTTP
	// CheckPort type is a simple check if a port is accepting connections.
	CheckPort
	// CheckAgentDown type is an inspection of how long ago an agent checked in.
	CheckAgentDown
	// CheckMemory type is a memory and/or swap threshold check.
	CheckMemory
	// CheckCPU type is a CPU use threshold check.
	CheckCPU
	// CheckDisk type is a disk use threshold check.
	CheckDisk
	// CheckVersion type is a Coordinator version update check.
	CheckVersion
)

func (ct CheckType) String() string {
	switch ct {
	case CheckExec:
		return "Exec"
	case CheckHTTP:
		return "HTTP"
	case CheckPort:
		return "Port"
	case CheckAgentDown:
		return "AgentDown"
	case CheckMemory:
		return "Memory"
	case CheckCPU:
		return "CPU"
	case CheckDisk:
		return "Disk"
	case CheckVersion:
		return "Version"
	default:
		return "None"
	}
}

// LocalCheckTypes lists the Checks that are executed locally by an agent.
var LocalCheckTypes = []CheckType{
	CheckExec,
	CheckHTTP,
	CheckPort,
	CheckMemory,
	CheckCPU,
	CheckDisk,
}

// RemoteCheckTypes lists the Checks that are executed remotely by a coordinator.
var RemoteCheckTypes = []CheckType{
	CheckAgentDown,
	CheckVersion,
}

// Check describes a single health check.
type Check struct {
	ID   uuid.UUID `bson:"_id,omitempty"`
	Name string

	// Type of check
	Type CheckType

	// Check parameters (varies by CheckType)
	Parameters map[string]string

	// Frequency in seconds to perform the check
	Interval int

	// Roles that this check applies to
	Roles []string

	// Tags for this check
	Tags []string

	// Timestamp the check was last modified
	Modified time.Time
}

// GetModified returns the last modified date of the Check.
func (c Check) GetModified() time.Time {
	return c.Modified
}

// IntervalDuration returns the Check's Interval as a time.Duration.
func (c Check) IntervalDuration() time.Duration {
	return time.Duration(c.Interval) * time.Second
}

// CheckStatus describes the basic severity of a check result.
type CheckStatus int

const (
	// StatusNone indicates a check that has not been executed
	StatusNone CheckStatus = iota
	// StatusOK is the result of a check that is passing.
	StatusOK
	// StatusWarning indicates a non-urgent or partial failure state.
	StatusWarning
	// StatusCritical indicates an urgent failure state.
	StatusCritical
	// StatusFailed is the result when a check fails to execute.
	StatusFailed = -1
)

func (cs CheckStatus) String() string {
	switch cs {
	case StatusOK:
		return "OK"
	case StatusWarning:
		return "Warning"
	case StatusCritical:
		return "Critical"
	case StatusFailed:
		return "Failed"
	default:
		return "None"
	}
}

// SubjectCheckID is a composite of subject ID and check ID
type SubjectCheckID struct {
	SubjectID uuid.UUID
	CheckID   uuid.UUID
}

func (scid SubjectCheckID) String() string {
	return fmt.Sprintf("%s/%s", scid.SubjectID.String(), scid.CheckID.String())
}

// CheckResult describes a single result of a single health check on a subject.
type CheckResult struct {
	SubjectCheckID
	ID     uuid.UUID `bson:"_id"`
	Time   time.Time
	Status CheckStatus
}

// NewCheckResult creates and initializes a new CheckResult.
func NewCheckResult(subjectID, checkID uuid.UUID, time time.Time, result CheckStatus) CheckResult {
	return CheckResult{SubjectCheckID{subjectID, checkID}, utils.NewTimeUUID(), time, result}
}

// GetModified returns the creation date of the CheckResult, as they are immutable.
func (cr CheckResult) GetModified() time.Time {
	return cr.Time
}

// CheckResultDetail includes the full details of a CheckResult's Subject and
// Check.
type CheckResultDetail struct {
	CheckResult
	Subject Subject
	Check   Check
}

// GetModified returns the most recent of the CheckResult date, the Subject
// modified date, and the Check modified date.
func (crd CheckResultDetail) GetModified() time.Time {
	return utils.LatestDate(crd.Time, crd.Subject.Modified, crd.Check.Modified)
}

// CheckState describes the status of a single check on a subject.
type CheckState struct {
	ID            SubjectCheckID `bson:"_id"`
	StatusChanged time.Time
	Updated       time.Time
	Status        CheckStatus
	Roles         []string `json:"-"`
	Tags          []string `json:"-"`
	Type          CheckType
	Owner         uuid.UUID            `json:"Owner,omitempty",bson:"omitempty"`
	Reminders     map[string]time.Time `json:"-",bson:"omitempty"`
}

// GetModified returns the last updated date of the CheckState.
func (cs CheckState) GetModified() time.Time {
	return cs.Updated
}

// CheckStateDetail includes the full details of a CheckState's Subject and
// Check.
type CheckStateDetail struct {
	CheckState
	Subject Subject
	Check   Check
}

// GetModified returns the most recent of the CheckState update date, the Subject
// modified date, and the Check modified date.
func (csd CheckStateDetail) GetModified() time.Time {
	return utils.LatestDate(csd.Updated, csd.Subject.Modified, csd.Check.Modified)
}

// StatusSummary describes a breakdown of counts by status.
type StatusSummary struct {
	Ok       int
	Warning  int
	Critical int
}

// CoordinatorLoad maps Coordinators to the number of remote checks assigned to them.
type CoordinatorLoad map[uuid.UUID]int

// AlertType is an enumeration of Alert types.
type AlertType int

const (
	AlertNone AlertType = iota
	// AlertExec type is an external executable. Parameter "command" defines
	// the command to execute.
	AlertExec
	// AlertEmail type is an email alert. Parameters "to", "cc", and "bcc" define
	// recipients, while "subject" and "body" are both templates for creating
	// the message. See https://golang.org/pkg/text/template/.
	AlertEmail
	// AlertPagerDuty type raises incidents using the PagerDuty API. Parameter
	// "service" gives the service endpoint to use, while "subject" and "body"
	// work as in AlertEmail.
	AlertPagerDuty
	// AlertMock is used for integration testing ONLY.
	AlertMock
)

func (at AlertType) String() string {
	switch at {
	case AlertExec:
		return "Exec"
	case AlertEmail:
		return "Email"
	case AlertPagerDuty:
		return "PagerDuty"
	default:
		return "None"
	}
}

// Alert describes an alert to be executed when a check fails.
type Alert struct {
	ID               uuid.UUID `bson:"_id,omitempty"`
	Name             string
	Type             AlertType
	Parameters       map[string]string
	ReminderInterval int
	Roles            []string
	Tags             []string
	Modified         time.Time
}

// ReminderDuration returns the Alert's ReminderInterval as a time.Duration.
func (a Alert) ReminderDuration() time.Duration {
	return time.Duration(a.ReminderInterval) * time.Minute
}

// GetModified returns the last modified date of the Alert.
func (a Alert) GetModified() time.Time {
	return a.Modified
}

// PeriodType is an enumeration of Period types.
type PeriodType int

const (
	// PeriodNone is the default, and has no effect.
	PeriodNone PeriodType = iota
	// PeriodBlackout ceases monitoring entirely.
	PeriodBlackout
	// PeriodQuiet ceases alerting but not monitoring.
	PeriodQuiet
)

// Period encapsulates a window of time where check and/or alert behavior is
// modified, such as a quiet, blackout, redirect, or vigilance period.
type Period struct {
	ID         uuid.UUID `bson:"_id,omitempty"`
	Name       string
	Type       PeriodType
	Modified   time.Time
	Start      time.Time
	End        time.Time
	Parameters map[string]string
	Roles      []string
	Tags       []string
	Subjects   []uuid.UUID
}

// GetModified returns the last modified date of the Period.
func (p Period) GetModified() time.Time {
	return p.Modified
}

// EffectiveModified returns the most recent time of effect for the Period. This
// is the later of the modified date, the start date (if in the past), and the
// end date (if in the past).
func (p Period) EffectiveModified() time.Time {
	ret := p.Modified
	now := time.Now()
	if p.Start.After(now) {
		return ret
	}
	if p.Start.After(ret) {
		ret = p.Start
	}
	if p.End.After(now) {
		return ret
	}
	if p.End.After(ret) {
		return p.End
	}
	return ret
}
