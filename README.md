# docker-deploy-server
[![License MIT](https://img.shields.io/npm/l/express.svg)](http://opensource.org/licenses/MIT)
[![Build Status](https://travis-ci.org/composer22/docker-deploy-server.svg?branch=master)](http://travis-ci.org/composer22/docker-deploy-server)
[![Current Release](https://img.shields.io/badge/release-v0.0.1-brightgreen.svg)](https://github.com/composer22/docker-deploy-server/releases/tag/v0.0.1)
[![Coverage Status](https://coveralls.io/repos/composer22/docker-deploy-server/badge.svg?branch=master)](https://coveralls.io/r/composer22/docker-deploy-server?branch=master)

An API server to allow deployment of Docker containers and metadata to a Docker machine or cluster written in [Golang.](http://golang.org)

## About

This is an API server that provides an endpoint to launch Docker containers on one or more machines.
It should be running on a "control center server", a server that is setup to access machines across your environments.

API calls are gated via API tokens passed by the restful requests.

## Requirements

Running on the control machine, the following should be installed and readily available:

Docker Utilities needed:

* docker-machine: command line utility for accessing and modifying nodes in the clusters.
* docker: daemon that is running for extracing and testing images before deploying.
* docker-compose: command line utility for launching and controlling containers across machines.

Additional Resources Needed:

* MySQL database - for storing accounts (API tokens and their rights), logs and other fixed configuration info.
* Redis - for storing queues and last deploy information for the environments.
* git - git client, github authorization information and a github private repository where application metadata is kept.

Assumptions:

* The docker-compose.yml file is stored in the docker image under the internal volume: /etc/docker/metadata/
* The metadata for all the application container launches is stored in a single repo under github.

For the DB schema, please see ./db/schema.sql
For an example repo directory structure, see examples.

## Command Line Usage

```
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

```
See the examples folder for configuration file and repo examples.

## HTTP API

Header for services other than /health should contain:

* Accept: application/json
* Authorization: Bearer with token
* Content-Type: application/json

Example cURL:

```
$ curl -i -H "Accept: application/json" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer S0M3B3EARERTOK3N" \
-X GET "http://0.0.0.0:8080/v1.0/info"

HTTP/1.1 200 OK
Content-Type: application/json;charset=utf-8
Date: Fri, 03 Apr 2015 17:29:17 +0000
Server: sf-deploy-service
X-Request-Id: DC8D9C2E-8161-4FC0-937F-4CA7037970D5
Content-Length: 0
```

Three API routes are provided for service measurement:

* http://localhost:8080/v1.0/health - GET: Is the server alive?
* http://localhost:8080/v1.0/info - GET: What are the params of the server?
* http://localhost:8080/v1.0/metrics - GET: What are the performance statistics of the server?


These routes handle and service deploy requests:

* http://localhost:8080/v1.0/deploy - POST: Make a request to deploy an image to an environment.
* http://localhost:8080/v1.0/status/:deployID - GET: Return the status of a previous deploy request.

The following is an example of a call for the route _deployment_:
```
POST http://localhost:8080/v1.0/deploy

curl -i -H "Accept: application/json" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer S0M3B3EARERTOK3N" \
-X POST "http://0.0.0.0:8080/v1.0/deploy" \
-d "<json payload see below>"
```
json payload:
```
{
  "environment":"dev",
  "imageName":"hello-world",
  "imageTag":"1.0.0-32"
}
```
Mandatory Parameters:
* environment: which environment you want to deploy to? Matches what is in the config (ex: dev, qa, stage, prod).
* imageName: This is the name of the application in the Docker repository assigned to the environment.
* imageTag: the tag assiged to the image. This defaults to "latest".

This API request will return a UUID for the deploy:
```
{
    "deployID": "051A9069-0E3A-41EC-9C98-E6D29E91FBB3"
}
```
...that can be used to check the deploy status at any time:
```
curl -i -H "Accept: application/json" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer S0M3B3EARERTOK3N" \
-X GET "http://0.0.0.0:8080/v1.0/status/051A9069-0E3A-41EC-9C98-E6D29E91FBB3"
```
which returns:
```
{
    "createdAt": "2015-08-27 18:58:16",
    "deployID": "051A9069-0E3A-41EC-9C98-E6D29E91FBB3",
    "environment":"dev",
    "imageName":"hello-world",
    "imageTag":"1.0.0-32",
    "log": "blabla...\nSUCCESS: Service deployed successfully.\n",
    "message": "Service deployed successfully.",
    "status": 2,
    "updatedAt": "2015-08-27 18:58:30"
}
```

## Building

This code currently requires version 1.6.2 or higher of Go.

Information on Golang installation, including pre-built binaries, is available at <http://golang.org/doc/install>.

Run `go version` to see the version of Go which you have installed.

Run `go build` inside the directory to build.

Run `go test ./...` to run the unit regression tests.

A successful build run produces no messages and creates an executable called `docker-deploy-server` in this
directory.

Run `go help` for more guidance, and visit <http://golang.org/> for tutorials, presentations, references and more.

## Docker Images

A prebuilt docker image is available at (http://www.docker.com) [docker-deploy-server](https://registry.hub.docker.com/u/composer22/docker-deploy-server/)

If you have docker installed, run:
```
docker pull composer22/docker-deploy-server:latest

or

docker pull composer22/docker-deploy-server:<version>

if available.
```
See /docker directory README for more information on how to run the container.

## License

(The MIT License)

Copyright (c) 2015 Pyxxel Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to
deal in the Software without restriction, including without limitation the
rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
IN THE SOFTWARE.
