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

const defaultObserveInterval = "30s"
const defaultCheckTimeout = "25s"

var observeInterval time.Duration
var checkTimeout time.Duration
var failingStreakThreshold = 5
var maxStopRetries = 5

func main() {
	var observeIntervalStr string
	flag.StringVar(&observeIntervalStr, "interval", defaultObserveInterval, "Interval to check containers")
	var checkTimeoutStr string
	flag.StringVar(&checkTimeoutStr, "timeout", defaultCheckTimeout, "Maximum duration of the total reap operation")
	var dryRun bool
	flag.BoolVar(&dryRun, "dry-run", false, "Observe health but don't act")
	flag.Parse()

	logPrefix := "[main]"

	var err error

	observeInterval, err = time.ParseDuration(observeIntervalStr)
	if err != nil {
		log.Fatalf("%s unable to parse interval %q: %+v", logPrefix, observeIntervalStr, err)
	}

	checkTimeout, err = time.ParseDuration(checkTimeoutStr)
	if err != nil {
		log.Fatalf("%s unable to parse timeout duration %q: %+v", logPrefix, checkTimeoutStr, err)
	}

	dockerd, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalf("%s error creating docker client: %+v", logPrefix, err)
	}

	func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err := dockerd.Ping(ctx)
		if err != nil {
			log.Fatalf("%s error pinging docker daemon: %+v", logPrefix, err)
		}
	}()

	log.Println(logPrefix, "observing every", observeInterval, "and check timeout is", checkTimeout)

	observe(dockerd, dryRun)
}

// observe runs a continuous loop reading running containers for health checks
// and stops unhealthy containers.
func observe(dockerd *client.Client, dryRun bool) {
	ticker := time.NewTicker(observeInterval)
	for range ticker.C {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
			defer cancel()
			containers, err := dockerd.ContainerList(ctx, types.ContainerListOptions{})
			if err != nil {
				log.Fatalf("error listing containers: %+v", err)
			}
			for _, container := range containers {
				container := container
				if container.State == "running" {
					var name string
					if len(container.Names) > 0 {
						name = container.Names[0]
					}
					go checkContainerWithTimeout(dockerd, container.ID, name, dryRun)
				}
			}
		}()
	}
}

func checkContainerWithTimeout(dockerd *client.Client, containerId, containerName string, dryRun bool) {
	logPrefix := "[" + containerId[:10] + " " + containerName + "]"
	logger := log.New(os.Stdout, "", log.LstdFlags)

	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	select {
	case <-checkContainer(ctx, dockerd, logger, logPrefix, containerId, dryRun):
	case <-ctx.Done():
		log.Println(logPrefix, "context exceeded checking... maybe my timeouts need tweaking")
	}
}

func checkContainer(ctx context.Context, dockerd *client.Client, logger *log.Logger, logPrefix, containerId string, dryRun bool) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		start := time.Now()
		container, err := dockerd.ContainerInspect(ctx, containerId)
		if err != nil {
			logger.Printf("%s error inspecting: %#v", logPrefix, err)
			done <- struct{}{}
			return
		}

		if container.State == nil {
			logger.Println(logPrefix, "skipping, no state data")
			done <- struct{}{}
			return
		}

		if container.State.Health == nil {
			logger.Println(logPrefix, "skipping, no health data")
			done <- struct{}{}
			return
		}

		if container.State.Health.Status == "healthy" {
			logger.Println(logPrefix, "skipping, healthy")
			done <- struct{}{}
			return
		}

		if container.State.Health.FailingStreak < failingStreakThreshold {
			logger.Printf("%s unhealthy but failing streak %d, less than threshold of %d",
				logPrefix,
				container.State.Health.FailingStreak,
				failingStreakThreshold)
			done <- struct{}{}
			return
		}
		
		if dryRun {
			logger.Printf("%s unhealthy but skipping since we dry running it", logPrefix)
			done <- struct{}{}
			return
		}

		for i := 0; i < maxStopRetries; i++ {
			err = dockerd.ContainerStop(ctx, containerId, nil)
			if err == nil {
				break
			}
		}
		if err != nil {
			log.Printf("%s unable to stop container after %d retries: %v %#v",
				logPrefix, maxStopRetries, err, err)
			done <- struct{}{}
			return
		}

		log.Printf("%s stopped, %s elapsed", logPrefix, time.Since(start))
		done <- struct{}{}
	}()
	return done
}
