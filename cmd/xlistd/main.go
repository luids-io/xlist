// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/luids-io/core/utils/serverd"
	"github.com/luids-io/xlist/cmd/xlistd/config"
)

//Variables for version output
var (
	Program  = "xlistd"
	Build    = "unknown"
	Version  = "unknown"
	Revision = "unknown"
)

var (
	cfg = config.Default(Program)
	//behaviour
	configFile = ""
	version    = false
	help       = false
	debug      = false
	dryRun     = false
)

func init() {
	//config mapped params
	cfg.PFlags()
	//behaviour params
	pflag.StringVar(&configFile, "config", configFile, "Use explicit config file.")
	pflag.BoolVar(&version, "version", version, "Show version.")
	pflag.BoolVarP(&help, "help", "h", help, "Show this help.")
	pflag.BoolVar(&debug, "debug", debug, "Enable debug.")
	pflag.BoolVar(&dryRun, "dry-run", dryRun, "Checks and construct list but not start service.")
	pflag.Parse()
}

func main() {
	if version {
		fmt.Printf("version: %s\nrevision: %s\nbuild: %s\n", Version, Revision, Build)
		os.Exit(0)
	}
	if help {
		pflag.Usage()
		os.Exit(0)
	}
	// load configuration
	err := cfg.LoadIfFile(configFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	//creates logger
	logger, err := createLogger(debug)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// echo version and config
	logger.Infof("%s (version: %s build: %s)", Program, Version, Build)
	if debug {
		logger.Debugf("configuration dump:\n%v", cfg.Dump())
	}

	// creates main server manager
	msrv := serverd.New(serverd.SetLogger(logger))

	// create api services and register
	apisvc, err := createAPIServices(msrv, logger)
	if err != nil {
		logger.Fatalf("couldn't create api registry: %v", err)
	}

	//setup event notifier
	err = setupEventNotify(apisvc, msrv, logger)
	if err != nil {
		logger.Fatalf("couldn't create event notify: %v", err)
	}

	// create lists
	lists, err := createLists(apisvc, msrv, logger)
	if err != nil {
		logger.Fatalf("couldn't create lists: %v", err)
	}

	if dryRun {
		fmt.Println("configuration seems ok")
		os.Exit(0)
	}

	// create grpc check server
	gsrv, err := createServer(msrv)
	if err != nil {
		logger.Fatalf("couldn't create check server: %v", err)
	}

	// create grpc service
	err = createCheckAPI(gsrv, lists, msrv, logger)
	if err != nil {
		logger.Fatalf("couldn't create check api: %v", err)
	}

	// creates health server
	err = createHealthSrv(msrv, logger)
	if err != nil {
		logger.Fatalf("creating health server: %v", err)
	}

	//run server
	err = msrv.Run()
	if err != nil {
		logger.Errorf("running server: %v", err)
	}
	logger.Infof("%s finished", Program)
}
