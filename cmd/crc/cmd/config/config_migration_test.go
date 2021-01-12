package config

import (
	"testing"

	cfg "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/units"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateOldConfig(t *testing.T) {
	oldConfig := "{\"disk-size\": 50, \"memory\": 10000 }"
	config, err := cfg.NewTestConfigWithInitialContent([]byte(oldConfig))
	require.NoError(t, err)
	RegisterSettings(config.Config)
	configMigrated := MigrateOldConfig(config.Config)
	assert.True(t, configMigrated)

	assert.Equal(t, cfg.SettingValue{
		Value:     units.New(50, units.GiB),
		IsDefault: false,
	}, config.Get(DiskSize))
	assert.Equal(t, cfg.SettingValue{
		Value:     units.New(10000, units.MiB),
		IsDefault: false,
	}, config.Get(Memory))
}

func TestNoMigration(t *testing.T) {
	newConfig := "{\"disk-size\": 50000000000, \"memory\": 10000000000 }"
	config, err := cfg.NewTestConfigWithInitialContent([]byte(newConfig))
	require.NoError(t, err)
	RegisterSettings(config.Config)
	configMigrated := MigrateOldConfig(config.Config)
	assert.False(t, configMigrated)

	assert.Equal(t, cfg.SettingValue{
		Value:     units.New(50, units.GB),
		IsDefault: false,
	}, config.Get(DiskSize))
	assert.Equal(t, cfg.SettingValue{
		Value:     units.New(10000, units.MB),
		IsDefault: false,
	}, config.Get(Memory))
}
