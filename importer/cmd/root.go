package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newRootCmd(version string) *cobra.Command {

	rootCmd := &cobra.Command{
		Use:     "polytomic-importer",
		Version: version,
		Short:   "Polytomic importer is a CLI tool to import existing Polytomic resources into Terraform",
	}

	var apiKey, url, partnerKey, deploymentKey string
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Polytomic API key")
	viper.BindPFlag("api-key", rootCmd.PersistentFlags().Lookup("api-key"))
	rootCmd.PersistentFlags().StringVar(&partnerKey, "partner-key", "", "Polytomic partner key (for multi-organization access)")
	viper.BindPFlag("partner-key", rootCmd.PersistentFlags().Lookup("partner-key"))
	rootCmd.PersistentFlags().StringVar(&deploymentKey, "deployment-key", "", "Polytomic deployment key (for multi-organization access)")
	viper.BindPFlag("deployment-key", rootCmd.PersistentFlags().Lookup("deployment-key"))
	rootCmd.PersistentFlags().StringVar(&url, "url", "", "Polytomic API URL")
	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))

	// Run flags
	var output, organizations string
	runCmd.PersistentFlags().StringVar(&output, "output", ".", "Output directory for generated files (defaults to current directory)")
	runCmd.PersistentFlags().StringVar(&organizations, "organizations", "", "Comma-separated list of organization IDs to import (partner-key or deployment-key only)")
	runCmd.PersistentFlags().Bool("replace", false, "Replace existing files")
	runCmd.PersistentFlags().Bool("include-permissions", false, "Include permission resources")
	viper.BindPFlag("output", runCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("organizations", runCmd.PersistentFlags().Lookup("organizations"))
	viper.BindPFlag("replace", runCmd.PersistentFlags().Lookup("replace"))
	viper.BindPFlag("include-permissions", runCmd.PersistentFlags().Lookup("include-permissions"))

	// Register commands
	rootCmd.AddCommand(runCmd)

	// Hide completions
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	return rootCmd
}

func Execute(version string) {
	rootCmd := newRootCmd(version)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
