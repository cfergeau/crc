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
	assert.Len(t, cfg.AllConfigs(), 12)
}

func TestCountPreflights(t *testing.T) {
	assert.Len(t, getPreflightChecks(false, false, network.BridgedNetworkingMode), 17)
	assert.Len(t, getPreflightChecks(true, true, network.BridgedNetworkingMode), 20)

	assert.Len(t, getPreflightChecks(false, false, network.VSockNetworkingMode), 18)
	assert.Len(t, getPreflightChecks(true, true, network.VSockNetworkingMode), 21)
}
