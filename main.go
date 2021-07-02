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
	Version = "dev build"
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

	var exitCode = cmd.Run(ctx, appName, Version)

	// Exit with the returned exit code
	os.Exit(exitCode)
}
