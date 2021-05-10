package network

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/spf13/cast"
)

type NameServer struct {
	IPAddress string
}

type SearchDomain struct {
	Domain string
}

type ResolvFileValues struct {
	SearchDomains []SearchDomain
	NameServers   []NameServer
}

type Mode string

const (
	BridgedNetworkingMode Mode = "bridged"
	VSockNetworkingMode   Mode = "vsock"
)

func parseMode(input string) (Mode, error) {
	switch input {
	case string(VSockNetworkingMode):
		return VSockNetworkingMode, nil
	case string(BridgedNetworkingMode), "default":
		return BridgedNetworkingMode, nil
	default:
		return BridgedNetworkingMode, fmt.Errorf("Cannot parse mode '%s'", input)
	}
}
func ParseMode(input string) Mode {
	mode, err := parseMode(input)
	if err != nil {
		logging.Errorf("unexpected network mode %s, using default", input)
		return BridgedNetworkingMode
	}
	return mode
}

func ValidateMode(val interface{}) (bool, string) {
	_, err := parseMode(cast.ToString(val))
	if err != nil {
		return false, fmt.Sprintf("network mode should be either %s or %s", BridgedNetworkingMode, VSockNetworkingMode)
	}
	return true, ""
}

func SuccessfullyAppliedMode(_ string, _ interface{}) string {
	return "Network mode changed. Please run `crc cleanup` and `crc setup`."
}
