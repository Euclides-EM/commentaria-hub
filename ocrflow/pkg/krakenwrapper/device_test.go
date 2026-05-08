package krakenwrapper

import (
	"testing"
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
