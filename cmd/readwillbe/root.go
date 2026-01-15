package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	port     string
	logLevel string
)

var rootCmd = &cobra.Command{
	Use:   "readwillbe",
	Short: "ReadWillBe - A daily reading plan tracker",
	Long: `ReadWillBe is a GOTH stack application that helps users track
progress through daily reading plans with support for CSV uploads
and manual plan creation.`,
	RunE: runServer,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./readwillbe.yaml)")
	rootCmd.PersistentFlags().StringVar(&port, "port", "", "port to listen on (default is 8080)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "log level (debug, info, warn, error)")

	_ = viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	_ = viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("readwillbe")
	}

	viper.SetEnvPrefix("READWILLBE")
	viper.AutomaticEnv()

	viper.SetDefault("port", "8080")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("db_path", "./tmp/readwillbe.db")
	viper.SetDefault("allow_signup", true)
	viper.SetDefault("seed_db", false)

	// Email configuration defaults
	viper.SetDefault("email_provider", "")
	viper.SetDefault("smtp_host", "")
	viper.SetDefault("smtp_port", 587)
	viper.SetDefault("smtp_username", "")
	viper.SetDefault("smtp_password", "")
	viper.SetDefault("smtp_from", "")
	viper.SetDefault("smtp_tls", "starttls")
	viper.SetDefault("resend_api_key", "")
	viper.SetDefault("resend_from", "")

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
			os.Exit(1)
		}
	}
}
