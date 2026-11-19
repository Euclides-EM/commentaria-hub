package krakenwrapper

import (
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/envexec"
)

func TestAnnotation_ApplyRules(t *testing.T) {
	arg := krakenDeviceArgs()

	if len(arg) != 2 {
		t.Fatalf("unexpected kraken device args length: %d", len(arg))
		return
	}

	if arg[0] != "--device" {
		t.Fatalf("unexpected kraken device arg key: %q", arg[0])
		return
	}

	if arg[1] != "cpu" && arg[1] != "cuda:0" {
		t.Fatalf("unexpected kraken device arg value: %q", arg[1])
		return
	}

}

func TestKrakenDeviceArgMatchesNvidiaSMIPresence(t *testing.T) {
	want := "cpu"
	if envexec.Cmd("nvidia-smi") != nil {
		want = "cuda:0"
	}
	if got := krakenDeviceArg(); got != want {
		t.Fatalf("krakenDeviceArg() = %q, want %q", got, want)
	}
}
