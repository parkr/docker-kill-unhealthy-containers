# docker-kill-unhealthy-containers

Monitor Docker containers and kill the unhealthy ones.

## Installation

```console
$ go get github.com/parkr/docker-kill-unhealthy-containers/...
$ docker-kill-unhealthy-containers -h
```

## Usage

```console
$ docker-kill-unhealthy-containers
```

This process will monitor all running docker containers which report a health check and kill them if they fail 5 times in a row.

## License

MIT. See LICENSE.
