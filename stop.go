package dockerkillunhealthycontainers

import (
	"errors"
	"net"
)

func StopContainer(dockerConn *net.Conn, containerId string) error {
	return errors.New("not implemented yet")
}
