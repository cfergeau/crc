package daemon

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/api"
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/network"

	"github.com/code-ready/gvisor-tap-vsock/pkg/types"
	"github.com/code-ready/gvisor-tap-vsock/pkg/virtualnetwork"

	"github.com/docker/go-units"
)

type Daemon struct {
	config            *config.Config
	virtualNetwork    *virtualnetwork.VirtualNetwork
	usesVsock         bool
	hostNetworkAccess bool
	debug             bool
	errorChannel      chan error
}

const hostVirtualIP = "192.168.127.254"

func (daemon *Daemon) getVirtualNetwork() (*virtualnetwork.VirtualNetwork, error) {
	if !daemon.usesVsock {
		logging.Debugf("vsock networking is disabled")
		return nil, nil
	}

	virtualNetworkConfig := types.Configuration{
		Debug:             false, // never log packets
		CaptureFile:       captureFile(daemon.debug),
		MTU:               4000, // Large packets slightly improve the performance. Less small packets.
		Subnet:            "192.168.127.0/24",
		GatewayIP:         constants.VSockGateway,
		GatewayMacAddress: "\x5A\x94\xEF\xE4\x0C\xDD",
		DNS: []types.Zone{
			{
				Name:      "apps-crc.testing.",
				DefaultIP: net.ParseIP("192.168.127.2"),
			},
			{
				Name: "crc.testing.",
				Records: []types.Record{
					{
						Name: "gateway",
						IP:   net.ParseIP("192.168.127.1"),
					},
					{
						Name: "api",
						IP:   net.ParseIP("192.168.127.2"),
					},
					{
						Name: "api-int",
						IP:   net.ParseIP("192.168.127.2"),
					},
					{
						Regexp: regexp.MustCompile("crc-(.*?)-master-0"),
						IP:     net.ParseIP("192.168.126.11"),
					},
				},
			},
		},
	}

	if daemon.hostNetworkAccess {
		logging.Debugf("Enabling host network access")
		for i := range virtualNetworkConfig.DNS {
			zone := &virtualNetworkConfig.DNS[i]
			if zone.Name != "crc.testing." {
				continue
			}
			logging.Debugf("Adding \"host\" -> %s DNS record to crc.testing. zone", hostVirtualIP)

			zone.Records = append(zone.Records, types.Record{Name: "host", IP: net.ParseIP(hostVirtualIP)})
		}

		if virtualNetworkConfig.NAT == nil {
			virtualNetworkConfig.NAT = make(map[string]string)
		}
		virtualNetworkConfig.NAT[hostVirtualIP] = "127.0.0.1"
	}

	return virtualnetwork.New(&virtualNetworkConfig)
}

func New(config *config.Config) (*Daemon, error) {
	daemon := &Daemon{}

	networkMode := network.ParseMode(config.Get(cmdConfig.NetworkMode).AsString())
	daemon.usesVsock = (networkMode == network.VSockMode)
	daemon.hostNetworkAccess = config.Get(cmdConfig.HostNetworkAccess).AsBool()
	daemon.errorChannel = make(chan error)

	return daemon, nil
}

func (daemon Daemon) SetDebug(debug bool) {
	daemon.debug = debug
}

func (daemon Daemon) Start() error {
	vn, err := daemon.getVirtualNetwork()
	if err != nil {
		return nil
	}

	if err := daemon.startHostListener(vn); err != nil {
		return err
	}

	if err := daemon.startVsockListener(vn); err != nil {
		return err
	}

	if err := daemon.startLegacyDaemon(); err != nil {
		return err
	}

	return nil
}

func (daemon Daemon) Run() error {
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigterm:
		return nil
	case err := <-daemon.errorChannel:
		return err
	}
}

func (daemon *Daemon) errorNotify(err error) {
	daemon.errorChannel <- err
}

func (daemon *Daemon) startHostListener(vn *virtualnetwork.VirtualNetwork) error {
	listener, err := httpListener()
	if err != nil {
		return err
	}

	machine := machine.NewClient(constants.DefaultName, daemon.debug, daemon.config)

	go func() {
		if listener == nil {
			return
		}
		mux := http.NewServeMux()
		if vn != nil {
			mux.Handle("/network/", http.StripPrefix("/network", vn.Mux()))
		}
		mux.Handle("/api/", http.StripPrefix("/api", api.NewMux(daemon.config, machine)))
		if err := http.Serve(listener, mux); err != nil {
			daemon.errorNotify(err)
		}
	}()

	return nil
}

func (daemon *Daemon) startVsockListener(vn *virtualnetwork.VirtualNetwork) error {
	if vn == nil {
		return fmt.Errorf("vsock networking is disabled")
	}

	vsockListener, err := vsockListener()
	if err != nil {
		return err
	}

	go func() {
		mux := http.NewServeMux()
		mux.Handle(types.ConnectPath, vn.Mux())
		if err := http.Serve(vsockListener, mux); err != nil {
			daemon.errorNotify(err)
		}
	}()

	if daemon.debug {
		go func() {
			for {
				fmt.Printf("%v sent to the VM, %v received from the VM\n", units.HumanSize(float64(vn.BytesSent())), units.HumanSize(float64(vn.BytesReceived())))
				time.Sleep(5 * time.Second)
			}
		}()
	}

	return nil
}

func (daemon *Daemon) startLegacyDaemon() error {
	machine := machine.NewClient(constants.DefaultName, daemon.debug, daemon.config)
	// Remove if an old socket is present
	os.Remove(constants.DaemonSocketPath)
	apiServer, err := api.CreateServer(constants.DaemonSocketPath, daemon.config, machine, logging.Memory)
	if err != nil {
		return err
	}

	go func() {
		if err := apiServer.Serve(); err != nil {
			daemon.errorNotify(err)
		}
	}()

	return nil
}

func captureFile(isDebug bool) string {
	if !isDebug {
		return ""
	}
	return filepath.Join(constants.CrcBaseDir, "capture.pcap")
}
