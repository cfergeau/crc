package preflight

import (
	"fmt"
	"reflect"
	"runtime"

	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/preset"
)

type Flags uint32

const (
	// Indicates a PreflightCheck should only be run as part of "crc setup"
	SetupOnly Flags = 1 << iota
	NoFix
	CleanUpOnly
	StartUpOnly
)

type commonOptions struct {
	networkMode network.Mode
	bundlePath  string
	preset      preset.Preset
}

func (opts *commonOptions) getNetworkMode() network.Mode {
	return opts.networkMode
}

func (opts *commonOptions) getBundlePath() string {
	return opts.bundlePath
}

func (opts *commonOptions) getPreset() preset.Preset {
	return opts.preset
}

func commonOptionsNew(networkMode network.Mode, bundlePath string, preset preset.Preset) options {
	return &commonOptions{
		networkMode: networkMode,
		bundlePath:  bundlePath,
		preset:      preset,
	}
}

type options interface {
	getNetworkMode() network.Mode
	getBundlePath() string
	getPreset() preset.Preset
}

type CheckFunc func(opts options) error
type FixFunc func(opts options) error
type CleanUpFunc func(opts options) error

type Check struct {
	configKeySuffix    string
	checkDescription   string
	check              CheckFunc
	fixDescription     string
	fix                FixFunc
	flags              Flags
	cleanupDescription string
	cleanup            CleanUpFunc

	labels labels
}

func (check *Check) getSkipConfigName() string {
	if check.configKeySuffix == "" {
		return ""
	}
	return "skip-" + check.configKeySuffix
}

func (check *Check) shouldSkip(config crcConfig.Storage) bool {
	if check.configKeySuffix == "" {
		return false
	}
	return config.Get(check.getSkipConfigName()).AsBool()
}

func (check *Check) doCheck(config crcConfig.Storage, opts options) error {
	if check.checkDescription == "" {
		panic(fmt.Sprintf("Should not happen, empty description for check '%s'", check.configKeySuffix))
	} else {
		logging.Infof("%s", check.checkDescription)
	}
	if check.shouldSkip(config) {
		logging.Warn("Skipping above check...")
		return nil
	}

	err := check.check(opts)
	if err != nil {
		logging.Debug(err.Error())
	}
	return err
}

func (check *Check) doFix(opts options) error {
	if check.fixDescription == "" {
		panic(fmt.Sprintf("Should not happen, empty description for fix '%s'", check.configKeySuffix))
	}
	if check.flags&NoFix == NoFix {
		return fmt.Errorf(check.fixDescription)
	}

	logging.Infof("%s", check.fixDescription)

	return check.fix(opts)
}

func (check *Check) doCleanUp(opts options) error {
	printCheck(*check)
	if check.cleanupDescription == "" {
		panic(fmt.Sprintf("Should not happen, empty description for cleanup '%s'", check.configKeySuffix))
	}

	logging.Infof("%s", check.cleanupDescription)

	return check.cleanup(opts)
}

func doPreflightChecks(config crcConfig.Storage, opts options, checks []Check) error {
	for _, check := range checks {
		if check.flags&SetupOnly == SetupOnly || check.flags&CleanUpOnly == CleanUpOnly {
			continue
		}
		printCheck(check)
		if err := check.doCheck(config, opts); err != nil {
			return err
		}
	}
	return nil
}

func doFixPreflightChecks(config crcConfig.Storage, opts options, checks []Check, checkOnly bool) error {
	for _, check := range checks {
		if check.flags&CleanUpOnly == CleanUpOnly || check.flags&StartUpOnly == StartUpOnly {
			continue
		}
		printCheck(check)
		err := check.doCheck(config, opts)
		if err == nil {
			continue
		} else if checkOnly {
			return err
		}
		if err = check.doFix(opts); err != nil {
			return err
		}
	}
	return nil
}

func doCleanUpPreflightChecks(opts options, checks []Check) error {
	var mErr errors.MultiError
	// Do the cleanup in reverse order to avoid any dependency during cleanup
	for i := len(checks) - 1; i >= 0; i-- {
		check := checks[i]
		if check.cleanup == nil {
			continue
		}
		err := check.doCleanUp(opts)
		if err != nil {
			// If an error occurs in a cleanup function
			// we log/collect it and  move to the  next
			logging.Debug(err)
			mErr.Collect(err)
		}
	}
	if len(mErr.Errors) == 0 {
		return nil
	}
	return mErr
}

func doRegisterSettings(cfg crcConfig.Schema, checks []Check) {
	for _, check := range checks {
		if check.configKeySuffix != "" {
			cfg.AddSetting(check.getSkipConfigName(), false, crcConfig.ValidateBool, crcConfig.SuccessfullyApplied,
				"Skip preflight check (true/false, default: false)")
		}
	}
}

// StartPreflightChecks performs the preflight checks before starting the cluster
func StartPreflightChecks(config crcConfig.Storage) error {
	filter := newFilterFromConfig(config)

	mode := crcConfig.GetNetworkMode(config)
	bundlePath := crcConfig.GetBundlePath(config)
	preset := crcConfig.GetPreset(config)
	opts := optionsNew(mode, bundlePath, preset)

	if err := doPreflightChecks(config, opts, getFilteredChecks(filter)); err != nil {
		return &errors.PreflightError{Err: err}
	}
	return nil
}

func newFilterFromConfig(config crcConfig.Storage) preflightFilter {
	experimentalFeatures := config.Get(crcConfig.ExperimentalFeatures).AsBool()
	networkMode := crcConfig.GetNetworkMode(config)
	trayAutostart := config.Get(crcConfig.AutostartTray).AsBool()

	filter := newFilter()
	filter.SetNetworkMode(networkMode)
	filter.SetExperimental(experimentalFeatures)
	filter.SetTray(trayAutostart)

	return filter
}

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func SetupHost(config crcConfig.Storage, checkOnly bool) error {
	filter := newFilterFromConfig(config)

	mode := crcConfig.GetNetworkMode(config)
	bundlePath := crcConfig.GetBundlePath(config)
	logging.Infof("Using bundle path %s", bundlePath)
	preset := crcConfig.GetPreset(config)
	opts := optionsNew(mode, bundlePath, preset)

	return doFixPreflightChecks(config, opts, getFilteredChecks(filter), checkOnly)
}

func RegisterSettings(config crcConfig.Schema) {
	doRegisterSettings(config, getAllPreflightChecks())
}

func CleanUpHost() error {
	// A user can use setup with experiment flag
	// and not use cleanup with same flag, to avoid
	// any extra step/confusion we are just adding the checks
	// which are behind the experiment flag. This way cleanup
	// perform action in a sane way.
	return doCleanUpPreflightChecks(&commonOptions{}, getAllPreflightChecks())
}

func funcToString(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func printCheck(check Check) {
	if check.configKeySuffix != "" {
		logging.Infof("check %p: configKey: %s", &check, check.configKeySuffix)
		return
	}
	if check.check != nil {
		logging.Infof("check %p: checkFunc: %s", &check, funcToString(check.check))
		return
	}
	if check.fix != nil {
		logging.Infof("check %p: fixFunc: %s", &check, funcToString(check.fix))
		return
	}
	if check.cleanup != nil {
		logging.Infof("check %p: cleanupFunc: %s", &check, funcToString(check.cleanup))
		return
	}
	logging.Infof("check %p: %v", &check, check)
}
