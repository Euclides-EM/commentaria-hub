package model

import (
	"strings"
)

type Edition struct {
	Meta
	Facsimiles []*Facsimile `json:"facsimiles,omitempty"`
}

type EditionExpandOptions string

const (
	EditionExpandFacsimiles EditionExpandOptions = "facsimiles"
)

func ToEditionExpandOptions(s string) []EditionExpandOptions {
	var opts []EditionExpandOptions
	for _, candidate := range strings.Split(s, ",") {
		switch EditionExpandOptions(candidate) {
		case EditionExpandFacsimiles:
			opts = append(opts, EditionExpandFacsimiles)
		}
	}
	return opts
}

type EditionOrderByOptions string

const (
	EditionOrderBySuggested EditionOrderByOptions = "suggested"
)

func ToEditionOrderByOptions(s string) []EditionOrderByOptions {
	var opts []EditionOrderByOptions
	for _, candidate := range strings.Split(s, ",") {
		switch EditionOrderByOptions(candidate) {
		case EditionOrderBySuggested:
			opts = append(opts, EditionOrderBySuggested)
		}
	}
	return opts
}
