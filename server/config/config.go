package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/osext"
	"github.com/satori/go.uuid"

	"github.com/aprice/observatory/database/mongo"
	"github.com/aprice/observatory/model"
	"github.com/aprice/observatory/utils"
)

// Configuration describes the configuration of the coordinator instance.
type Configuration struct {
	ID             uuid.UUID
	Up             bool
	Peers          *Peers
	ContextFactory model.AppContextFactory

	Port                      int
	Address                   string
	AllowCors                 string
	AgentUpdateInterval       int
	PeerUpdateInterval        int
	PeerCheckInterval         int
	RemoteCheckUpdateInterval int
	RemoteCheckAssignInterval int
	MongoHost                 string
	MongoDatabase             string
	MongoUser                 string
	MongoPassword             string
	BootstrapPeers            []string
	SMTPHost                  string
	SMTPPort                  int
	SMTPUser                  string
	SMTPPassword              string
	EmailFrom                 string
}

// New produces a Configuration filled with defaults.
func New() Configuration {
	ip, err := getLocalIP()
	if err != nil {
		ip = "127.0.0.1"
	}
	id := utils.NewTimeUUID()
	return Configuration{
		ID:                        id,
		Port:                      13100,
		Address:                   ip,
		AgentUpdateInterval:       30,
		PeerUpdateInterval:        30,
		PeerCheckInterval:         5,
		RemoteCheckUpdateInterval: 20,
		RemoteCheckAssignInterval: 60,
		MongoHost:                 "localhost",
		MongoDatabase:             "Observatory",
		BootstrapPeers:            []string{},
		Peers:                     NewPeers(),
	}
}

// AgentUpdateDuration from AgentUpdateInterval
func (c Configuration) AgentUpdateDuration() time.Duration {
	return time.Duration(c.AgentUpdateInterval) * time.Second
}

// PeerUpdateDuration from PeerUpdateInterval
func (c Configuration) PeerUpdateDuration() time.Duration {
	return time.Duration(c.PeerUpdateInterval) * time.Second
}

// PeerCheckDuration from PeerCheckInterval
func (c Configuration) PeerCheckDuration() time.Duration {
	return time.Duration(c.PeerCheckInterval) * time.Second
}

// RemoteCheckUpdateDuration from RemoteCheckUpdateInterval
func (c Configuration) RemoteCheckUpdateDuration() time.Duration {
	return time.Duration(c.RemoteCheckUpdateInterval) * time.Second
}

// RemoteCheckAssignDuration from RemoteCheckAssignInterval
func (c Configuration) RemoteCheckAssignDuration() time.Duration {
	return time.Duration(c.RemoteCheckAssignInterval) * time.Second
}

// Endpoint address for this coordinator (Address:Port)
func (c Configuration) Endpoint() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

// URLForPath returns the full URL for a given API route path.
func (c Configuration) URLForPath(path string) string {
	return fmt.Sprintf("http://%s/%s", c.Endpoint(), path)
}

// IsLeader returns true if this instance is the cluster leader.
func (c Configuration) IsLeader() bool {
	return c.Peers.AlivePeerSet().IsLeader(c.ID)
}

// Init initializes configuration based on current settings.
func (c *Configuration) Init() {
	// Peers
	c.Peers.Run(*c)

	// DB
	var err error
	c.ContextFactory, err = mongo.InitConnection(c.MongoHost, c.MongoDatabase, c.MongoUser, c.MongoPassword)
	if err != nil {
		log.Panic("failed to establish Mongo connection", err)
	}
}

// LoadConfig from file, overriding defaults with values from file.
func LoadConfig(filePath string) (Configuration, error) {
	config := New()
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

// DefaultConfigFilePath returns the context-sensitive default configuration
// file path.
func DefaultConfigFilePath() string {
	switch runtime.GOOS {
	case "linux":
		return "/etc/observatory/coordinator.conf.json"
	default:
		dir, err := osext.ExecutableFolder()
		if err != nil {
			return "./coordinator.conf.json"
		}
		return dir + "/coordinator.conf.json"
	}
}

// Cleaned up from - where else - http://stackoverflow.com/a/31551220/7426
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !strings.HasPrefix(ipnet.IP.String(), "169.254.") {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("No non-local IP found.")
}
