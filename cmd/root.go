package cmd

import (
	"fmt"
	"github.com/amoydavid/godeployer/pkg/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "deployer",
	Short: "Deployer is a tool for deploying applications",
	Long: `A flexible and powerful deployment tool written in Go,
           inspired by PHP's Deployer.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of godeployer",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.GetVersionInfo())
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
