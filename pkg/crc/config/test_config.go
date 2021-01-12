package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	//"testing"

	//"github.com/code-ready/crc/pkg/units"

	"github.com/spf13/pflag"
	//"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
)

type TestConfig struct {
	*Config
	configDir    string
	viperStorage *ViperStorage
}

func (config *TestConfig) Close() {
	os.RemoveAll(config.configDir)
}

func (config *TestConfig) StorageBindFlagSet(flagSet *pflag.FlagSet) error {
	return config.viperStorage.BindFlagSet(flagSet)
}

func (config *TestConfig) FileContent() ([]byte, error) {
	configFile := filepath.Join(config.configDir, "crc.json")
	return ioutil.ReadFile(configFile)
}

func newTestConfig() (*TestConfig, error) {
	return NewTestConfigWithInitialContent([]byte{})
}

func NewTestConfigWithInitialContent(fileContent []byte) (*TestConfig, error) {
	dir, err := ioutil.TempDir("", "cfg")
	if err != nil {
		return nil, err
	}
	configFile := filepath.Join(dir, "crc.json")
	testConfig := TestConfig{configDir: dir}

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

func (config *TestConfig) Reload() error {
	configFile := filepath.Join(config.configDir, "crc.json")
	storage, err := NewViperStorage(configFile, "CRC")
	if err != nil {
		return err
	}
	config.Config = New(storage)
	config.viperStorage = storage
	return nil
}
