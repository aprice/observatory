package database

import (
	"log"
	"time"

	"github.com/aprice/observatory/model"
)

// FillSampleData creates sample data in the given context.
func FillSampleData(ctx model.AppContext) {
	for _, check := range sampleChecks {
		check.Modified = time.Now()
		err := ctx.CheckRepo().Create(&check)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Created check %s with ID %s", check.Name, check.ID)
		}
	}

	for _, alert := range sampleAlerts {
		alert.Modified = time.Now()
		err := ctx.AlertRepo().Create(&alert)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Created alert %s with ID %s", alert.Name, alert.ID)
		}
	}
}

var sampleChecks = [...]model.Check{
	model.Check{
		Name: "Test OK",
		Type: model.CheckExec,
		Parameters: map[string]string{
			"command": "testcheck --code 0",
		},
		Interval: 30,
		Roles:    []string{"healthy"},
		Tags:     []string{"default"},
	},
	model.Check{
		Name: "Test Warning",
		Type: model.CheckExec,
		Parameters: map[string]string{
			"command": "testcheck --code 1",
		},
		Interval: 30,
		Roles:    []string{"unhealthy"},
		Tags:     []string{"default"},
	},
	model.Check{
		Name: "Test Critical",
		Type: model.CheckExec,
		Parameters: map[string]string{
			"command": "testcheck --code 2",
		},
		Interval: 30,
		Roles:    []string{"dying"},
		Tags:     []string{"default"},
	},
	model.Check{
		Name: "Coordinator Up",
		Type: model.CheckHTTP,
		Parameters: map[string]string{
			"url": "http://localhost:13100/up",
		},
		Interval: 30,
		Roles:    []string{"coordinator"},
		Tags:     []string{"default"},
	},
	model.Check{
		Name: "Agent Checked In",
		Type: model.CheckAgentDown,
		Parameters: map[string]string{
			"warning":  "35s",
			"critical": "65s",
		},
		Interval: 30,
		Roles:    []string{"default"},
		Tags:     []string{"default"},
	},
}

var sampleAlerts = []model.Alert{
	model.Alert{
		Name: "Test",
		Type: model.AlertExec,
		Parameters: map[string]string{
			"command": "testcheck -alert '{{.Check.Name}} status {{.Status}}'",
		},
		Roles: []string{"default"},
		Tags:  []string{"default"},
	},
	model.Alert{
		Name: "Basic Email",
		Type: model.AlertEmail,
		Parameters: map[string]string{
			"to":      "nobody@example.com",
			"subject": "{{.Status}}: {{.Subject.Name}} - {{.Check.Name}}",
			"body":    "Check: {{.Check.Name}}\r\nSubject: {{.Subject.Name}}\r\nResult: {{.Status}}\r\nTime: {{.Time}}",
		},
		Roles: []string{"_disabled"},
		Tags:  []string{"_disabled"},
	},
}
