package main

import (
	_ "embed"
	"flag"
	"fmt"
	"github.com/highpower/inject-hashes/internal/config"
	"github.com/highpower/inject-hashes/pkg/injector"
	"io"
	"log"
	"os"
)

type options struct {
	config  string
	version bool
}

//go:embed version.txt
var version string

func (o *options) Valid() bool {
	return o.version == (o.config == "")
}

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
		stat, err := os.Stat(r.NewLocation)
		switch {
		case os.IsNotExist(err):
			break
		case err != nil:
			return err
		case !stat.Mode().IsRegular():
			return fmt.Errorf("%s is not a regular config", r.NewLocation)
		default:
			continue
		}
		if err := os.Link(r.Location, r.NewLocation); err != nil {
			return err
		}
	}
	return nil
}

func usage(writer io.Writer) {
	_, _ = fmt.Fprintln(writer, "Usage: inject-hashes -config <config> [-verbose]")
	flag.CommandLine.SetOutput(writer)
	flag.CommandLine.PrintDefaults()
}

func main() {

	opts := &options{}
	flag.StringVar(&opts.config, "config", "", "config config to read")
	flag.BoolVar(&opts.version, "version", false, "print version and exit")

	var verbose bool
	flag.BoolVar(&verbose, "verbose", false, "verbose output")

	flag.Usage = func() { usage(os.Stdout) }
	flag.Parse()

	if !flag.Parsed() || !opts.Valid() {
		usage(os.Stderr)
		os.Exit(1)
	}

	if opts.version {
		_, _ = fmt.Fprintf(os.Stdout, "%s\n", version)
		return
	}

	log.Default().SetFlags(0)
	if !verbose {
		log.SetOutput(io.Discard)
	}

	c, err := config.New(opts.config)
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
