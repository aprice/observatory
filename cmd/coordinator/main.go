package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	flag "github.com/ogier/pflag"

	"github.com/aprice/observatory"
	"github.com/aprice/observatory/remotecheck"
	"github.com/aprice/observatory/server"
	"github.com/aprice/observatory/server/config"
	"github.com/aprice/observatory/utils"
)

func main() {
	t1 := time.Now()
	conf := buildConfig()
	fmt.Println(observatory.VersionInfo())

	conf.Init()
	defer conf.ContextFactory.Close()

	peerQuit := make(utils.SentinelChannel)
	remoteQuit := make(utils.SentinelChannel)
	go server.Start(&conf)
	conf.Up = true
	go remotecheck.UpdateRemoteChecks(conf, remoteQuit)
	t2 := time.Now()
	log.Printf("Initialized in %v", t2.Sub(t1))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	s := <-c
	log.Println("Got signal: ", s)
	conf.Up = false
	go func() { peerQuit <- utils.Nothing }()
	go func() { remoteQuit <- utils.Nothing }()
	time.Sleep(time.Duration(1) * time.Second)
	os.Exit(0)
}

func buildConfig() config.Configuration {
	var (
		confFile string
		help     bool
		version  bool
		conf     config.Configuration
		port     int
		address  string
		peerList string
		err      error
	)

	defaults := config.New()
	cli := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cli.StringVarP(&confFile, "config", "c", config.DefaultConfigFilePath(), "Configuration file path")
	cli.BoolVarP(&help, "help", "h", false, "Print usage information")
	cli.BoolVarP(&version, "version", "v", false, "Print version information and exit")
	cli.IntVar(&port, "port", defaults.Port, "REST API port")
	cli.StringVar(&address, "address", defaults.Address, "Advertised REST API IP address or hostname")
	cli.StringVar(&peerList, "peers", "", "Comma-separated list of peers to bootstrap to")
	cli.Parse(os.Args[1:])

	if version {
		fmt.Println(observatory.VersionInfo())
		os.Exit(0)
	}

	if help {
		cli.PrintDefaults()
		os.Exit(0)
	} else {
		conf, err = config.LoadConfig(confFile)
		if err != nil && confFile != config.DefaultConfigFilePath() {
			log.Println(err)
		}
	}

	// Configure
	if address != defaults.Address {
		conf.Address = address
	}
	if port != defaults.Port {
		conf.Port = port
	}
	if peerList != "" {
		peers := strings.Split(peerList, ",")
		conf.BootstrapPeers = append(conf.BootstrapPeers, peers...)
	}
	return conf
}
