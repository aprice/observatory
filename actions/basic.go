package actions

import (
	"log"
	"time"

	"github.com/aprice/observatory"
	"github.com/aprice/observatory/collections"
	"github.com/aprice/observatory/model"
	"github.com/aprice/observatory/server/config"
	uuid "github.com/satori/go.uuid"
)

// Info returns information about the running Coordinator instance.
func Info(conf config.Configuration) model.CoordinatorInfo {
	return model.CoordinatorInfo{
		ID:         conf.ID,
		Version:    observatory.Version,
		Build:      observatory.Build,
		APIVersion: observatory.APIVersion,
		Leader:     conf.IsLeader(),
	}
}

// DataStats returns counts of each entity type.
func DataStats(conf config.Configuration) (map[string]int, error) {
	ctx, err := conf.ContextFactory.Get()
	if err != nil {
		return nil, err
	}
	defer ctx.Close()

	ds := make(map[string]int, 8)

	val, err := ctx.SubjectRepo().Count()
	if err != nil {
		return nil, err
	}
	ds["Subjects"] = val

	val, err = ctx.RoleRepo().CountRoles()
	if err != nil {
		return nil, err
	}
	ds["Roles"] = val

	val, err = ctx.CheckRepo().Count()
	if err != nil {
		return nil, err
	}
	ds["Checks"] = val

	val, err = ctx.TagRepo().CountTags()
	if err != nil {
		return nil, err
	}
	ds["Tags"] = val

	val, err = ctx.AlertRepo().Count()
	if err != nil {
		return nil, err
	}
	ds["Alerts"] = val

	val, err = ctx.PeriodRepo().Count()
	if err != nil {
		return nil, err
	}
	ds["Periods"] = val

	val, err = ctx.CheckStateRepo().Count()
	if err != nil {
		return nil, err
	}
	ds["CheckStates"] = val

	val, err = ctx.CheckResultRepo().Count()
	if err != nil {
		return nil, err
	}
	ds["CheckResults"] = val

	return ds, nil
}

var configAffectingPeriods = []model.PeriodType{model.PeriodBlackout}

// ConfigureAgent builds an agent config, creating the Subject if necessary.
func ConfigureAgent(conf config.Configuration, name string, roles []string) (model.AgentConfig, error) {
	ctx, err := conf.ContextFactory.Get()
	if err != nil {
		return model.AgentConfig{}, err
	}
	defer ctx.Close()
	// Init
	subject, err := ctx.SubjectRepo().Named(name)
	if err == nil {
		subject.LastCheckIn = time.Now()
		err = ctx.SubjectRepo().Update(subject)
		if err != nil {
			log.Println(err)
		}
	} else if err == model.ErrNotFound {
		// Save new subject
		subject = model.Subject{Name: name, Roles: roles, LastCheckIn: time.Now()}
		err = ctx.SubjectRepo().Create(&subject)
		if err != nil {
			log.Println(err)
		}
	} else {
		return model.AgentConfig{}, err
	}

	periods, err := ctx.PeriodRepo().FindForSubject(subject, configAffectingPeriods)
	var checks []model.Check
	if len(periods) > 0 {
		checks = []model.Check{}
	} else {
		// Look up checks
		allChecks, err := ctx.CheckRepo().OfTypesForRoles(model.LocalCheckTypes, subject.Roles)
		if err != nil && err != model.ErrNotFound {
			return model.AgentConfig{}, err
		}
		tagsUsed := make(collections.StringSet, len(allChecks))
		for _, check := range allChecks {
			tagsUsed.Add(check.Tags...)
		}
		periods, err = ctx.PeriodRepo().FindForSubjectChecks(subject, tagsUsed.ToArray(), configAffectingPeriods)
		if err != nil && err != model.ErrNotFound {
			return model.AgentConfig{}, err
		}
		if len(periods) == 0 {
			checks = allChecks
		} else {
			blackoutTags := make(collections.StringSet, len(periods))
			for _, period := range periods {
				blackoutTags.Add(period.Tags...)
			}
			for _, check := range allChecks {
				if !blackoutTags.ContainsAny(check.Tags...) {
					checks = append(checks, check)
				}
			}
		}
	}

	// Configure
	coordinators := append(conf.Peers.AlivePeerSet().EndpointArray(), conf.Endpoint())
	agentConf := model.NewAgentConfig(subject, checks, periods, coordinators)

	return agentConf, nil
}

// AddPeer registers a new peer coordinator.
func AddPeer(conf config.Configuration, iam uuid.UUID, ep string) {
	if iam != uuid.Nil && ep != "" {
		kp := conf.Peers.KnownPeerSet()
		if _, known := kp[iam]; !known && iam != conf.ID {
			conf.Peers.AddPeer(ep)
		}
	}
}

// GetPeers returns all currently known live peers. If an id of the requesting
// coordinator is provided, it will be added to this node's known peers.
func GetPeers(conf config.Configuration) map[string]string {
	peers := map[string]string{conf.ID.String(): conf.Endpoint()}
	for id, url := range conf.Peers.AlivePeerSet() {
		peers[id.String()] = url
	}
	return peers
}
