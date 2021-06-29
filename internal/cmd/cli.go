/*
Copyright © 2018-2021 Neil Hemming
*/

//Package cmd provides the command line interface to oauthproxy
package cmd

import (
	"context"
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// ExitCodeSuccess indicates a successful exit
	ExitCodeSuccess = 0

	// ExitCodeError indicates a non successful process exit
	ExitCodeError = 1

	envPrefix  = "OAP"
	flagConfig = "config"
)

type (
	cli struct {
		appName    string
		rootCmd    *cobra.Command
		configFile string
		ctx        context.Context
	}
)

// Run executes the command line interface to the app.  The passed ctx is used to cancel long running tasks.
// appName is the name of the application and forms the suffix of the dot config file
func Run(ctx context.Context, appName string) int {

	cli := &cli{
		appName: appName,
		rootCmd: &cobra.Command{
			Use:           appName,
			Short:         "oauth2 token proxy",
			Long:          "Provides a oauth2 token proxy, designed to reduce load on the downstream authentication provider",
			Args:          cobra.NoArgs,
			SilenceErrors: true,
		},
		ctx: ctx,
	}

	serverCmd := &cobra.Command{
		Use:           "serve",
		Short:         "run the oauth2 token proxy server",
		Long:          "runs the auth2 token proxy server",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE:          cli.runServerCmd,
	}

	requestCmd := &cobra.Command{
		Use:           "request (secretsfile)",
		Short:         "request a oauth2 token from a server",
		Long:          "request a auth2 token from a server",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		RunE:          cli.requestTokenCmd,
	}

	cli.rootCmd.PersistentFlags().StringVar(&cli.configFile, flagConfig, "",
		fmt.Sprintf("specify a configuration file (default is ./%s)", cli.appName))

	cli.rootCmd.AddCommand(serverCmd)
	cli.rootCmd.AddCommand(requestCmd)

	cli.bindServeFlagsAndConfig(serverCmd)

	// Register the config hook, until svr.rootCmd.Execute() is in progress
	// the flags will not have been read.
	cobra.OnInitialize(cli.initConfig)

	// Execute the root command
	if err := cli.rootCmd.Execute(); err != nil {
		log.Error(err)
		return ExitCodeError
	}

	// Exit with success
	return ExitCodeSuccess
}

// initConfig is called during the cobra start up process to init the config settings
func (cli *cli) initConfig() {

	// Establish logging
	isCustomConfig := false
	viper.SetConfigType("yaml")

	if cli.configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cli.configFile)
		isCustomConfig = true
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Error(err)
			os.Exit(ExitCodeError)
		}

		// Search config in home directory with name ".(appName)" (without extension).
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigName("." + cli.appName)
	}

	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	cfgName := viper.ConfigFileUsed()

	if isCustomConfig && err != nil {
		log.Error(err)
		os.Exit(ExitCodeError)
	} else if cfgName != "" {
		log.Println(fmt.Sprintf("using config %s", viper.ConfigFileUsed()))
	}
}
