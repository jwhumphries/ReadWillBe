package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Validate and display the current configuration",
	Long:  `Shows the merged configuration from defaults, config file, environment variables, and flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Current Configuration:")
		fmt.Println("======================")

		if viper.ConfigFileUsed() != "" {
			fmt.Printf("Config file: %s\n", viper.ConfigFileUsed())
		} else {
			fmt.Println("Config file: none (using defaults and environment variables)")
		}

		fmt.Println()
		fmt.Printf("  port:         %s\n", viper.GetString("port"))
		fmt.Printf("  log_level:    %s\n", viper.GetString("log_level"))
		fmt.Printf("  db_path:      %s\n", viper.GetString("db_path"))
		fmt.Printf("  allow_signup: %t\n", viper.GetBool("allow_signup"))
		fmt.Printf("  seed_db:      %t\n", viper.GetBool("seed_db"))

		if viper.IsSet("tz") {
			fmt.Printf("  tz:           %s\n", viper.GetString("tz"))
		}

		if viper.IsSet("cookie_secret") {
			fmt.Printf("  cookie_secret: [REDACTED]\n")
		} else {
			fmt.Printf("  cookie_secret: [NOT SET - REQUIRED]\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
