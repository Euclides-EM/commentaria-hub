package krakenwrapper

import (
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/envexec"
	"github.com/samber/lo"
)

func krakenDeviceArg() string {
	return lo.Ternary(envexec.Cmd("nvidia-smi") != nil, "cpu", "cuda:0")
}

func krakenDeviceArgs() []string {
	return []string{"--device", krakenDeviceArg()}
}
