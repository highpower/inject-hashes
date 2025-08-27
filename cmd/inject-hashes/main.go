package main

import (
	"flag"
	"fmt"
	"inject-hashes/internal/config"
	"inject-hashes/pkg/injector"
	"io"
	"log"
	"os"
)

func run(c *config.Data) error {
	rules := ([]injector.Rule)(nil)
	for _, f := range c.Files {
		r, err := injector.NewRule(f.Ref, f.Location)
		if err != nil {
			return err
		}
		rules = append(rules, r)
	}
	if err := createFiles(rules); err != nil {
		return err
	}
	return injector.Walk(c.Mask, c.Dirs, rules)
}

func createFiles(rules []injector.Rule) error {
	for _, r := range rules {
		if err := os.Rename(r.Location, r.NewLocation); err != nil {
			return err
		}
	}
	return nil
}

func usage(writer io.Writer) {
	_, _ = fmt.Fprintln(writer, "Usage: inject-hashes -config <file> [-verbose]")
	flag.CommandLine.SetOutput(writer)
	flag.CommandLine.PrintDefaults()
}

func main() {

	var file string
	flag.StringVar(&file, "config", "", "config file to read")

	var verbose bool
	flag.BoolVar(&verbose, "verbose", false, "verbose output")

	flag.Usage = func() { usage(os.Stdout) }
	flag.Parse()

	if !flag.Parsed() || file == "" {
		usage(os.Stderr)
		os.Exit(1)
	}

	log.Default().SetFlags(0)
	if !verbose {
		log.SetOutput(io.Discard)
	}

	c, err := config.New(file)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	log.Printf("%s", c.String())
	if err := run(c); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
