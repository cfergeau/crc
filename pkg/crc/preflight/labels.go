package preflight

import (
	"runtime"

	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/version"
)

type LabelName uint32

const (
	Os LabelName = iota
	NetworkMode
	BuildType
	DaemonStartup

	// Keep it last
	// will be used in OS-specific go files to extend LabelName
	lastLabelName // nolint
)

type LabelValue uint32

const (
	// os
	Darwin LabelValue = iota
	Linux
	Windows

	// network mode
	User
	System

	// build type
	Installer
	Standalone
	// daemon startup
	Manual
	SocketActivated

	// Keep it last
	// will be used in OS-specific go files to extend LabelValue
	lastLabelValue // nolint
)

var (
	None = labels{}
)

type labels map[LabelName]LabelValue

type preflightFilter map[LabelName]LabelValue

func newFilter() preflightFilter {
	filter := preflightFilter{}
	filter.setBuildType()
	filter.setOs()
	filter.SetDaemonStartup(Manual)

	return filter
}

func (filter preflightFilter) setOs() {
	switch runtime.GOOS {
	case "darwin":
		filter[Os] = Darwin
	case "linux":
		filter[Os] = Linux
	case "windows":
		filter[Os] = Windows
	default:
		// In case of different platform (Should not happen)
		filter[Os] = Linux
	}
}

func (filter preflightFilter) SetDaemonStartup(startupMode LabelValue) {
	filter[DaemonStartup] = startupMode
}

func (filter preflightFilter) SetNetworkMode(networkMode network.Mode) {
	switch networkMode {
	case network.SystemNetworkingMode:
		filter[NetworkMode] = System
	case network.UserNetworkingMode:
		filter[NetworkMode] = User
	}
}

func (filter preflightFilter) setBuildType() {
	if version.IsInstaller() {
		filter[BuildType] = Installer
	}
	filter[BuildType] = Standalone
}

/* This will iterate over 'checks' and only keep the checks which match the filter:
 * - if a key is present in the filter and not in the check labels, the check is kept
 * - if a key is present in the check labels, but not in the filter, the check is kept
 * - if a key is present both in the filter and in the check labels, the check
 *   is kept only if they have the same value, it is dropped if their values differ
 */
func (filter preflightFilter) Apply(checks []Check) []Check {
	var filteredChecks []Check

	for _, check := range checks {
		if !skipCheck(check, filter) {
			filteredChecks = append(filteredChecks, check)
		}

	}

	return filteredChecks
}

func skipCheck(check Check, filter preflightFilter) bool {
	for filterKey, filterValue := range filter {
		checkValue, present := check.labels[filterKey]
		if present && checkValue != filterValue {
			return true
		}
	}

	return false
}
