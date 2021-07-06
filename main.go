/*
Copyright © 2018-2021 Neil Hemming
*/
package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/nehemming/oauthproxy/internal/cmd"
)

var (
	// Version is the built version of the software.
	version               = "dev build"
	commit, date, builtBy string
)

func main() {
	// Create main app context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Attach signal handler
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()

	// Main service entrypoint
	appName := strings.ToLower(filepath.Base(os.Args[0]))

	exitCode := cmd.Run(ctx, appName, Version(version, commit, date, builtBy))

	// Exit with the returned exit code
	os.Exit(exitCode)
}

func Version(version, commit, date, builtBy string) string {
	if builtBy != "" {
		builtBy = "[" + builtBy + "]"
	}

	v := version
	for _, p := range []string{commit, date, builtBy} {
		if p != "" {
			v = v + " " + p
		}
	}

	return v
}
