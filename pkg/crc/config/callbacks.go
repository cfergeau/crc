package config

import (
	"fmt"

	"github.com/spf13/cast"
)

func requiresRestartMsg(key string, _ interface{}) string {
	return fmt.Sprintf("Changes to configuration property '%s' are only applied when the CRC instance is started.\n"+
		"If you already have a running CRC instance, then for this configuration change to take effect, "+
		"stop the CRC instance with 'crc stop' and restart it with 'crc start'.", key)
}

func SuccessfullyApplied(key string, value interface{}) string {
	return fmt.Sprintf("Successfully configured %s to %s", key, cast.ToString(value))
}

func disableEnableTrayAutostart(key string, value interface{}) string {
	if cast.ToBool(value) {
		return fmt.Sprintf(
			"Successfully configured '%s' to '%s'. Run 'crc setup' for it to take effect.",
			key, cast.ToString(value),
		)
	}
	return fmt.Sprintf(
		"Successfully configured '%s' to '%s'. Run 'crc cleanup' and then 'crc setup' for it to take effect.",
		key, cast.ToString(value),
	)
}
