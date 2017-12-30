package alert

import (
	"reflect"
	"testing"

	"github.com/aprice/observatory/model"
)

//handleAlertTemplate(templateText string, ai model.CheckResultDetail) (string, error)
func TestHandleAlertTemplate(t *testing.T) {
	crd := model.CheckResultDetail{
		CheckResult: model.CheckResult{Status: model.StatusOK},
		Subject:     model.Subject{Name: "Test Subject"},
		Check:       model.Check{Name: "Test Check"},
	}

	var tests = []struct {
		tpl      string
		expected string
		err      error
	}{
		{
			"{{.Status}}: {{.Subject.Name}} - {{.Check.Name}}",
			"OK: Test Subject - Test Check",
			nil,
		},
	}

	for _, tt := range tests {
		actual, actualerr := handleAlertTemplate(tt.tpl, crd)
		if !reflect.DeepEqual(actual, tt.expected) || !reflect.DeepEqual(actualerr, tt.err) {
			t.Errorf("handleAlertTemplate(%v, %v): expected %v,%v; actual %v,%v",
				tt.tpl, crd, tt.expected, tt.err, actual, actualerr)
		}
	}
}
