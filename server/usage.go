package server

import (
	"fmt"
	"os"
)

const usageText = `
Description: docker-deploy-server is a server for deploying services to one or more
Docker nodes. It can be used for single nodes or across a Docker Swarm cluster.

Usage: docker-deploy-server [options...]

Server options:
    -p, --config-path PATH           PATH to the config.yml file (default: "~/.docker-deploy-server").
    -x, --config-prefix PREFIX       PREFIX of the <config>.yml (default: "docker-deploy-server").

    -d, --debug                      Enable debugging output (default: false)

Common options:
    -h, --help                       Show this message
    -V, --version                    Show version

Example:

	# Config file name: ~/.docker-deploy-service/sf-deploy-service.yml

    docker-deploy-server \
	  --config-path " ~/.docker-deploy-service" \
	  --config-prefix "sf-deploy-service"

`

// PrintUsageAndExit is used to print out command line options.
func PrintUsageAndExit() {
	fmt.Printf("%s\n", usageText)
	os.Exit(0)
}
