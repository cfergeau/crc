package preflight

import (
	"errors"
	"testing"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/network"
	crcpreset "github.com/code-ready/crc/pkg/crc/preset"

	"github.com/stretchr/testify/assert"
)

func TestCheckPreflight(t *testing.T) {
	check, calls := sampleCheck(nil, nil)
	cfg := config.New(config.NewEmptyInMemoryStorage())
	doRegisterSettings(cfg, []Check{*check})

	opts := optionsNew(network.UserNetworkingMode, constants.GetDefaultBundlePath(crcpreset.OpenShift), crcpreset.OpenShift)
	assert.NoError(t, doPreflightChecks(cfg, opts, []Check{*check}))
	assert.True(t, calls.checked)
	assert.False(t, calls.fixed)
}

func TestSkipPreflight(t *testing.T) {
	check, calls := sampleCheck(nil, nil)
	cfg := config.New(config.NewEmptyInMemoryStorage())
	doRegisterSettings(cfg, []Check{*check})
	_, err := cfg.Set("skip-sample", true)
	assert.NoError(t, err)

	opts := optionsNew(network.UserNetworkingMode, constants.GetDefaultBundlePath(crcpreset.OpenShift), crcpreset.OpenShift)
	assert.NoError(t, doPreflightChecks(cfg, opts, []Check{*check}))
	assert.False(t, calls.checked)
}

func TestFixPreflight(t *testing.T) {
	check, calls := sampleCheck(errors.New("check failed"), nil)
	cfg := config.New(config.NewEmptyInMemoryStorage())
	doRegisterSettings(cfg, []Check{*check})

	opts := optionsNew(network.UserNetworkingMode, constants.GetDefaultBundlePath(crcpreset.OpenShift), crcpreset.OpenShift)
	assert.NoError(t, doFixPreflightChecks(cfg, opts, []Check{*check}, false))
	assert.True(t, calls.checked)
	assert.True(t, calls.fixed)
}

func TestFixPreflightCheckOnly(t *testing.T) {
	check, calls := sampleCheck(errors.New("check failed"), nil)
	cfg := config.New(config.NewEmptyInMemoryStorage())
	doRegisterSettings(cfg, []Check{*check})

	opts := optionsNew(network.UserNetworkingMode, constants.GetDefaultBundlePath(crcpreset.OpenShift), crcpreset.OpenShift)
	assert.Error(t, doFixPreflightChecks(cfg, opts, []Check{*check}, true))
	assert.True(t, calls.checked)
	assert.False(t, calls.fixed)
}

func sampleCheck(checkErr, fixErr error) (*Check, *status) {
	status := &status{}
	return &Check{
		configKeySuffix:  "sample",
		checkDescription: "Sample check",
		check: func(_ options) error {
			status.checked = true
			return checkErr
		},
		fixDescription: "sample fix",
		fix: func(_ options) error {
			status.fixed = true
			return fixErr
		},
	}, status
}

type status struct {
	checked, fixed bool
}
