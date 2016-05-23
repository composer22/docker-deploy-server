#!/usr/bin/env bash

set -e

# Download the metadata repo from git
#
# GIT_REPO - githum repo for metadata.
# GIT_ROOT - github prefix for clone command via .ssh/config etc.
# TEMP_PATH - directory to download repo on local.
#
export GIT_REPO=$1
export GIT_ROOT=$2
export TEMP_DIRECTORY=$3

rm -rf ${TEMP_DIRECTORY}/${GIT_REPO} 2> /dev/null || echo > /dev/null
git clone -v ${GIT_ROOT}/${GIT_REPO}.git ${TEMP_DIRECTORY}/${GIT_REPO} 2> /dev/null || echo > /dev/null

rm -rf ${TEMP_DIRECTORY}/${GIT_REPO}/.git 2> /dev/null || echo > /dev/null
