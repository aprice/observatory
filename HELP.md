# Observatory
By Adrian Price

## Components

### Coordinator
The Coordinator is a service intended to run on dedicated instances, with
multiple instances for redundancy and capacity. It serves a REST API that acts
as the interface between clients (including the web UI) and the data store. It
also performs any remote checks that have been configured, automatically load
balancing remote checks across all active Coordinator nodes. When a Coordinator
receives a check, it also handles any Alerts that need to fire for that check.

### Agent
The Agent is a lightweight service that runs on monitored instances, performing
checks and sending the results back to the Coordinator for processing. It can be
provided a single Coordinator to connect to, and will populate from that node a
full list of all available Coordinators, randomly balancing requests between them.

### Web UI
The Web UI is a set of static files which utilize the Coordinator REST API to
interact with the system.

### Data Store
The only data store currently supported is MongoDB, though support for other
backing stores is planned.

## Concepts
- Subject: a distinct entity under observation.
- Check: a distinct facet for observation.
- Alert: an action to be taken when a check returns an unhealthy result.
- Period: a window of time during which monitoring is modified for some set of
subjects and checks.
- Role: a classification for subjects.
- Tag: a classification for checks.
- Agent: a service that can be run on a subject to execute checks.
- Coordinator: a service that exposes a rest API for interacting with the system,
and executes remote checks.

### Subjects

A subject is any entity being monitored by observatory. A subject has a unique
name and a set of roles. The roles determine what checks, alerts, and periods
apply to the subject. When the agent is first run on a host, it automatically
registers a new subject.

### Checks

A check is anything being monitored on some set of subjects. A check has a name,
some parameters (varying by type of check), a set of roles which determine what
subjects the check applies to, and a set of tags. One common parameter among all
check types is interval, which defines how often the check s should run.

### Alerts

An alert is an action executed when a check fails. An alert has a name, some
parameters (varying by type of alert), and a set of roles and tags that determine
when the alert should fire: when a check fails on a subject, any alerts matching
the subject's roles and the check's tags will be executed.

Alert parameters allow for the use of templates to dynamically populate alerts
with relevant information from the failing check. See the template section below.

#### Alert Templates
Templates allow the generation of alerts with pertinent information dynamically
included in the alert. The template syntax is documented specifically here, and
the following data is available to a template:

- `Time` - the timestamp of the check execution
- `Status` - the status of the check: OK, Warning, or Critical
- `Subject`
  - `ID` - the ID of the subject
  - `Name` - the name of the subject
  - `Roles` - an array of roles assigned to the subject
- `Check`
  - `ID` - the ID of the check
  - `Name` - the name of the check
  - `Interval` - the frequency of check execution
  - `Roles` - the array of roles the check applies to
  - `Tags` - the array of tags assigned to the check

For example: `{{.Status}}: {{.Subject.Name}} - {{.Check.Name}}`

### Periods

A period is a window of time during which some check's behavior is modified in
some way, depending on the type of period. Like alerts, periods have a set of
roles and tags determining where they apply.

Blackout periods stop the execution of checks during the period. Quiet periods
do not impact check execution, but prevent the execution of alerts. During a
quiet period, failed checks will still be exposed in the UI. Other period types
are planned.
