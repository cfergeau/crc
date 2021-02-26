package cmd

import (
	"fmt"
	"net"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/gvisor-tap-vsock/pkg/transport"
)

func vsockListener() (net.Listener, error) {
	return transport.Listen(transport.DefaultURL)
}

func httpListener() (net.Listener, error) {
	return transport.Listen(fmt.Sprintf("unix://%s", constants.DaemonHTTPSocketPath))
}
