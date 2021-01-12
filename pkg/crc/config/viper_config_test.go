package config

import (
	"testing"

	"github.com/code-ready/crc/pkg/units"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	CPUs       = "cpus"
	NameServer = "nameservers"
	Memory     = "memory"
)

func addSettings(config *TestConfig) {
	config.AddSetting(CPUs, 4, ValidateCPUs, RequiresRestartMsg)
	config.AddSetting(NameServer, "", ValidateIPAddress, SuccessfullyApplied)
	config.AddSetting(Memory, units.New(7, units.GiB), ValidateSize, SuccessfullyApplied)
}

func TestViperConfigUnknown(t *testing.T) {
	config, err := newTestConfig()
	require.NoError(t, err)
	addSettings(config)
	defer config.Close()

	assert.Equal(t, SettingValue{
		Invalid: true,
	}, config.Get("foo"))
}

func TestViperConfigSizeType(t *testing.T) {
	config, err := newTestConfig()
	require.NoError(t, err)
	addSettings(config)
	defer config.Close()

	_, err = config.Set(Memory, units.New(8, units.KiB))
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     units.Size(8192),
		IsDefault: false,
	}, config.Get(Memory))

	_, err = config.Set(Memory, "16 MB")
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     units.Size(16_000_000),
		IsDefault: false,
	}, config.Get(Memory))

	/* Test behaviour with unitless values */
	_, err = config.Set(Memory, "4")
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     units.Size(4),
		IsDefault: false,
	}, config.Get(Memory))

	/*
		_, err = config.Set(Memory, 4)
		assert.NoError(t, err)

		assert.Equal(t, SettingValue{
			Value:     units.Size(4),
			IsDefault: false,
		}, config.Get(Memory))
	*/

}

/*
func TestViperConfigSizeBackwardCompat(t *testing.T) {
	config, err := NewTestConfigWithInitialContent([]byte("{\"memory\": 10123}"))
	require.NoError(t, err)
	addSettings(config)
	defer config.Close()

	assert.Equal(t, SettingValue{
		Value:     units.Size(4) ,
		IsDefault: false,
	}, config.Get(Memory))

	// Test behaviour with unitless values
	_, err = config.Set(Memory, "4")
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     units.Size(4),
		IsDefault: false,
	}, config.Get(Memory))

	_, err = config.Set(Memory, 4)
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     units.Size(4),
		IsDefault: false,
	}, config.Get(Memory))
}
*/

func TestViperConfigSetAndGet(t *testing.T) {
	config, err := newTestConfig()
	require.NoError(t, err)
	addSettings(config)
	defer config.Close()

	_, err = config.Set(CPUs, 5)
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config.Get(CPUs))

	bin, err := config.FileContent()
	assert.NoError(t, err)
	assert.JSONEq(t, `{"cpus":5}`, string(bin))
}

func TestViperConfigUnsetAndGet(t *testing.T) {
	config, err := NewTestConfigWithInitialContent([]byte("{\"cpus\": 5}"))
	require.NoError(t, err)
	addSettings(config)
	defer config.Close()

	_, err = config.Unset(CPUs)
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(CPUs))

	bin, err := config.FileContent()
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(bin))
}

func TestViperConfigSetReloadAndGet(t *testing.T) {
	config, err := newTestConfig()
	require.NoError(t, err)
	addSettings(config)
	defer config.Close()

	_, err = config.Set(CPUs, 5)
	require.NoError(t, err)

	err = config.Reload()
	require.NoError(t, err)
	addSettings(config)

	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config.Get(CPUs))
}

func TestViperConfigLoadDefaultValue(t *testing.T) {
	config, err := newTestConfig()
	require.NoError(t, err)
	addSettings(config)
	defer config.Close()

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(CPUs))

	_, err = config.Set(CPUs, 4)
	assert.NoError(t, err)

	bin, err := config.FileContent()
	assert.NoError(t, err)
	assert.JSONEq(t, `{"cpus":4}`, string(bin))

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(CPUs))

	err = config.Reload()
	require.NoError(t, err)
	addSettings(config)

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(CPUs))
}

func TestViperConfigBindFlagSet(t *testing.T) {
	config, err := newTestConfig()
	require.NoError(t, err)
	config.AddSetting(CPUs, 4, ValidateCPUs, RequiresRestartMsg)
	config.AddSetting(NameServer, "", ValidateIPAddress, SuccessfullyApplied)
	defer config.Close()

	flagSet := pflag.NewFlagSet("start", pflag.ExitOnError)
	flagSet.IntP(CPUs, "c", 4, "")
	flagSet.StringP(NameServer, "n", "", "")

	_ = config.StorageBindFlagSet(flagSet)

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(CPUs))
	assert.Equal(t, SettingValue{
		Value:     "",
		IsDefault: true,
	}, config.Get(NameServer))

	assert.NoError(t, flagSet.Set(CPUs, "5"))

	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config.Get(CPUs))

	_, err = config.Set(CPUs, "6")
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     6,
		IsDefault: false,
	}, config.Get(CPUs))
}

func TestViperConfigCastSet(t *testing.T) {
	config, err := newTestConfig()
	require.NoError(t, err)
	addSettings(config)
	defer config.Close()

	_, err = config.Set(CPUs, "5")
	require.NoError(t, err)

	err = config.Reload()
	require.NoError(t, err)
	addSettings(config)

	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config.Get(CPUs))

	bin, err := config.FileContent()
	assert.NoError(t, err)
	assert.JSONEq(t, `{"cpus": 5}`, string(bin))
}

func TestCannotSetWithWrongType(t *testing.T) {
	config, err := newTestConfig()
	require.NoError(t, err)
	addSettings(config)
	defer config.Close()

	_, err = config.Set(CPUs, "helloworld")
	assert.EqualError(t, err, "Value 'helloworld' for configuration property 'cpus' is invalid, reason: unable to cast \"helloworld\" of type string to int")
}

func TestCannotGetWithWrongType(t *testing.T) {
	config, err := NewTestConfigWithInitialContent([]byte("{\"cpus\": \"hello\"}"))
	require.NoError(t, err)
	addSettings(config)
	defer config.Close()

	assert.True(t, config.Get(CPUs).Invalid)
}
