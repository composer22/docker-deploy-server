#!/usr/bin/env bash

set -e

# Download a Docker image from the repo and extract out any metadata for the launch.
#
# DOCKER_IMAGE_TAG - Docker image tag to download.
# DOCKER_REGISTRY - Docker registry.
# DOCKER_REPO - Docker application repository.
# TEMP_DIRECTORY - local temporary directory where we should extract metadata.
#
export DOCKER_IMAGE_TAG=$1
export DOCKER_REGISTRY=$2
export DOCKER_REPO=$3
export TEMP_DIRECTORY=$4

export DOCKER_IMAGE_FULLNAME="${DOCKER_REGISTRY}/${DOCKER_REPO}:${DOCKER_IMAGE_TAG}"

docker rmi ${DOCKER_IMAGE_FULLNAME} 2> /dev/null || echo > /dev/null
docker pull ${DOCKER_IMAGE_FULLNAME}

export cid=$(docker create ${DOCKER_IMAGE_FULLNAME})

cd ${TEMP_DIRECTORY}
docker cp $cid:/etc/docker/metadata/ - > docker-metadata-$cid.tar
docker rm -v $cid
mkdir docker-metadata-${cid}
tar -xf docker-metadata-${cid}.tar -C docker-metadata-${cid} --strip-components=1
mv docker-metadata-${cid}/docker-compose.yml ${TEMP_DIRECTORY}
rm -rf docker-metadata-${cid}
rm -rf docker-metadata-${cid}.tar
docker rmi ${DOCKER_IMAGE_FULLNAME}
unset cid
#
exit 0
