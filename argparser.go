package main

import (
	"github.com/alecthomas/kong"
)

// CommandLine is used to define flags when calling the program
type CommandLine struct {
	Config  string `short:"c" help:"Configuration file location" default:"configs/gpmux.yml"`
	Listen  bool   `short:"l" help:"Specify whether to listen as a server rather than connect"`
	Domain  string `short:"d" help:"The ip or domain to use" default:"localhost"`
	Port    uint16 `short:"p" help:"The port to use" default:"14695"`
	Verbose bool   `short:"v" help:"Increase verbosity level"`
}

// Parse the command line arguments
// This should only be called from the main Parse method in this package
func argParse() (cli CommandLine) {
	ctx := kong.Parse(&cli)
	switch ctx.Command() {
	// case "config":
	// 	log.Println("Foo")
	default:
		return
	}
}
