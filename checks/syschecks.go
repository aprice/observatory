//+build !windows

package checks

import (
	"log"
	"strconv"

	"github.com/aprice/observatory/model"
	sigar "github.com/cloudfoundry/gosigar"
	uuid "github.com/satori/go.uuid"
)

func loadCheck(subjectID uuid.UUID, params map[string]string) (model.CheckStatus, error) {
	cpu := sigar.Cpu{}
	err := cpu.Get()
	if err != nil {
		return model.StatusFailed, nil
	}
	cpuFree := float64(cpu.Idle) / float64(cpu.Total())
	cpuUse := (1.0 - cpuFree) * 100.0
	log.Printf("CPU: %d idle, %d total, %.1f%% used", cpu.Idle, cpu.Total(), cpuUse)
	thCrit, err := strconv.ParseFloat(params["warning"], 64)
	if err != nil {
		return model.StatusFailed, err
	}
	if thCrit < cpuUse {
		return model.StatusCritical, nil
	}
	thWarn, err := strconv.ParseFloat(params["critical"], 64)
	if err != nil {
		return model.StatusFailed, err
	}
	if thWarn < cpuUse {
		return model.StatusWarning, nil
	}
	return model.StatusOK, nil
}
