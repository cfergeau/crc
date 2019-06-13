package ipc

import (
	"github.com/code-ready/crc/pkg/crc/machine"
)

func Start(startConfig StartConfig) (StartResult, error) {
	return machine.Start(startConfig)
}

func Stop(stopConfig StopConfig) (StopResult, error) {
	result := &StopResult{Name: stopConfig.Name}

	result.State, err = machine.Stop(stopConfig.Name, stopConfig.Debug)
	result.Error = err.Error()
	result.Success = (err != nil)

	return result
}

func PowerOff(powerOffConfig PowerOffConfig) (PowerOffResult, error) {
	result := &PowerOffResult{Name: powerOffConfig.Name}

	err := machine.PowerOff(powerOffConfig.Name)
	result.Error = err.Error()
	result.Success = (err != nil)

	return result
}

func Delete(deleteConfig deleteConfig) (deleteResult, error) {
	result := &deleteResult{Name: deleteConfig.Name}

	err := machine.Delete(deleteConfig.Name)
	result.Error = err.Error()
	result.Success = (err != nil)

	return result
}
