// docker-deploy-server is a server for deploying docker containers into a set of Docker machines.
// The machines can also be a group of machines in a Docker Swarm cluster.
package main

import (
	"os"
	"strings"

	"github.com/composer22/docker-deploy-server/logger"
	"github.com/composer22/docker-deploy-server/server"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	log *logger.Logger
)

func init() {
	log = logger.New(logger.UseDefault, false)
}

// main entry point for the application or server launch.
func main() {
	var showVersion bool
	opt := &server.Options{}
	flag.StringVarP(&opt.ConfigPath, "config-path", "p", "", "Path to the config file ex: /path/to/dir")
	flag.StringVarP(&opt.ConfigPrefix, "config-prefix", "x", server.DefaultConfigPrefix, "Config file prefix: ex. <prefix>.yml.")
	flag.BoolVarP(&opt.Debug, "debug", "d", false, "Enable debugging output.")
	flag.BoolVarP(&showVersion, "version", "V", false, "Show version.")
	flag.Usage = server.PrintUsageAndExit
	flag.Parse()

	// Version flag request?
	if showVersion {
		server.PrintVersionAndExit()
	}

	// Check additional commands beyond the flags.
	for _, arg := range flag.Args() {
		switch strings.ToLower(arg) {
		case "version":
			server.PrintVersionAndExit()
		case "help":
			server.PrintUsageAndExit()
		}
	}

	// Read config file into server options.
	v := viper.New()
	opt.SetConfigDefaults(v)
	if err := v.ReadInConfig(); err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}
	opt.FillConfig(v)

	// Boot the server.
	s := server.New(opt, log)
	if err := s.Start(); err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
