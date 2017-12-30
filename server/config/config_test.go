package config

import (
	"reflect"
	"testing"
	"time"
)

func TestConfigureFromFile(t *testing.T) {
	conf, err := LoadConfig("../test2.json")
	if err != nil {
		t.Error(err)
	}
	expected := Configuration{
		Port:                13101,
		Address:             "192.168.13.1",
		AgentUpdateInterval: 15,
		PeerUpdateInterval:  15,
		PeerCheckInterval:   5,
		MongoHost:           "localhost",
		MongoDatabase:       "Observatory",
		BootstrapPeers:      []string{"192.168.13.1:13100"},
	}
	if conf.Port != expected.Port {
		t.Errorf("Port: %d, expected: %d", conf.Port, expected.Port)
	}
	if conf.Address != expected.Address {
		t.Errorf("Address: %s, expected: %s", conf.Address, expected.Address)
	}
	// MongoHost isn't specified in test2, so ensure we get the default
	if conf.MongoHost != expected.MongoHost {
		t.Errorf("MongoHost: %s, expected: %s", conf.MongoHost, expected.MongoHost)
	}
	if !reflect.DeepEqual(conf.BootstrapPeers, expected.BootstrapPeers) {
		t.Errorf("BootstrapPeers: %v, expected: %v", conf.BootstrapPeers, expected.BootstrapPeers)
	}
	if conf.PeerUpdateInterval != expected.PeerUpdateInterval {
		t.Errorf("PeerUpdateInterval: %d, expected: %d", conf.PeerUpdateInterval, expected.PeerUpdateInterval)
	}
}

func TestHelpers(t *testing.T) {
	given := Configuration{
		Port:                13101,
		Address:             "192.168.13.1",
		AgentUpdateInterval: 15,
		PeerUpdateInterval:  15,
		PeerCheckInterval:   5,
	}
	expEndpoint := "192.168.13.1:13101"
	expAgUpDur := time.Duration(15) * time.Second
	expPeerUpDur := time.Duration(15) * time.Second
	expPeerCkDur := time.Duration(5) * time.Second
	if given.Endpoint() != expEndpoint {
		t.Errorf("Endpoint %s, expected %s", given.Endpoint(), expEndpoint)
	}
	if given.AgentUpdateDuration() != expAgUpDur {
		t.Errorf("AgentUpdateDuration %v, expected %v", given.AgentUpdateDuration(), expAgUpDur)
	}
	if given.PeerUpdateDuration() != expPeerUpDur {
		t.Errorf("PeerUpdateDuration %v, expected %v", given.PeerUpdateDuration(), expPeerUpDur)
	}
	if given.PeerCheckDuration() != expPeerCkDur {
		t.Errorf("PeerCheckDuration %v, expected %v", given.PeerCheckDuration(), expPeerCkDur)
	}
}
