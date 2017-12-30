package alert

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os/exec"
	"strings"
	"time"

	"gopkg.in/gomail.v2"

	"github.com/aprice/observatory/collections"
	"github.com/aprice/observatory/model"
	"github.com/aprice/observatory/server/config"
	"github.com/aprice/observatory/utils"
)

// MockAlertExecutions records when Mock alerts are executed for testing.
var MockAlertExecutions = collections.StringSet{}

// ExecuteAlerts for a given check.
//TODO: Skip over any alerts that have a tag in a blackout period that covers this CheckResult,
//	even if the check does not have the tag.
func ExecuteAlerts(result model.CheckResultDetail, prevStatus model.CheckStatus, ctx model.AppContext, conf config.Configuration) error {
	now := time.Now()
	state, err := ctx.CheckStateRepo().Find(result.SubjectCheckID)
	if err != nil {
		return err
	}
	if result.Status <= model.StatusOK && prevStatus <= model.StatusOK {
		// Was OK before, still OK now, nothing to do here.
		return err
	}
	alerts, err := ctx.AlertRepo().FindByFilter(result.Subject.Roles, result.Check.Tags)
	if err != nil {
		return err
	}
	for _, alert := range alerts {
		// If we've recovered, or are newly in problem state, reset reminders. If this was a non-issue (e.g. OK -> OK) we would have bailed further up
		if prevStatus == model.StatusOK || result.Status == model.StatusOK {
			state.Reminders = map[string]time.Time{}
		}
		if lastAlert, ok := state.Reminders[alert.ID.String()]; !ok || (alert.ReminderInterval > 0 && now.Sub(lastAlert) >= alert.ReminderDuration()) {
			err = executeAlert(result, alert, conf)
			if err != nil {
				log.Printf("Firing alert %s failed: %v", alert.Name, err)
			} else {
				state.Reminders[alert.ID.String()] = now
				err = ctx.CheckStateRepo().Upsert(state)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}

	return nil
}

func executeAlert(result model.CheckResultDetail, alert model.Alert, conf config.Configuration) error {
	var err error
	switch alert.Type {
	case model.AlertExec:
		err = executeAlertExec(result, alert.Parameters)
	case model.AlertEmail:
		err = executeAlertEmail(result, alert.Parameters, conf)
	case model.AlertMock:
		MockAlertExecutions.Add(fmt.Sprintf("%s/%s", result.SubjectID, result.CheckID))
	default:
		err = fmt.Errorf("Unknown alert type: %d", alert.Type)
	}
	return err
}

func executeAlertExec(result model.CheckResultDetail, params map[string]string) error {
	tpl, err := handleAlertTemplate(params["command"], result)
	if err != nil {
		return err
	}
	args := utils.StringToArgs(tpl)
	log.Printf("Executing %s with %d args: %v", args[0], len(args)-1, args[1:])
	cmd := exec.Command(args[0])
	cmd.Args = args

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func executeAlertEmail(result model.CheckResultDetail, params map[string]string, conf config.Configuration) error {
	var err error
	from := conf.EmailFrom
	to := strings.Split(params["to"], ",")
	subject := params["subject"]
	subject, err = handleAlertTemplate(subject, result)
	if err != nil {
		subject = err.Error()
	}
	body := params["body"]
	body, err = handleAlertTemplate(body, result)
	if err != nil {
		body = err.Error()
	}
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(conf.SMTPHost, conf.SMTPPort, conf.SMTPUser, conf.SMTPPassword)

	log.Printf("Sending %s to %s", subject, strings.Join(to, ","))
	err = d.DialAndSend(m)
	return err
}

func handleAlertTemplate(templateText string, ai model.CheckResultDetail) (string, error) {
	//TODO: Cache for re-use; templates are thread-safe once parsed. Requires
	//ensuring that we expire them when they change or use a short TTL.
	tmpl, err := template.New("alert").Parse(templateText)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	tmpl.Execute(buf, ai)
	return buf.String(), nil
}
