package main

import (
	"fmt"

	"readwillbe/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ReadWillBe",
	Long:  `All software has versions. This is ReadWillBe's`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("ReadWillBe %s\n", version.Tag)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
