package config

import (
	"sort"
	"strings"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/units"
	"github.com/spf13/cobra"
)

const (
	Bundle                  = "bundle"
	CPUs                    = "cpus"
	Memory                  = "memory"
	DiskSize                = "disk-size"
	NameServer              = "nameserver"
	PullSecretFile          = "pull-secret-file"
	DisableUpdateCheck      = "disable-update-check"
	ExperimentalFeatures    = "enable-experimental-features"
	NetworkMode             = "network-mode"
	HTTPProxy               = "http-proxy"
	HTTPSProxy              = "https-proxy"
	NoProxy                 = "no-proxy"
	ProxyCAFile             = "proxy-ca-file"
	ConsentTelemetry        = "consent-telemetry"
	EnableClusterMonitoring = "enable-cluster-monitoring"
)

func RegisterSettings(cfg *config.Config) {
	// Start command settings in config
	cfg.AddSetting(Bundle, constants.DefaultBundlePath, config.ValidateBundle, config.SuccessfullyApplied)
	cfg.AddSetting(CPUs, constants.DefaultCPUs, config.ValidateCPUs, config.RequiresRestartMsg)
	cfg.AddSetting(Memory, constants.DefaultMemory, config.ValidateMemory, config.RequiresRestartMsg)
	cfg.AddSetting(DiskSize, constants.DefaultDiskSize, config.ValidateDiskSize, config.RequiresRestartMsg)
	cfg.AddSetting(NameServer, "", config.ValidateIPAddress, config.SuccessfullyApplied)
	cfg.AddSetting(PullSecretFile, "", config.ValidatePath, config.SuccessfullyApplied)
	cfg.AddSetting(DisableUpdateCheck, false, config.ValidateBool, config.SuccessfullyApplied)
	cfg.AddSetting(ExperimentalFeatures, false, config.ValidateBool, config.SuccessfullyApplied)
	cfg.AddSetting(NetworkMode, string(network.DefaultMode), network.ValidateMode, network.SuccessfullyAppliedMode)
	// Proxy Configuration
	cfg.AddSetting(HTTPProxy, "", config.ValidateURI, config.SuccessfullyApplied)
	cfg.AddSetting(HTTPSProxy, "", config.ValidateURI, config.SuccessfullyApplied)
	cfg.AddSetting(NoProxy, "", config.ValidateNoProxy, config.SuccessfullyApplied)
	cfg.AddSetting(ProxyCAFile, "", config.ValidatePath, config.SuccessfullyApplied)

	cfg.AddSetting(EnableClusterMonitoring, false, config.ValidateBool, config.SuccessfullyApplied)

	// Telemeter Configuration
	cfg.AddSetting(ConsentTelemetry, "", config.ValidateYesNo, config.SuccessfullyApplied)
}

func MigrateOldConfig(cfg *config.Config) bool {
	var configMigrated = false
	/* Prior to crc 1.2x, 'memory' was stored in MiB in the config file, and 'disk-size' was stored in GiB.
	 * After this version, they are stored in bytes. This function will convert these 2 keys to the new
	 * values
	 */
	memorySize := cfg.Get(Memory).AsSize().ToBytes()
	defaultMemory := constants.DefaultMemory.ToBytes()
	/* old value in config file is something like 9216 (old default, 9MiB)
	 * new value will be the same, but in bytes (9,663,676,416)
	 */
	if memorySize <= defaultMemory && memorySize*uint64(units.MiB) >= defaultMemory {
		logging.Infof("Migrating old configuration file, updating 'memory' setting from '%d' to '%d MiB'", memorySize, memorySize)
		_, err := cfg.Set(Memory, units.New(memorySize, units.MiB))
		if err != nil {
			logging.Warnf("Failed to migrate memory setting: %v", err)
		}
		configMigrated = true
	}

	/* old value in config file was in GiB (old default, 35)
	 * new value is stored in bytes
	 */
	diskSize := cfg.Get(DiskSize).AsSize().ToBytes()
	defaultDiskSize := constants.DefaultDiskSize.ConvertTo(units.Bytes)
	if diskSize <= defaultDiskSize && diskSize*uint64(units.GiB) >= defaultDiskSize {
		logging.Infof("Migrating old configuration file, updating 'disk-size' setting from '%d' to '%d GiB'", diskSize, diskSize)
		_, err := cfg.Set(DiskSize, units.New(diskSize, units.GiB))
		if err != nil {
			logging.Warnf("Failed to migrate disk size setting: %v", err)
		}
		configMigrated = true
	}

	return configMigrated
}

func isPreflightKey(key string) bool {
	return strings.HasPrefix(key, "skip-")
}

// less is used to sort the config keys. We want to sort first the regular keys, and
// then the keys related to preflight starting with a skip- prefix.
func less(lhsKey, rhsKey string) bool {
	if isPreflightKey(lhsKey) {
		if isPreflightKey(rhsKey) {
			// ignore skip prefix
			return lhsKey[4:] < rhsKey[4:]
		}
		// lhs is preflight, rhs is not preflight
		return false
	}

	if isPreflightKey(rhsKey) {
		// lhs is not preflight, rhs is preflight
		return true
	}

	// lhs is not preflight, rhs is not preflight
	return lhsKey < rhsKey
}

func configurableFields(config *config.Config) string {
	var fields []string
	var keys []string

	for key := range config.AllConfigs() {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return less(keys[i], keys[j])
	})
	for _, key := range keys {
		fields = append(fields, " * "+key)
	}
	return strings.Join(fields, "\n")
}

func GetConfigCmd(config *config.Config) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config SUBCOMMAND [flags]",
		Short: "Modify crc configuration",
		Long: `Modifies crc configuration properties.
Configurable properties (enter as SUBCOMMAND): ` + "\n\n" + configurableFields(config),
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	configCmd.AddCommand(configGetCmd(config))
	configCmd.AddCommand(configSetCmd(config))
	configCmd.AddCommand(configUnsetCmd(config))
	configCmd.AddCommand(configViewCmd(config))
	return configCmd
}
