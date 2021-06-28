/*
Copyright © 2018-2021 Neil Hemming
*/
package cmd

import (
	"fmt"
	"time"

	"github.com/nehemming/oauthproxy/internal/proxy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	FlagEndpoint = "downstream"
	FlagPort     = "port"
	FlagSilent   = "silent"

	CfgEndpoint = "serve.downstream"
	CfgPort     = "serve.port"
	CfgCacheTtl = "serve.cacheTTL"
	CfgTimeout  = "serve.timeout"
	CfgShutdown = "serve.shutdown"
	CfgSilent   = "serve.silent"
	CfgPoolSize = "serve.poolSize"
)

func (cli *cli) runServerCmd(cmd *cobra.Command, args []string) error {

	// Reaching this stage we can silence errors generating usage
	cmd.SilenceUsage = true

	// coonfigure the proxy settings based off defaults and viper settings
	settings, err := configureSettings(proxy.DefaultSettings())
	if err != nil {
		return err
	}

	// Run the service and return any errors
	return proxy.Run(cli.ctx, settings)
}

func (cli *cli) bindServeFlagsAndConfig(cmd *cobra.Command) {
	pf := cmd.PersistentFlags()

	pf.String(FlagEndpoint, "", "downstream url")
	viper.BindPFlag(CfgEndpoint, pf.Lookup(FlagEndpoint))

	pf.Uint(FlagPort, 8090, "port proxy listening on")
	viper.BindPFlag(CfgPort, pf.Lookup(FlagPort))
	viper.SetDefault(CfgPort, 8090)

	viper.SetDefault(CfgCacheTtl, 15)
	viper.SetDefault(CfgShutdown, 10)
	viper.SetDefault(CfgTimeout, 30)
	viper.SetDefault(CfgPoolSize, 2)

	pf.Bool(FlagSilent, false, "silence all output logging")
	viper.BindPFlag(CfgSilent, pf.Lookup(FlagSilent))
	viper.SetDefault(CfgSilent, false)
}

// configureSettings configures the applications settings
func configureSettings(settings proxy.Settings) (proxy.Settings, error) {

	//	Add in the settings
	endpoint := viper.GetString(CfgEndpoint)
	port := viper.GetUint(CfgPort)

	settings.CacheTTL = time.Duration(viper.GetUint(CfgCacheTtl) * uint(time.Minute))
	settings.ShutdownGracePeriod = time.Duration(viper.GetUint(CfgShutdown) * uint(time.Second))
	settings.RequestTimeout = time.Duration(viper.GetUint(CfgTimeout) * uint(time.Second))
	settings.PoolSize = viper.GetInt(CfgPoolSize)

	var logger proxy.LoggerFunc

	if !viper.GetBool(CfgSilent) {
		logger = func(isError bool, format string, args ...interface{}) {
			if isError {
				log.Errorln(fmt.Sprintf(format, args...))
			} else {
				log.Println(fmt.Sprintf(format, args...))
			}
		}
	}

	return settings.
		WithEndpoint(endpoint).
		WithLogger(logger).
		WithHttpPort(port), nil
}
