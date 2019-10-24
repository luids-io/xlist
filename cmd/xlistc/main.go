// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/cmd/xlistc/config"
)

//Variables for version output
var (
	Program  = "xlistc"
	Build    = "unknown"
	Version  = "unknown"
	Revision = "unknown"
)

var (
	cfg = config.Default(Program)
	//behaviour
	configFile = ""
	version    = false
	debug      = false
	help       = false
	//input
	inStdin = false
	inFile  = ""
)

func init() {
	//config mapped params
	cfg.PFlags()
	//behaviour params
	pflag.StringVar(&configFile, "config", configFile, "Use explicit config file.")
	pflag.BoolVar(&version, "version", version, "Show version.")
	pflag.BoolVarP(&help, "help", "h", help, "Show this help.")
	pflag.BoolVar(&debug, "debug", debug, "Enable debug.")
	//input params
	pflag.BoolVar(&inStdin, "stdin", inStdin, "From stdin.")
	pflag.StringVarP(&inFile, "file", "f", inFile, "File for input.")
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

	// creates logger
	logger, err := createLogger(debug)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	// create grpc client
	client, err := createClient(logger)
	if err != nil {
		logger.Fatalf("couldn't create client: %v", err)
	}
	defer client.Close()

	if len(pflag.Args()) == 0 && !inStdin && inFile == "" {
		logger.Debugf("testing service")
		startp := time.Now()
		err = client.Ping()
		if err != nil {
			logger.Fatalf("test failed: %v", err)
		}
		resources := client.Resources()
		fmt.Fprintf(os.Stdout, "%v (%v)\n", resources, time.Since(startp))
		return
	}

	remoteResources := client.Resources()
	//reads from args
	if !inStdin && inFile == "" {
		for _, arg := range pflag.Args() {
			t, err := xlist.ResourceType(arg, remoteResources)
			if err != nil {
				logger.Fatalf("invalid name '%s': %v", arg, err)
			}
			startc := time.Now()
			r, err := client.Check(context.Background(), arg, t)
			if err != nil {
				logger.Fatalf("check '%s' returned error: %v", arg, err)
			}
			fmt.Fprintf(os.Stdout, "%s,%s: %v,\"%s\",%v (%v)\n", t, arg, r.Result, r.Reason, r.TTL, time.Since(startc))
		}
		return
	}

	// read from file or stdin
	reader := os.Stdin
	if inFile != "" {
		file, err := os.Open(inFile)
		if err != nil {
			logger.Fatalf("opening file: %v", err)
		}
		defer file.Close()
		reader = file
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		arg := fields[0]
		t, err := xlist.ResourceType(arg, remoteResources)
		if err != nil {
			logger.Fatalf("invalid name '%s': %v", arg, err)
			continue
		}
		startc := time.Now()
		r, err := client.Check(context.Background(), arg, t)
		if err != nil {
			logger.Fatalf("check '%s' returned error: %v", arg, err)
			continue
		}
		fmt.Fprintf(os.Stdout, "%s,%s: %v,\"%s\",%v (%v)\n", t, arg, r.Result, r.Reason, r.TTL, time.Since(startc))
	}
	if err := scanner.Err(); err != nil {
		logger.Errorf("reading: %v", err)
	}
}
