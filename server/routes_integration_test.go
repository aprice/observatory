//+build mongo

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aprice/observatory/alert"
	"github.com/aprice/observatory/database"
	"github.com/aprice/observatory/database/mongo"
	"github.com/aprice/observatory/model"
	"github.com/aprice/observatory/server/config"

	uuid "github.com/satori/go.uuid"
)

/*** TESTS ***/

// GET /up
func TestUp(t *testing.T) {
	execRouteTests(t, []testCase{
		testCase{
			Method:   "GET",
			Route:    "/up",
			ReqBody:  "",
			Status:   200,
			RespBody: "true",
		},
	})
}

func BenchmarkUp(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testRoute("GET", "/up", "")
		}
	})
}

// GET /configuration/:name?roles=x,y,z
func TestConfigure(t *testing.T) {
	execRouteTests(t, []testCase{
		// Bootstrap new subject
		testCase{
			Name:      "bootstrap",
			Method:    "GET",
			Route:     "/configuration/bootstrapper?roles=bootstrap,healthy",
			ReqBody:   "",
			Status:    200,
			RespRegex: `"Name":"bootstrapper","Coordinators":\["127.0.0.1:13100"],"Checks":\[\{[^}]*"Name":"Test OK"[^}]*}[^}]*},\{[^}]*"Name":"Quiet tag check"[^}]*}[^}]*}],`,
		},
		// Reconfigure existing subject
		testCase{
			Name:      "reconfigure",
			Method:    "GET",
			Route:     "/configuration/bootstrapper",
			ReqBody:   "",
			Status:    200,
			RespRegex: `"Name":"bootstrapper","Coordinators":\["127.0.0.1:13100"],"Checks":\[\{[^}]*"Name":"Test OK"[^}]*}[^}]*},\{[^}]*"Name":"Quiet tag check"[^}]*}[^}]*}],`,
		},
		// Reconfigure existing subject with blackout period
		testCase{
			Name:      "reconfigure-blackout",
			Method:    "GET",
			Route:     "/configuration/blackout?roles=blackout,default",
			ReqBody:   "",
			Status:    200,
			RespRegex: `"Name":"blackout","Coordinators":\["127.0.0.1:13100"],"Checks":\[],`,
		},
	})
}

func BenchmarkConfigure(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testRoute("GET", "/configuration/bootstrapper?roles=bootstrap,healthy", "")
		}
	})
}

// POST /checkresults
func TestPostResult(t *testing.T) {
	ctx, err := conf.ContextFactory.Get()
	if err != nil {
		t.Error(err)
	}
	defer ctx.Close()
	brapper, _ := ctx.SubjectRepo().Named("bootstrapper")
	checkTmp, _ := ctx.CheckRepo().Search("Test OK", "", "")
	testOK := checkTmp[0]
	execRouteTests(t, []testCase{
		// Normal
		testCase{
			Name:   "ok",
			Method: "POST",
			Route:  "/checkresults",
			ReqBody: fmt.Sprintf(`{
				"SubjectID": "%s",
				"CheckID": "%s",
				"Time": "%s",
				"Status": 1
			}`, brapper.ID, testOK.ID, time.Now().Format(time.RFC3339)),
			Status: 200,
		},
		// Trigger mock alert
		testCase{
			Name:   "withAlert",
			Method: "POST",
			Route:  "/checkresults",
			ReqBody: fmt.Sprintf(`{
				"SubjectID": "%s",
				"CheckID": "%s",
				"Time": "%s",
				"Status": 3
			}`, brapper.ID, testOK.ID, time.Now().Format(time.RFC3339)),
			Status: 200,
		},
		// Blackout period on check
		testCase{
			Name:   "blackout",
			Method: "POST",
			Route:  "/checkresults",
			ReqBody: fmt.Sprintf(`{
				"SubjectID": "%s",
				"CheckID": "%s",
				"Time": "%s",
				"Status": 3
			}`, brapper.ID, blackoutTagCheckID, time.Now().Format(time.RFC3339)),
			Status: 200,
		},
		// Quiet period on check
		testCase{
			Name:   "quiet",
			Method: "POST",
			Route:  "/checkresults",
			ReqBody: fmt.Sprintf(`{
				"SubjectID": "%s",
				"CheckID": "%s",
				"Time": "%s",
				"Status": 3
			}`, brapper.ID, quietTagCheckID, time.Now().Format(time.RFC3339)),
			Status: 200,
		},
	})

	if !alert.MockAlertExecutions.Contains(fmt.Sprintf("%s/%s", brapper.ID, testOK.ID)) {
		t.Error("Mock alert not executed")
	}
	if alert.MockAlertExecutions.Contains(fmt.Sprintf("%s/%s", brapper.ID, blackoutTagCheckID)) {
		t.Error("Alert executed under blackout period")
	}
	if alert.MockAlertExecutions.Contains(fmt.Sprintf("%s/%s", brapper.ID, quietTagCheckID)) {
		t.Error("Alert executed under quiet period")
	}
}

func BenchmarkPostCheckResult(b *testing.B) {
	ctx, err := conf.ContextFactory.Get()
	if err != nil {
		b.Error(err)
	}
	brapper, _ := ctx.SubjectRepo().Named("bootstrapper")
	checkTmp, _ := ctx.CheckRepo().Search("Test OK", "", "")
	testOK := checkTmp[0]
	ctx.Close()
	b.ResetTimer()

	b.Run("ok", func(bb *testing.B) {
		bb.RunParallel(func(pb *testing.PB) {
			var body string
			for pb.Next() {
				body = fmt.Sprintf(`{
						"SubjectID": "%s",
						"CheckID": "%s",
						"Time": "%s",
						"Status": 1
					}`,
					brapper.ID, testOK.ID, time.Now().Format(time.RFC3339))
				testRoute("POST", "/checkresults", body)
			}
		})
	})

	b.Run("alert", func(bb *testing.B) {
		bb.RunParallel(func(pb *testing.PB) {
			var body string
			for pb.Next() {
				body = fmt.Sprintf(`{
						"SubjectID": "%s",
						"CheckID": "%s",
						"Time": "%s",
						"Status": 3
					}`,
					brapper.ID, testOK.ID, time.Now().Format(time.RFC3339))
				testRoute("POST", "/checkresults", body)
			}
		})
	})
}

func TestCheckSearch(t *testing.T) {
	t.Run("all", func(tt *testing.T) {
		method := "GET"
		route := "/checks"
		status, body := testRoute(method, route, "")
		if status != 200 {
			tt.Errorf("%s %s: Expected: %d, Actual: %d", method, route, 200, status)
		}
		payload := []model.Check{}
		err := json.NewDecoder(strings.NewReader(body)).Decode(&payload)
		if err != nil {
			tt.Errorf("%s %s: Failed to decode body:\n\t%s\n\t%s", method, route, err, body)
		}
		expected := 8
		actual := len(payload)
		if expected != actual {
			tt.Errorf("%s %s: Expected %d checks, actual %d checks", method, route, expected, actual)
		}
	})
	t.Run("name", func(tt *testing.T) {
		method := "GET"
		route := "/checks?name=test"
		status, body := testRoute(method, route, "")
		if status != 200 {
			tt.Errorf("%s %s: Expected: %d, Actual: %d", method, route, 200, status)
		}
		payload := []model.Check{}
		err := json.NewDecoder(strings.NewReader(body)).Decode(&payload)
		if err != nil {
			tt.Errorf("%s %s: Failed to decode body:\n\t%s\n\t%s", method, route, err, body)
		}
		expected := 3
		actual := len(payload)
		if expected != actual {
			tt.Errorf("%s %s: Expected %d checks, actual %d checks", method, route, expected, actual)
		}
	})
	t.Run("tag", func(tt *testing.T) {
		method := "GET"
		route := "/checks?tag=default"
		status, body := testRoute(method, route, "")
		if status != 200 {
			tt.Errorf("%s %s: Expected: %d, Actual: %d", method, route, 200, status)
		}
		payload := []model.Check{}
		err := json.NewDecoder(strings.NewReader(body)).Decode(&payload)
		if err != nil {
			tt.Errorf("%s %s: Failed to decode body:\n\t%s\n\t%s", method, route, err, body)
		}
		expected := 6
		actual := len(payload)
		if expected != actual {
			tt.Errorf("%s %s: Expected %d checks, actual %d checks", method, route, expected, actual)
		}
	})
	t.Run("role", func(tt *testing.T) {
		method := "GET"
		route := "/checks?role=unhealthy"
		status, body := testRoute(method, route, "")
		if status != 200 {
			tt.Errorf("%s %s: Expected: %d, Actual: %d", method, route, 200, status)
		}
		payload := []model.Check{}
		err := json.NewDecoder(strings.NewReader(body)).Decode(&payload)
		if err != nil {
			tt.Errorf("%s %s: Failed to decode body:\n\t%s\n\t%s", method, route, err, body)
		}
		expected := 1
		actual := len(payload)
		if expected != actual {
			tt.Errorf("%s %s: Expected %d checks, actual %d checks", method, route, expected, actual)
		}
	})
}

func TestCheckCrud(t *testing.T) {
	var id uuid.UUID
	t.Run("create", func(tt *testing.T) {
		method := "POST"
		route := "/checks"
		body := `{"Name": "CRUD", "Type": 1, "Parameters":{}, "Interval": 13, "Roles": [], "Tags": []}`
		status, body := testRoute(method, route, body)
		if status != http.StatusCreated {
			tt.Errorf("%s %s: Expected: %d, Actual: %d - %s", method, route, http.StatusOK, status, body)
		}
		payload := model.Check{}
		err := json.NewDecoder(strings.NewReader(body)).Decode(&payload)
		if err != nil {
			tt.Errorf("%s %s: Failed to decode body:\n\t%s\n\t%s", method, route, err, body)
		}
		id = payload.ID
		expected := "CRUD"
		actual := payload.Name
		if expected != actual {
			tt.Errorf("%s %s: Expected %s, actual %s", method, route, expected, actual)
		}
	})
	t.Run("retrieve", func(tt *testing.T) {
		if id == uuid.Nil {
			tt.Skip("Skipping because create failed.")
		}
		method := "GET"
		route := "/checks/" + id.String()
		status, body := testRoute(method, route, "")
		if status != http.StatusOK {
			tt.Errorf("%s %s: Expected: %d, Actual: %d - %s", method, route, http.StatusOK, status, body)
		}
		payload := model.Check{}
		err := json.NewDecoder(strings.NewReader(body)).Decode(&payload)
		if err != nil {
			tt.Errorf("%s %s: Failed to decode body:\n\t%s\n\t%s", method, route, err, body)
		}
		expected := "CRUD"
		actual := payload.Name
		if expected != actual {
			tt.Errorf("%s %s: Expected %s, actual %s", method, route, expected, actual)
		}
	})
	t.Run("update", func(tt *testing.T) {
		if id == uuid.Nil {
			tt.Skip("Skipping because create failed.")
		}
		method := "PUT"
		route := "/checks/" + id.String()
		body := fmt.Sprintf(`{"ID":"%s","Name": "SCRUD", "Type": 1, "Parameters":{}, "Interval": 13, "Roles": [], "Tags": []}`, id)
		status, body := testRoute(method, route, body)
		if status != http.StatusNoContent {
			tt.Errorf("%s %s: Expected: %d, Actual: %d - %s", method, route, http.StatusNoContent, status, body)
			return
		}
		method = "GET"
		route = "/checks/" + id.String()
		status, body = testRoute(method, route, "")
		if status != http.StatusOK {
			tt.Errorf("%s %s: Expected: %d, Actual: %d - %s", method, route, http.StatusOK, status, body)
		}
		payload := model.Check{}
		err := json.NewDecoder(strings.NewReader(body)).Decode(&payload)
		if err != nil {
			tt.Errorf("%s %s: Failed to decode body:\n\t%s\n\t%s", method, route, err, body)
		}
		expected := "SCRUD"
		actual := payload.Name
		if expected != actual {
			tt.Errorf("%s %s: Expected %s, actual %s", method, route, expected, actual)
		}
	})
	t.Run("delete", func(tt *testing.T) {
		if id == uuid.Nil {
			tt.Skip("Skipping because create failed.")
		}
		method := "DELETE"
		route := "/checks/" + id.String()
		status, _ := testRoute(method, route, "")
		if status != http.StatusNoContent {
			tt.Errorf("%s %s: Expected: %d, Actual: %d", method, route, http.StatusNoContent, status)
			return
		}
		method = "GET"
		route = "/checks/" + id.String()
		status, _ = testRoute(method, route, "")
		if status != http.StatusNotFound && status != http.StatusGone {
			tt.Errorf("%s %s: Expected: %d/%d, Actual: %d", method, route, http.StatusNotFound, http.StatusGone, status)
		}
	})
}

/*** Test Harness ***/
var (
	dbName  string
	handler http.Handler
	conf    config.Configuration
)
var blackoutRoleCheckID, blackoutTagCheckID, mockAlertID, blackoutRolePeriodID, blackoutTagPeriodID, quietTagPeriodID, quietTagCheckID uuid.UUID

func TestMain(m *testing.M) {
	// Prepare DB
	dbName = fmt.Sprintf("test_%d", time.Now().Nanosecond())
	log.Printf("Testing with DB: %s", dbName)
	conf = config.New()
	conf.Address = "127.0.0.1"
	conf.MongoDatabase = dbName
	conf.Init()
	defer conf.ContextFactory.Close()

	ctx, err := conf.ContextFactory.Get()
	if err != nil {
		log.Fatal(err)
	}
	database.FillSampleData(ctx)
	fillTestData(ctx)
	ctx.Close()

	// Prepare server
	handler = NewObservatoryMux(&conf)
	conf.Up = true
	retCode := m.Run()

	// Clean up
	ctx, err = conf.ContextFactory.Get()
	if err != nil {
		log.Println(err)
	}
	err = ctx.(*mongo.AppContext).DropDatabase()
	if err != nil {
		log.Println(err)
	}
	ctx.Close()
	os.Exit(retCode)
}

func fillTestData(ctx model.AppContext) {
	c := model.Check{
		Name:     "Blackout role check",
		Type:     model.CheckExec,
		Interval: 10,
		Roles:    []string{"default"},
		Tags:     []string{"default"},
		Modified: time.Now(),
	}
	ctx.CheckRepo().Create(&c)
	blackoutRoleCheckID = c.ID

	c = model.Check{
		Name:     "Blackout tag check",
		Type:     model.CheckExec,
		Interval: 10,
		Roles:    []string{"default", "healthy"},
		Tags:     []string{"blackout"},
		Modified: time.Now(),
	}
	ctx.CheckRepo().Create(&c)
	blackoutTagCheckID = c.ID

	c = model.Check{
		Name:     "Quiet tag check",
		Type:     model.CheckExec,
		Interval: 10,
		Roles:    []string{"default", "healthy"},
		Tags:     []string{"quiet"},
		Modified: time.Now(),
	}
	ctx.CheckRepo().Create(&c)
	quietTagCheckID = c.ID

	a := model.Alert{
		Name:     "Mock alert",
		Type:     model.AlertMock,
		Roles:    []string{"healthy", "unhealthy", "default"},
		Tags:     []string{"default"},
		Modified: time.Now(),
	}
	ctx.AlertRepo().Create(&a)
	mockAlertID = a.ID

	p := model.Period{
		Name:     "Blackout role test",
		Type:     model.PeriodBlackout,
		Modified: time.Now(),
		Start:    time.Now(),
		End:      time.Now().Add(time.Hour),
		Roles:    []string{"blackout"},
	}
	ctx.PeriodRepo().Create(&p)
	blackoutRolePeriodID = p.ID

	p = model.Period{
		Name:     "Blackout tag test",
		Type:     model.PeriodBlackout,
		Modified: time.Now(),
		Start:    time.Now(),
		End:      time.Now().Add(time.Hour),
		Tags:     []string{"blackout"},
	}
	ctx.PeriodRepo().Create(&p)
	blackoutTagPeriodID = p.ID

	p = model.Period{
		Name:     "Quiet tag test",
		Type:     model.PeriodQuiet,
		Modified: time.Now(),
		Start:    time.Now(),
		End:      time.Now().Add(time.Hour),
		Tags:     []string{"quiet"},
	}
	ctx.PeriodRepo().Create(&p)
	quietTagPeriodID = p.ID
}

func testRoute(method, route, reqBody string) (status int, respBody string) {
	r := httptest.NewRequest(method, route, bytes.NewReader([]byte(reqBody)))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w.Code, w.Body.String()
}

type testCase struct {
	Name       string
	Method     string
	Route      string
	ReqBody    string
	ReqPayload interface{}
	Status     int
	RespBody   string
	RespRegex  string
	AsyncWait  time.Duration
}

func execRouteTests(ot *testing.T, tests []testCase) {
	for _, tt := range tests {
		ot.Run(tt.Name, func(t *testing.T) {
			var reqBody string
			if tt.ReqPayload == nil {
				reqBody = tt.ReqBody
			} else {
				bytes, err := json.Marshal(tt.ReqPayload)
				if err != nil {
					t.Errorf("json.Marshal failed: %s", err)
					return
				}
				reqBody = string(bytes)
			}
			s, b := testRoute(tt.Method, tt.Route, reqBody)
			b = strings.TrimSpace(b)
			if tt.RespRegex != "" {
				re := regexp.MustCompile(tt.RespRegex)
				if s != tt.Status || !re.Match([]byte(b)) {
					t.Errorf("%s %s:\n\tBody: %s\n\tExpected: %d /%s/\n\tActual: %d %s",
						tt.Method, tt.Route, tt.ReqBody,
						tt.Status, tt.RespRegex,
						s, b)
				}
			} else if tt.RespBody != "" {
				if s != tt.Status || tt.RespBody != "" && b != tt.RespBody {
					t.Errorf("%s %s:\n\tBody: %s\n\tExpected: %d %q\n\tActual: %d %q",
						tt.Method, tt.Route, tt.ReqBody,
						tt.Status, tt.RespBody,
						s, b)
				}
			}
		})
		time.Sleep(tt.AsyncWait)
	}
}
