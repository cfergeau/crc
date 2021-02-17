package adminhelper

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/Microsoft/go-winio"
	"github.com/code-ready/admin-helper/pkg/client"
	"github.com/code-ready/crc/pkg/os/windows/powershell"
)

func execute(args ...string) error {
	_, _, err := powershell.ExecuteAsAdmin("modifying hosts file", strings.Join(append([]string{adminHelperPath}, args...), " "))
	return err
}

func instance() helper {
	return client.New(&http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return winio.DialPipeContext(ctx, `\\.\pipe\crc-admin-helper`)
			},
		},
	}, "http://unix")
}
