package preflight

import (
	"testing"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/stretchr/testify/assert"
)

func TestCountConfigurationOptions(t *testing.T) {
	cfg := config.New(config.NewEmptyInMemoryStorage())
	RegisterSettings(cfg)
	assert.Len(t, cfg.AllConfigs(), 11)
}

func TestCountPreflights(t *testing.T) {
	assert.Len(t, getFilteredChecks(newTestFilter(true, false, network.SystemNetworkingMode)), 17)
	assert.Len(t, getFilteredChecks(newTestFilter(true, true, network.SystemNetworkingMode)), 17)

	assert.Len(t, getFilteredChecks(newTestFilter(true, false, network.UserNetworkingMode)), 16)
	assert.Len(t, getFilteredChecks(newTestFilter(true, true, network.UserNetworkingMode)), 16)
}

/*
func newTestFilter(experimental bool, trayAutostart bool, networkMode network.Mode) preflightFilter {
	filter := newFilter()
	filter.SetNetworkMode(networkMode)
	filter.SetExperimental(experimental)
	filter.SetTray(trayAutostart)

	return filter
}
*/
