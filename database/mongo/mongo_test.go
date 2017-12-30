package mongo

import (
	"testing"

	"github.com/aprice/observatory/model"
)

// Validate that context & repos implement the desired interfaces.
func TestInterfaceImplementation(t *testing.T) {
	var _ model.AppContextFactory = (*AppContextFactory)(nil)
	var _ model.AppContext = (*AppContext)(nil)
	var _ model.SubjectRepo = (*SubjectRepo)(nil)
	var _ model.CheckRepo = (*CheckRepo)(nil)
	var _ model.AlertRepo = (*AlertRepo)(nil)
	var _ model.PeriodRepo = (*PeriodRepo)(nil)
	var _ model.TagRepo = (*CheckRepo)(nil)
	var _ model.RoleRepo = (*SubjectRepo)(nil)
	var _ model.CheckStateRepo = (*CheckStateRepo)(nil)
	var _ model.CheckResultRepo = (*CheckResultRepo)(nil)
}
