package preflight

import (
	"testing"

	"github.com/code-ready/crc/pkg/crc/network"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/stretchr/testify/assert"
)

func TestCountConfigurationOptions(t *testing.T) {
	cfg := config.New(config.NewEmptyInMemoryStorage())
	RegisterSettings(cfg)
	assert.Len(t, cfg.AllConfigs(), 9)
}

func TestCountPreflights(t *testing.T) {
	assert.Len(t, getPreflightChecks(true, false, network.BridgedNetworkingMode), 15)
	assert.Len(t, getPreflightChecks(true, true, network.BridgedNetworkingMode), 15)

	assert.Len(t, getPreflightChecks(true, false, network.VSockNetworkingMode), 14)
	assert.Len(t, getPreflightChecks(true, true, network.VSockNetworkingMode), 14)
}
