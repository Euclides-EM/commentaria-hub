package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const supportedWorkflows = `Supported conversions:
  - pdf -> png
  - image -> png
  - image -> pdf
  - pagexml -> alto
  - alto -> markdown`

func main() {
	config, err := parseConfig(os.Args[1:], os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	if err := run(config); err != nil {
		log.Fatal(err)
	}
}

func parseConfig(args []string, output io.Writer) (Config, error) {
	var config Config
	flags := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ContinueOnError)
	flags.SetOutput(output)
	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage: %s -from FORMAT -to FORMAT -input PATH -output PATH [options]\n\n%s\n\nOptions:\n", flags.Name(), supportedWorkflows)
		flags.PrintDefaults()
	}
	flags.StringVar(&config.From, "from", "", "input format")
	flags.StringVar(&config.To, "to", "", "output format")
	flags.StringVar(&config.InputPath, "input", "", "input file or flat directory")
	flags.StringVar(&config.OutputPath, "output", "", "output file or directory, as required by the selected conversion")
	flags.StringVar(&config.PageRange, "range", "", "PDF pages to process (for example, 1,3-5)")
	flags.Float64Var(&config.DPI, "dpi", 300, "resolution used when rendering PDF pages")
	if err := flags.Parse(args); err != nil {
		return Config{}, err
	}
	if flags.NArg() != 0 {
		return Config{}, fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	return config, nil
}
