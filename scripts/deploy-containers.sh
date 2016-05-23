#!/usr/bin/env bash

# Deploy the metadata repo from git to all machines.
#
# DOCKER_IMAGE_NAME - name of the docker image = repository.
# DOCKER_IMAGE_TAG - tag of the image to deploy ex: latest
# DOCKER_LAST_IMAGE_TAG - last tag for this image that was deployed.
# DOCKER_REGISTRY - docker registry where images are kept.
# DOCKER_SERVICE_NAME - service name to scale ex; foo-bar image name => foo_bar service name.
# MACHINE - the machine or master (in swarm) to set the environment to, as delivered form docker-machine.
# NUM_CONTAINERS - lnumber of containers to scale.
# PROJECT - a project name for all containers in this environment.
# SWARM - if set, then additional param of --swarm added to docker-machine env.
# TEMP_DIRECTORY - temp diirectory where docker-compose.yml is kept.
#
export DOCKER_IMAGE_NAME=$1
export DOCKER_IMAGE_TAG=$2
export DOCKER_LAST_IMAGE_TAG=$3
export DOCKER_REGISTRY=$4
export DOCKER_SERVICE_NAME=$5
export MACHINE=$6
export NUM_CONTAINERS=$7
export PROJECT=$8
export SWARM=$9
export TEMP_DIRECTORY="${10}"

export DOCKER_REPO_NAME="${DOCKER_IMAGE_NAME}"

export swarm_sw=""
if [ "$SWARM" == "true" ]
then
  swarm_sw="--swarm"
fi

# set working directory.
cd "${TEMP_DIRECTORY}"

eval $(docker-machine env ${swarm_sw} ${MACHINE})

# Destroy old environment if its not the same version tag (1.0.0-31 vs 1.0.0-32).
if [ "${DOCKER_LAST_IMAGE_TAG}" != "" ] && [ "${DOCKER_IMAGE_TAG}" != "${DOCKER_LAST_IMAGE_TAG}" ]
then
  docker-compose -f docker-compose.yml -p ${PROJECT} stop ${DOCKER_SERVICE_NAME}
  docker-compose -f docker-compose.yml -p ${PROJECT} rm -v --force ${DOCKER_SERVICE_NAME}
  docker rmi ${DOCKER_REGISTRY}/${DOCKER_IMAGE_NAME}:${DOCKER_LAST_IMAGE_TAG}
else
  docker-compose -f docker-compose.yml -p ${PROJECT} \
     pull ${DOCKER_REGISTRY}/${DOCKER_IMAGE_NAME}:${DOCKER_LAST_IMAGE_TAG} \
	 2> /dev/null || echo > /dev/null
fi

# Launch and scale.
docker-compose -f docker-compose.yml -p ${PROJECT} up -d --force-recreate --no-build ${DOCKER_SERVICE_NAME} \
  2> /dev/null || echo > /dev/null
docker-compose -f docker-compose.yml -p ${PROJECT} scale ${DOCKER_SERVICE_NAME}=${NUM_CONTAINERS}
