package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	flag "github.com/ogier/pflag"

	"github.com/aprice/observatory"
	"github.com/aprice/observatory/checks"
	"github.com/aprice/observatory/utils"
)

func main() {
	var (
		err         error
		name        string
		coordinator string
		roles       string
		help        bool
		version     bool
	)
	cli := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cli.StringVarP(&name, "name", "n", "", "Subject's unique name (default hostname)")
	cli.StringVarP(&coordinator, "coordinator", "c", "localhost:13100", "Address of a Coordinator node")
	cli.StringVarP(&roles, "roles", "r", "default", "Comma-separated list of roles to bootstrap to")
	cli.BoolVarP(&help, "help", "h", false, "Print usage information")
	cli.BoolVarP(&version, "version", "v", false, "Print version information and exit")
	cli.Parse(os.Args[1:])
	if help {
		cli.Usage()
		os.Exit(0)
	}
	if version {
		fmt.Println(observatory.VersionInfo())
		os.Exit(0)
	}
	if name == "" {
		name, err = os.Hostname()
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(observatory.VersionInfo())

	// Start up reconfigure goroutine
	log.Println("Starting reconfig routine")
	quit := make(utils.SentinelChannel)
	go Reconfigure(name, coordinator, roles, quit)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	s := <-c
	log.Println("Got signal: ", s)
	log.Println("Stopping configuration loop.")
	quit <- utils.Nothing
	log.Println("Stopping checks.")
	checks.StopAllChecks()
}
