package checks

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"

	sigar "github.com/cloudfoundry/gosigar"
	uuid "github.com/satori/go.uuid"

	"github.com/aprice/observatory/model"
	"github.com/aprice/observatory/utils"
)

func executeCheck(subjectID uuid.UUID, params map[string]string) (model.CheckStatus, error) {
	args := utils.StringToArgs(params["command"])
	status, _, _ := utils.Execute(args...)
	var checkStatus model.CheckStatus
	if status == 0 {
		checkStatus = model.StatusOK
	} else if status == 1 {
		checkStatus = model.StatusWarning
	} else {
		checkStatus = model.StatusCritical
	}
	return checkStatus, nil
}

var checkClient = &http.Client{}

func httpCheck(subjectID uuid.UUID, params map[string]string) (model.CheckStatus, error) {
	req, err := http.NewRequest("GET", params["url"], nil)
	if err != nil {
		return model.StatusNone, err
	}
	resp, err := checkClient.Do(req)
	if resp != nil {
		defer func() {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}()
	}
	if err != nil || resp.StatusCode >= 400 {
		return model.StatusCritical, err
	}
	return model.StatusOK, nil
}

func memCheck(subjectID uuid.UUID, params map[string]string) (model.CheckStatus, error) {
	mem := sigar.Mem{}
	swap := sigar.Swap{}
	if err := mem.Get(); err != nil {
		return model.StatusFailed, err
	}
	if err := swap.Get(); err != nil {
		return model.StatusFailed, err
	}
	pctUsed := (float64(mem.Used) / float64(mem.Total)) * 100
	pctSwap := (float64(swap.Used) / float64(swap.Total)) * 100
	log.Printf("Memory: %s/%s (%.1f%%), Swap: %s/%s (%.1f%%)",
		utils.HumanReadableBytesSI(int64(mem.Used), 3),
		utils.HumanReadableBytesSI(int64(mem.Total), 3),
		pctUsed,
		utils.HumanReadableBytesSI(int64(swap.Used), 3),
		utils.HumanReadableBytesSI(int64(swap.Total), 3),
		pctSwap)
	if tholdRaw, ok := params["usedcrit"]; ok {
		thold, err := strconv.ParseFloat(tholdRaw, 64)
		if err != nil {
			return model.StatusFailed, err
		}
		if pctUsed > thold {
			return model.StatusCritical, nil
		}
	}
	if tholdRaw, ok := params["usedwarn"]; ok {
		thold, err := strconv.ParseFloat(tholdRaw, 64)
		if err != nil {
			return model.StatusFailed, err
		}
		if pctUsed > thold {
			return model.StatusWarning, nil
		}
	}
	if tholdRaw, ok := params["swapcrit"]; ok {
		thold, err := strconv.ParseFloat(tholdRaw, 64)
		if err != nil {
			return model.StatusFailed, err
		}
		if pctSwap > thold {
			return model.StatusCritical, nil
		}
	}
	if tholdRaw, ok := params["swapwarn"]; ok {
		thold, err := strconv.ParseFloat(tholdRaw, 64)
		if err != nil {
			return model.StatusFailed, err
		}
		if pctSwap > thold {
			return model.StatusWarning, nil
		}
	}
	return model.StatusOK, nil
}

func diskCheck(subjectID uuid.UUID, params map[string]string) (model.CheckStatus, error) {
	var err error
	fsName := params["filesystem"]
	acRaw := params["critical"]
	awRaw := params["warning"]
	ac, err := strconv.ParseFloat(acRaw, 64)
	if err != nil {
		return model.StatusFailed, err
	}
	aw, err := strconv.ParseFloat(awRaw, 64)
	if err != nil {
		return model.StatusFailed, err
	}

	usage := sigar.FileSystemUsage{}
	err = usage.Get(fsName)
	if err != nil {
		return model.StatusFailed, err
	}
	usedPct := usage.UsePercent()
	log.Printf("%s: %s/%s (%.1f%%)",
		fsName,
		utils.HumanReadableBytesSI(int64(usage.Used), 3),
		utils.HumanReadableBytesSI(int64(usage.Total), 3),
		usedPct)
	if ac < usedPct {
		return model.StatusCritical, nil
	}
	if aw < usedPct {
		return model.StatusWarning, nil
	}
	return model.StatusOK, nil
}

func portCheck(subjectID uuid.UUID, params map[string]string) (model.CheckStatus, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%s", params["port"]))
	if err != nil {
		return model.StatusCritical, err
	}
	conn.Close()
	return model.StatusOK, nil
}
