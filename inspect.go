package dockerkillunhealthycontainers

import "errors"

type ContainerInfo struct {
}

func InspectContainer(containerId string) (ContainerInfo, error) {
	return ContainerInfo{}, errors.New("not implemented yet")
}
