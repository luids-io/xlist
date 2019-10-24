// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"

	"github.com/luids-io/xlist/cmd/xlget/config"
)

//Variables for version output
var (
	Program  = "xlget"
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
	auto       = false
)

func init() {
	//config mapped params
	cfg.PFlags()
	//behaviour params
	pflag.StringVar(&configFile, "config", configFile, "Use explicit config file.")
	pflag.BoolVar(&version, "version", version, "Show version.")
	pflag.BoolVarP(&help, "help", "h", help, "Show this help.")
	pflag.BoolVar(&debug, "debug", debug, "Enable debug.")
	pflag.BoolVar(&dryRun, "dry-run", dryRun, "Check configuration and updates.")
	pflag.BoolVar(&auto, "auto", auto, "Run in auto mode.")
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

	mgr, err := createManager(logger)
	if err != nil {
		logger.Fatalf("couldn't create xlget manager: %v", err)
	}

	// dryrun mode
	if dryRun {
		updates := mgr.NeedsUpdate()
		fmt.Println("configuration ok")
		if len(updates) == 0 {
			fmt.Println("no entry needs to be updated")
			os.Exit(0)
		}
		summary := "needs update:"
		for _, e := range updates {
			summary = fmt.Sprintf("%s '%s'", summary, e)
		}
		fmt.Println(summary)
		os.Exit(0)
	}

	// auto mode
	if auto {
		//launch goroutine for shutdown
		close := make(chan bool, 1)
		go func() {
			sigint := make(chan os.Signal, 1)
			signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
			<-sigint
			logger.Debugf("received interrupt signal, stopping")
			close <- true
		}()

	PROCESSLOOP:
		for {
			stop, done, err := mgr.Update()
			if err != nil {
				logger.Fatalf("updating entries: %v", err)
			}
			defer stop()

			select {
			case <-done:
			case <-close:
				stop()
				break PROCESSLOOP
			}

			timer := time.NewTimer(time.Minute)
			select {
			case <-close:
				break PROCESSLOOP
			case <-timer.C:
				logger.Debugf("tick")
			}
		}
		logger.Infof("%s finished", Program)
		os.Exit(0)
	}

	// manual mode
	stop, done, err := mgr.Update()
	if err != nil {
		logger.Fatalf("updating entries: %v", err)
	}
	defer stop()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigint:
		stop()
	case <-done:
	}

	logger.Infof("%s finished", Program)
}
