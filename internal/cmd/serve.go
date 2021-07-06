/*
Copyright © 2018-2021 Neil Hemming
*/

package cmd

import (
	"time"

	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/oauthproxy/internal/proxy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagEndpoint = "downstream"
	flagPort     = "port"
	flagSilent   = "silent"

	cfgEndpoint = "serve.downstream"
	cfgPort     = "serve.port"
	cfgCacheTTL = "serve.cacheTTL"
	cfgTimeout  = "serve.timeout"
	cfgShutdown = "serve.shutdown"
	cfgSilent   = "serve.silent"
	cfgPoolSize = "serve.poolSize"
)

func (cli *cli) runServerCmd(cmd *cobra.Command, args []string) error {
	// Reaching this stage we can silence errors generating usage
	cmd.SilenceUsage = true

	// coonfigure the proxy settings based off defaults and viper settings
	settings := configureSettings(proxy.DefaultSettings())

	// Run the service and return any errors
	return proxy.Run(cli.ctx, settings)
}

func (cli *cli) bindServeFlagsAndConfig(cmd *cobra.Command) {
	pf := cmd.PersistentFlags()

	pf.String(flagEndpoint, "", "downstream url")
	_ = viper.BindPFlag(cfgEndpoint, pf.Lookup(flagEndpoint))

	pf.Uint(flagPort, 8090, "port proxy listening on")
	_ = viper.BindPFlag(cfgPort, pf.Lookup(flagPort))
	viper.SetDefault(cfgPort, 8090)

	viper.SetDefault(cfgCacheTTL, 15)
	viper.SetDefault(cfgShutdown, 10)
	viper.SetDefault(cfgTimeout, 30)
	viper.SetDefault(cfgPoolSize, 2)

	pf.Bool(flagSilent, false, "silence all output logging")
	_ = viper.BindPFlag(cfgSilent, pf.Lookup(flagSilent))
	viper.SetDefault(cfgSilent, false)
}

// configureSettings configures the applications settings.
func configureSettings(settings proxy.Settings) proxy.Settings {
	// Add in the settings
	endpoint := viper.GetString(cfgEndpoint)
	port := viper.GetUint(cfgPort)

	settings.CacheTTL = time.Duration(viper.GetUint64(cfgCacheTTL)) * time.Minute
	settings.ShutdownGracePeriod = time.Duration(viper.GetUint64(cfgShutdown)) * time.Second
	settings.RequestTimeout = time.Duration(viper.GetUint64(cfgTimeout)) * time.Second
	settings.PoolSize = viper.GetInt(cfgPoolSize)

	var logger proxy.LoggerFunc

	if !viper.GetBool(cfgSilent) {
		logger = func(isError bool, format string, args ...interface{}) {
			if isError {
				loggee.Errorf(format, args...)
			} else {
				loggee.Infof(format, args...)
			}
		}
	}

	return settings.
		WithEndpoint(endpoint).
		WithLogger(logger).
		WithHTTPPort(port)
}
