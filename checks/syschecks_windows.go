package checks

import (
	"bytes"
	"log"
	"os/exec"
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
	cpuInt, err := getCPUWin()
	cpuUse := float64(cpuInt)
	if err != nil {
		return model.StatusFailed, err
	}
	log.Printf("CPU: %.0f%% used", cpuUse)
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

func getCPUWin() (int64, error) {
	out, err := exec.Command("wmic", "cpu", "get", "loadpercentage").Output()
	if err != nil {
		return -1, err
	}

	bb := bytes.Split(out, []byte("\n"))
	b := bytes.TrimSpace(bb[1])
	return strconv.ParseInt(string(b), 10, 32)
}
