package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

type testConfig struct {
	*Config
	configDir string
}

func (config *testConfig) Close() {
	os.RemoveAll(config.configDir)
}

func (config *testConfig) FileContent() ([]byte, error) {
	configFile := filepath.Join(config.configDir, "crc.json")
	return ioutil.ReadFile(configFile)
}

func newTestConfig() (*testConfig, error) {
	return newTestConfigWithInitialContent([]byte{})
}

func newTestConfigWithInitialContent(fileContent []byte) (*testConfig, error) {
	dir, err := ioutil.TempDir("", "cfg")
	if err != nil {
		return nil, err
	}
	configFile := filepath.Join(dir, "crc.json")
	testConfig := testConfig{configDir: dir}

	if len(fileContent) != 0 {
		err := ioutil.WriteFile(configFile, fileContent, 0600)
		if err != nil {
			return nil, err
		}
	}

	err = testConfig.Reload()
	if err != nil {
		testConfig.Close()
		return nil, err
	}
	return &testConfig, nil
}

func (config *testConfig) Reload() error {
	configFile := filepath.Join(config.configDir, "crc.json")
	storage, err := NewViperStorage(configFile, "CRC")
	if err != nil {
		return err
	}
	cfg := New(storage)
	cfg.AddSetting(CPUs, 4, ValidateCPUs, RequiresRestartMsg)
	cfg.AddSetting(NameServer, "", ValidateIPAddress, SuccessfullyApplied)
	cfg.AddSetting(Memory, units.New(7, units.GiB), ValidateSize, SuccessfullyApplied)
	config.Config = cfg
	return nil
}

func TestViperConfigUnknown(t *testing.T) {
	config, err := newTestConfig()
	require.NoError(t, err)
	defer config.Close()

	assert.Equal(t, SettingValue{
		Invalid: true,
	}, config.Get("foo"))
}

func TestViperConfigSizeType(t *testing.T) {
	config, err := newTestConfig()
	require.NoError(t, err)
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
		Value:     units.Size(4) ,
		IsDefault: false,
	}, config.Get(Memory))

	_, err = config.Set(Memory, 4)
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     units.Size(4) ,
		IsDefault: false,
	}, config.Get(Memory))

}

func TestViperConfigSizeBackwardCompat(t *testing.T) {
	config, err := newTestConfig()
	require.NoError(t, err)
	defer config.Close()

	_, err = config.Set(Memory, units.New(8, units.KiB))
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     units.Size(8192) ,
		IsDefault: false,
	}, config.Get(Memory))

	_, err = config.Set(Memory, "16 MB")
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     units.Size(16_000_000) ,
		IsDefault: false,
	}, config.Get(Memory))

	/* Test behaviour with unitless values */
	_, err = config.Set(Memory, "4")
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     units.Size(4) ,
		IsDefault: false,
	}, config.Get(Memory))

	_, err = config.Set(Memory, 4)
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     units.Size(4) ,
		IsDefault: false,
	}, config.Get(Memory))
}

func TestViperConfigSetAndGet(t *testing.T) {
	config, err := newTestConfig()
	require.NoError(t, err)
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
	config, err := newTestConfigWithInitialContent([]byte("{\"cpus\": 5}"))
	require.NoError(t, err)
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
	defer config.Close()

	_, err = config.Set(CPUs, 5)
	require.NoError(t, err)

	err = config.Reload()
	require.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config.Get(CPUs))
}

func TestViperConfigLoadDefaultValue(t *testing.T) {
	config, err := newTestConfig()
	require.NoError(t, err)
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

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(CPUs))
}

func TestViperConfigBindFlagSet(t *testing.T) {
	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	configFile := filepath.Join(dir, "crc.json")

	storage, err := NewViperStorage(configFile, "CRC")
	require.NoError(t, err)
	config := New(storage)
	config.AddSetting(CPUs, 4, ValidateCPUs, RequiresRestartMsg)
	config.AddSetting(NameServer, "", ValidateIPAddress, SuccessfullyApplied)

	flagSet := pflag.NewFlagSet("start", pflag.ExitOnError)
	flagSet.IntP(CPUs, "c", 4, "")
	flagSet.StringP(NameServer, "n", "", "")

	_ = storage.BindFlagSet(flagSet)

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
	defer config.Close()

	_, err = config.Set(CPUs, "5")
	require.NoError(t, err)

	err = config.Reload()
	require.NoError(t, err)

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
	defer config.Close()

	_, err = config.Set(CPUs, "helloworld")
	assert.EqualError(t, err, "Value 'helloworld' for configuration property 'cpus' is invalid, reason: unable to cast \"helloworld\" of type string to int")
}

func TestCannotGetWithWrongType(t *testing.T) {
	config, err := newTestConfigWithInitialContent([]byte("{\"cpus\": \"hello\"}"))
	require.NoError(t, err)
	defer config.Close()

	assert.True(t, config.Get(CPUs).Invalid)
}
