package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "nexusclaw",
	Short: "NexusClaw - MCP Gateway Platform",
	Long:  "NexusClaw is a unified gateway for MCP server management, credential vaulting, and AI firewall.",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./nexusclaw.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("server-url", "http://localhost:8080", "NexusClaw server URL")
	rootCmd.PersistentFlags().String("token", "", "authentication token")
	viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("server.url", rootCmd.PersistentFlags().Lookup("server-url"))
	viper.BindPFlag("auth.token", rootCmd.PersistentFlags().Lookup("token"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("nexusclaw")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}
	viper.AutomaticEnv()
	viper.ReadInConfig() // ignore error if no config file
}
