package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aprice/observatory/checks"
	"github.com/aprice/observatory/client"
	"github.com/aprice/observatory/model"
	"github.com/aprice/observatory/utils"
)

// Reconfigure agent in a loop.
func Reconfigure(name, initialEndpoint, initialRoles string, quit utils.SentinelChannel) {
	config := &model.AgentConfig{}
	endpointSanitized := url.QueryEscape(initialEndpoint)
	endpointSanitized = strings.Replace(endpointSanitized, "%3A", ":", -1)
	endpoints := []string{endpointSanitized}
	interval := time.Duration(0)
	headers := make(map[string]string)
	url := fmt.Sprintf("/configuration/%s?roles=%s", name, initialRoles)
	expires, err := client.GetObject(url, headers, endpoints, config)
	if err != nil {
		log.Fatalf("Failed to get initial configuration: %s\n", err)
	} else {
		endpoints = config.Coordinators
		checks.UpdateChecks(*config)
		interval = expires.Sub(time.Now())
		log.Printf("Configured and running.\nConfiguration (updates in %v): %v", interval, *config)
		reconfigureLoop(*config, interval, quit)
	}
}

func reconfigureLoop(initConfig model.AgentConfig, interval time.Duration, quit utils.SentinelChannel) {
	config := &initConfig
	lastUpdate := time.Now()
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		select {
		case <-quit:
			log.Println("Reconfiguration stopped.")
			timer.Stop()
			return
		case <-timer.C:
			url := fmt.Sprintf("/configuration/%s", config.Name)
			headers := make(map[string]string)
			headers["If-Modified-Since"] = lastUpdate.Format(http.TimeFormat)
			expires, err := client.GetObject(url, headers, config.Coordinators, config)
			if err != nil {
				log.Println(err)
			} else {
				lastUpdate = time.Now()
				checks.UpdateChecks(*config)
				interval = expires.Sub(time.Now())
			}
			timer = time.NewTimer(interval)
		}
	}
}
