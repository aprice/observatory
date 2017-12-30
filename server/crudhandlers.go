package server

import (
	"fmt"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/aprice/observatory/actions"
	"github.com/aprice/observatory/model"
	"github.com/aprice/observatory/server/config"
)

/** Check CRUD **/
type checksCrud struct {
	conf *config.Configuration
}

func (c checksCrud) entity() interface{} {
	return &model.Check{}
}

func (c checksCrud) search(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return nil, err
	}
	defer ctx.Close()
	name := r.URL.Query().Get("name")
	role := r.URL.Query().Get("role")
	tag := r.URL.Query().Get("tag")
	return ctx.CheckRepo().Search(name, role, tag)
}

func (c checksCrud) create(w http.ResponseWriter, r *http.Request, entity interface{}) (string, error) {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return "", err
	}
	defer ctx.Close()
	check := entity.(*model.Check)
	check.Modified = time.Now()
	err = ctx.CheckRepo().Create(check)
	return c.conf.URLForPath("checks/" + check.ID.String()), err
}

func (c checksCrud) retrieve(w http.ResponseWriter, r *http.Request, id uuid.UUID) (interface{}, error) {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return nil, err
	}
	defer ctx.Close()
	return ctx.CheckRepo().Find(id)
}

func (c checksCrud) update(w http.ResponseWriter, r *http.Request, id uuid.UUID, entity interface{}) error {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return err
	}
	defer ctx.Close()
	check := entity.(*model.Check)

	if id != check.ID {
		return fmt.Errorf("URL ID %s and body ID %s do not match", id.String(), check.ID.String())
	}
	dbCheck, err := ctx.CheckRepo().Find(check.ID)
	if err != nil {
		return err
	}
	check.Modified = time.Now()
	err = ctx.CheckRepo().Update(*check)
	if err != nil {
		return err
	}

	// Execute cleanup async so we can respond immediately
	go actions.UpdatedCheckCleanup(*c.conf, check.ID, dbCheck.Roles, check.Roles)
	return nil
}

func (c checksCrud) delete(w http.ResponseWriter, r *http.Request, id uuid.UUID) error {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return err
	}
	defer ctx.Close()
	err = ctx.CheckRepo().Delete(id)
	if err != nil {
		return err
	}
	// Execute cleanup async so we can respond immediately
	go actions.DeletedCheckCleanup(*c.conf, id)
	return nil
}

/** Subject CRUD **/
type subjectsCrud struct {
	conf *config.Configuration
}

func (c subjectsCrud) entity() interface{} {
	return &model.Subject{}
}

func (c subjectsCrud) search(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return nil, err
	}
	defer ctx.Close()
	name := r.URL.Query().Get("name")
	role := r.URL.Query().Get("role")
	return ctx.SubjectRepo().Search(name, role)
}

func (c subjectsCrud) create(w http.ResponseWriter, r *http.Request, entity interface{}) (string, error) {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return "", err
	}
	defer ctx.Close()
	subject := entity.(*model.Subject)
	subject.Modified = time.Now()
	err = ctx.SubjectRepo().Create(subject)
	return c.conf.URLForPath("subjects/" + subject.ID.String()), err
}

func (c subjectsCrud) retrieve(w http.ResponseWriter, r *http.Request, id uuid.UUID) (interface{}, error) {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return nil, err
	}
	defer ctx.Close()
	return ctx.SubjectRepo().Find(id)
}

func (c subjectsCrud) update(w http.ResponseWriter, r *http.Request, id uuid.UUID, entity interface{}) error {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return err
	}
	defer ctx.Close()
	subject := entity.(*model.Subject)

	if id != subject.ID {
		return fmt.Errorf("URL ID %s and body ID %s do not match", id.String(), subject.ID.String())
	}
	dbSubject, err := ctx.SubjectRepo().Find(subject.ID)
	if err != nil {
		return err
	}

	subject.Modified = time.Now()
	err = ctx.SubjectRepo().Update(*subject)
	if err != nil {
		return err
	}

	// Execute cleanup async so we can respond immediately
	go actions.UpdatedSubjectCleanup(*c.conf, id, subject.Roles, dbSubject.Roles)
	return nil
}

func (c subjectsCrud) delete(w http.ResponseWriter, r *http.Request, id uuid.UUID) error {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return err
	}
	defer ctx.Close()
	err = ctx.SubjectRepo().Delete(id)
	if err != nil {
		return err
	}

	// Execute cleanup async so we can respond immediately
	go actions.DeletedSubjectCleanup(*c.conf, id)
	return nil
}

/** Alert CRUD **/
type alertsCrud struct {
	conf *config.Configuration
}

func (c alertsCrud) entity() interface{} {
	return &model.Alert{}
}

func (c alertsCrud) search(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return nil, err
	}
	defer ctx.Close()
	name := r.URL.Query().Get("name")
	role := r.URL.Query().Get("role")
	tag := r.URL.Query().Get("tag")
	return ctx.AlertRepo().Search(name, role, tag)
}

func (c alertsCrud) create(w http.ResponseWriter, r *http.Request, entity interface{}) (string, error) {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return "", err
	}
	defer ctx.Close()
	alert := entity.(*model.Alert)
	alert.Modified = time.Now()
	err = ctx.AlertRepo().Create(alert)
	return c.conf.URLForPath("alerts/" + alert.ID.String()), err
}

func (c alertsCrud) retrieve(w http.ResponseWriter, r *http.Request, id uuid.UUID) (interface{}, error) {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return nil, err
	}
	defer ctx.Close()
	return ctx.AlertRepo().Find(id)
}

func (c alertsCrud) update(w http.ResponseWriter, r *http.Request, id uuid.UUID, entity interface{}) error {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return err
	}
	defer ctx.Close()
	alert := entity.(*model.Alert)

	if id != alert.ID {
		return fmt.Errorf("URL ID %s and body ID %s do not match", id.String(), alert.ID.String())
	}
	alert.Modified = time.Now()
	err = ctx.AlertRepo().Update(*alert)
	if err != nil {
		return err
	}
	return nil
}

func (c alertsCrud) delete(w http.ResponseWriter, r *http.Request, id uuid.UUID) error {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return err
	}
	defer ctx.Close()
	err = ctx.AlertRepo().Delete(id)
	if err != nil {
		return err
	}
	return nil
}

/** Period CRUD **/
type periodsCrud struct {
	conf *config.Configuration
}

func (c periodsCrud) entity() interface{} {
	return &model.Period{}
}

func (c periodsCrud) search(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return nil, err
	}
	defer ctx.Close()
	name := r.URL.Query().Get("name")
	role := r.URL.Query().Get("role")
	tag := r.URL.Query().Get("tag")
	return ctx.PeriodRepo().Search(name, role, tag)
}

func (c periodsCrud) create(w http.ResponseWriter, r *http.Request, entity interface{}) (string, error) {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return "", err
	}
	defer ctx.Close()
	period := entity.(*model.Period)
	period.Modified = time.Now()
	err = ctx.PeriodRepo().Create(period)
	return c.conf.URLForPath("periods/" + period.ID.String()), err
}

func (c periodsCrud) retrieve(w http.ResponseWriter, r *http.Request, id uuid.UUID) (interface{}, error) {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return nil, err
	}
	defer ctx.Close()
	return ctx.PeriodRepo().Find(id)
}

func (c periodsCrud) update(w http.ResponseWriter, r *http.Request, id uuid.UUID, entity interface{}) error {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return err
	}
	defer ctx.Close()
	period := entity.(*model.Period)

	if id != period.ID {
		return fmt.Errorf("URL ID %s and body ID %s do not match", id.String(), period.ID.String())
	}
	period.Modified = time.Now()
	err = ctx.PeriodRepo().Update(*period)
	if err != nil {
		return err
	}
	return nil
}

func (c periodsCrud) delete(w http.ResponseWriter, r *http.Request, id uuid.UUID) error {
	ctx, err := c.conf.ContextFactory.Get()
	if err != nil {
		return err
	}
	defer ctx.Close()
	err = ctx.PeriodRepo().Delete(id)
	if err != nil {
		return err
	}
	return nil
}
