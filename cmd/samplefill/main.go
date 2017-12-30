package main

import (
	"fmt"
	"log"
	"os"

	flag "github.com/ogier/pflag"

	"github.com/aprice/observatory"
	"github.com/aprice/observatory/database"
	"github.com/aprice/observatory/server/config"
)

func main() {
	conf := buildConfig()
	fmt.Println(observatory.VersionInfo())
	conf.Init()
	defer conf.ContextFactory.Close()

	ctx, err := conf.ContextFactory.Get()
	if err != nil {
		log.Fatal(err)
	}

	database.FillSampleData(ctx)
	ctx.Close()
}

func buildConfig() config.Configuration {
	var (
		confFile string
		help     bool
		version  bool
		conf     config.Configuration
		err      error
	)

	cli := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cli.StringVarP(&confFile, "config", "c", config.DefaultConfigFilePath(), "Configuration file path")
	cli.BoolVarP(&help, "help", "h", false, "Print usage information")
	cli.BoolVarP(&version, "version", "v", false, "Print version information and exit")
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

	return conf
}
