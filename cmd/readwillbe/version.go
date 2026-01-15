package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"readwillbe/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ReadWillBe",
	Long:  `All software has versions. This is ReadWillBe's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ReadWillBe %s\n", version.Tag)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
