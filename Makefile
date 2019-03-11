REV:=$(shell git rev-parse HEAD)
DOCKER_IMAGE:=parkr/docker-kill-unhealthy-containers:$(REV)

default: test

build:
	docker build -t $(DOCKER_IMAGE) .

dive: build
	dive $(DOCKER_IMAGE)

test: build
	docker run --rm -v /var/run/docker.sock:/var/run/docker.sock $(DOCKER_IMAGE)

publish: build
	docker push $(DOCKER_IMAGE)