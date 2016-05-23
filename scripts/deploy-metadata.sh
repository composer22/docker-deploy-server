#!/usr/bin/env bash

# Deploy the metadata repo from git to all machines.
#
# ENV_TAG - wildcard to find machines that are in a particular environment.
# GIT_REPO - name of the git repo that holds the metadata.
# METADATA_MOUNT - remote machines tarket directory where to push it.
# TEMP_PATH - local temporary directory where the metadata is stored now.
#
export ENV_TAG=$1
export GIT_REPO=$2
export METADATA_MOUNT=$3
export TEMP_DIRECTORY=$4

docker-machine ls --filter name=${ENV_TAG} --format "{{.Name}}" | \
   xargs  -I machinename docker-machine scp --recursive \
            ${TEMP_DIRECTORY}/${GIT_REPO} \
             machinename:${METADATA_MOUNT}/
