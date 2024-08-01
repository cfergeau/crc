package machine

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"runtime"
	"strconv"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/daemonclient"
	crcErrors "github.com/crc-org/crc/v2/pkg/crc/errors"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	crcPreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	crcssh "github.com/crc-org/crc/v2/pkg/crc/ssh"
	"github.com/pkg/errors"
)

func exposePorts(sshRunner *crcssh.Runner, preset crcPreset.Preset, ingressHTTPPort, ingressHTTPSPort uint) error {
	portsToExpose := vsockPorts(preset, ingressHTTPPort, ingressHTTPSPort)
	logging.Infof("ports to expose: %v", portsToExpose)
	alreadyOpenedPorts, err := list(sshRunner)
	if err != nil {
		logging.Infof("listOpenPorts: %v", err)
		return err
	}
	logging.Infof("already opened ports: %v", portsToExpose)
	var missingPorts []types.ExposeRequest
	for _, port := range portsToExpose {
		if !isOpened(alreadyOpenedPorts, port) {
			missingPorts = append(missingPorts, port)
		}
	}
	logging.Infof("missing ports: %v", missingPorts)
	for i := range missingPorts {
		port := &missingPorts[i]
		req, err := json.Marshal(port)
		if err != nil {
			return errors.Wrapf(err, "failed to expose port %s -> %s", port.Local, port.Remote)
		}
		stdout, stderr, err := sshRunner.Run("curl", "-X", "POST", "-d", fmt.Sprintf("'%s'", string(req)), "http://192.168.127.1/services/forwarder/expose")
		logging.Infof("exposePorts stdout: %s\nstderr: %s\nerr: %v", stdout, stderr, err)
		if err != nil {
			return errors.Wrapf(err, "failed to expose port %s -> %s", port.Local, port.Remote)
		}

		/*
			if err := daemonClient.NetworkClient.Expose(port); err != nil {
				return errors.Wrapf(err, "failed to expose port %s -> %s", port.Local, port.Remote)
			}
		*/
	}
	return nil
}

func isOpened(exposed []types.ExposeRequest, port types.ExposeRequest) bool {
	logging.Infof("testing if %+v is already open", port)
	if port.Local == "127.0.0.1:2222" {
		return true
	}
	for _, alreadyOpenedPort := range exposed {
		if port == alreadyOpenedPort {
			return true
		}
	}
	return false
}

func unexposePorts() error {
	var mErr crcErrors.MultiError
	daemonClient := daemonclient.New()
	alreadyOpenedPorts, err := listOpenPorts(daemonClient)
	if err != nil {
		return err
	}
	for _, port := range alreadyOpenedPorts {
		if err := daemonClient.NetworkClient.Unexpose(&types.UnexposeRequest{Protocol: port.Protocol, Local: port.Local}); err != nil {
			mErr.Collect(errors.Wrapf(err, "failed to unexpose port %s ", port.Local))
		}
	}
	if len(mErr.Errors) == 0 {
		return nil
	}
	return mErr
}

func listOpenPorts(daemonClient *daemonclient.Client) ([]types.ExposeRequest, error) {
	alreadyOpenedPorts, err := daemonClient.NetworkClient.List()
	if err != nil {
		logging.Error("Is 'crc daemon' running? Network mode 'vsock' requires 'crc daemon' to be running, run it manually on different terminal/tab")
		return nil, err
	}
	return alreadyOpenedPorts, nil
}

func list(sshRunner *crcssh.Runner) ([]types.ExposeRequest, error) {
	stdout, stderr, err := sshRunner.Run("curl", "http://192.168.127.1/services/forwarder/all")
	logging.Infof("listPorts stdout: %s\nstderr: %s\nerr: %v", stdout, stderr, err)
	if err != nil {
		return nil, err
	}
	var ports []types.ExposeRequest
	err = json.Unmarshal([]byte(stdout), &ports)
	if err != nil {
		return nil, err
	}
	return ports, nil
}

const (
	virtualMachineIP = "192.168.127.2"
	hostVirtualIP    = "192.168.127.254"
	internalSSHPort  = "22"
	remoteHTTPPort   = "80"
	remoteHTTPSPort  = "443"
	apiPort          = "6443"
	cockpitPort      = "9090"
)

func vsockPorts(preset crcPreset.Preset, ingressHTTPPort, ingressHTTPSPort uint) []types.ExposeRequest {
	socketProtocol := types.UNIX
	socketLocal := constants.GetHostDockerSocketPath()
	if runtime.GOOS == "windows" {
		socketProtocol = types.NPIPE
		socketLocal = constants.DefaultPodmanNamedPipe
	}
	exposeRequest := []types.ExposeRequest{
		{
			Protocol: "tcp",
			Local:    net.JoinHostPort(constants.LocalIP, strconv.Itoa(2223)),
			Remote:   net.JoinHostPort(virtualMachineIP, internalSSHPort),
		},
		{
			Protocol: socketProtocol,
			Local:    socketLocal,
			Remote:   getSSHTunnelURI(),
		},
	}

	switch preset {
	case crcPreset.OpenShift, crcPreset.OKD, crcPreset.Microshift:
		exposeRequest = append(exposeRequest,
			types.ExposeRequest{
				Protocol: "tcp",
				Local:    net.JoinHostPort(constants.LocalIP, apiPort),
				Remote:   net.JoinHostPort(virtualMachineIP, apiPort),
			},
			types.ExposeRequest{
				Protocol: "tcp",
				Local:    fmt.Sprintf(":%d", ingressHTTPSPort),
				Remote:   net.JoinHostPort(virtualMachineIP, remoteHTTPSPort),
			},
			types.ExposeRequest{
				Protocol: "tcp",
				Local:    fmt.Sprintf(":%d", ingressHTTPPort),
				Remote:   net.JoinHostPort(virtualMachineIP, remoteHTTPPort),
			})
	default:
		logging.Errorf("Invalid preset: %s", preset)
	}

	return exposeRequest
}

func getSSHTunnelURI() string {
	u := url.URL{
		Scheme:     "ssh-tunnel",
		User:       url.User("core"),
		Host:       net.JoinHostPort(virtualMachineIP, internalSSHPort),
		Path:       "/run/podman/podman.sock",
		ForceQuery: false,
		RawQuery:   fmt.Sprintf("key=%s", url.QueryEscape(constants.GetPrivateKeyPath())),
	}
	return u.String()
}
