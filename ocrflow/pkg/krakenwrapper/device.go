package krakenwrapper

import (
	"log"
	"os"
	"strings"
)

const krakenDeviceEnvVar = "OCRFLOW_KRAKEN_DEVICE"

func krakenDeviceArg() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(krakenDeviceEnvVar))) {
	case "", "cpu":
		return "cpu"
	case "gpu":
		return "cuda:0"
	default:
		log.Printf("Invalid %s value %q, defaulting to cpu", krakenDeviceEnvVar, os.Getenv(krakenDeviceEnvVar))
		return "cpu"
	}
}

func krakenDeviceArgs() []string {
	return []string{"--device", krakenDeviceArg()}
}
