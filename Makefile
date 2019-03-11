REV:=$(shell git rev-parse HEAD)
DOCKER_IMAGE:=parkr/docker-kill-unhealthy-containers:$(REV)

default: test

build:
	docker build -t $(DOCKER_IMAGE) .

dive: build
	dive $(DOCKER_IMAGE)

test: build run-unhealthy
	docker run --rm \
		--name docker-reaper-test \
		-v /var/run/docker.sock:/var/run/docker.sock \
		$(DOCKER_IMAGE)

publish: build
	docker push $(DOCKER_IMAGE)

run-unhealthy: build-unhealthy
	docker run -d --rm \
		--name unhealthy \
		unhealthy \
		sleep 6000

build-unhealthy:
	docker build -t unhealthy unhealthy
