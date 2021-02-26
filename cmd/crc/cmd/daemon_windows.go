package cmd

import (
	"net"

	"github.com/code-ready/gvisor-tap-vsock/pkg/transport"
)

func vsockListener() (net.Listener, error) {
	return transport.Listen(transport.DefaultURL)
}

func httpListener() (net.Listener, error) {
	return nil, nil
}
