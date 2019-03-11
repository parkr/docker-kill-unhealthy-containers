package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var observeInterval = 10 * time.Second
var checkTimeout = 5 * time.Second
var stopTimeout = 5 * time.Second
var failingStreakThreshold = 5

func main() {
	log.Println("booting...")

	flag.Parse()

	dockerd, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalf("error creating docker client: %+v", err)
	}

	observe(dockerd)
}

// observe runs a continuous loop reading running containers for health checks
// and stops unhealthy containers.
func observe(dockerd *client.Client) {
	ticker := time.NewTicker(observeInterval)
	for t := range ticker.C {
		log.Println("Tick at", t)
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
			defer cancel()
			containers, err := dockerd.ContainerList(ctx, types.ContainerListOptions{})
			if err != nil {
				log.Fatalf("error listing containers: %+v", err)
			}
			for _, container := range containers {
				go checkContainer(dockerd, container.ID)
			}
		}()
	}
}

func checkContainer(dockerd *client.Client, containerId string) {
	logger := log.New(os.Stdout, "["+containerId[:10]+"]", log.LstdFlags)
	logger.Println("checking...")

	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	container, err := dockerd.ContainerInspect(ctx, containerId)
	if err != nil {
		logger.Printf("error inspecting: %#v", err)
		return
	}

	if container.State == nil {
		logger.Println("no state")
		return
	}

	if container.State.Health == nil {
		logger.Println("no health")
		return
	}

	if container.State.Health.Status == "healthy" {
		logger.Println("healthy")
		return
	}

	if container.State.Health.FailingStreak < failingStreakThreshold {
		logger.Printf("unhealthy but failing streak %d, less than threshold of %d",
			container.State.Health.FailingStreak,
			failingStreakThreshold)
	}

	err = dockerd.ContainerStop(ctx, containerId, &stopTimeout)
	if err != nil {
		log.Printf("")
	}
	log.Printf("stopped")
}
