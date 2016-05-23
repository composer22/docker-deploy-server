### Working scripts

A number of shell scripts to call from the application for docker,
docker-machine, docker-swarm and git.

* deploy-containers.sh - removes old services on the machines and starts up a new service with scaling.
* deploy-metadata.sh - SSH deploys git metadata to volumes on all machines.
* download-image.sh - downloads an images to the local machine to verify it exists and to extract metadata.
* download-metadata.sh - downloads metadata from the git repository on github.
