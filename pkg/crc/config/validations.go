package config

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/validation"
	"github.com/spf13/cast"
)

func validationError(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// ValidateBool is a fail safe in the case user
// makes a typo for boolean config values
func ValidateBool(value interface{}) error {
	if _, err := cast.ToBoolE(value); err != nil {
		return validationError("must be true or false")
	}

	return nil
}

func ValidateString(value interface{}) error {
	if _, err := cast.ToStringE(value); err != nil {
		return validationError("must be a valid string")
	}
	return nil
}

// ValidateDiskSize checks if provided disk size is valid in the config
func ValidateDiskSize(value interface{}) error {
	diskSize, err := cast.ToIntE(value)
	if err != nil {
		return validationError("could not convert '%s' to integer", value)
	}
	if err := validation.ValidateDiskSize(diskSize); err != nil {
		return validationError(err.Error())
	}

	return nil
}

// ValidateCPUs checks if provided cpus count is valid in the config
func ValidateCPUs(value interface{}) error {
	v, err := cast.ToIntE(value)
	if err != nil {
		return validationError("requires integer value >= %d", constants.DefaultCPUs)
	}
	if err := validation.ValidateCPUs(v); err != nil {
		return validationError(err.Error())
	}
	return nil
}

// ValidateMemory checks if provided memory is valid in the config
func ValidateMemory(value interface{}) error {
	v, err := cast.ToIntE(value)
	if err != nil {
		return validationError("requires integer value in MiB >= %d", constants.DefaultMemory)
	}
	if err := validation.ValidateMemory(v); err != nil {
		return validationError(err.Error())
	}
	return nil
}

// ValidateBundlePath checks if the provided bundle path is valid
func ValidateBundlePath(value interface{}) error {
	if err := validation.ValidateBundlePath(cast.ToString(value)); err != nil {
		return validationError(err.Error())
	}
	return nil
}

// ValidateIP checks if provided IP is valid
func ValidateIPAddress(value interface{}) error {
	if err := validation.ValidateIPAddress(cast.ToString(value)); err != nil {
		return validationError(err.Error())
	}
	return nil
}

// ValidatePath checks if provided path is exist
func ValidatePath(value interface{}) error {
	if err := validation.ValidatePath(cast.ToString(value)); err != nil {
		return validationError(err.Error())
	}
	return nil
}

// ValidateHTTPProxy checks if given URI is valid for a HTTP proxy
func ValidateHTTPProxy(value interface{}) error {
	if err := network.ValidateProxyURL(cast.ToString(value), false); err != nil {
		return validationError(err.Error())
	}
	return nil
}

// ValidateHTTPSProxy checks if given URI is valid for a HTTPS proxy
func ValidateHTTPSProxy(value interface{}) error {
	if err := network.ValidateProxyURL(cast.ToString(value), true); err != nil {
		return validationError(err.Error())
	}
	return nil
}

// ValidateNoProxy checks if the NoProxy string has the correct format
func ValidateNoProxy(value interface{}) error {
	if strings.Contains(cast.ToString(value), " ") {
		return validationError("NoProxy string can't contain spaces")
	}
	return nil
}

func ValidateYesNo(value interface{}) error {
	if cast.ToString(value) == "yes" || cast.ToString(value) == "no" {
		return nil
	}
	return validationError("must be yes or no")
}

func validateTrayAutostart(value interface{}) error {
	if runtime.GOOS == "linux" {
		return validationError("Tray autostart is only supported on macOS and windows")
	}
	return ValidateBool(value)
}
