package cmd

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

	var apiKey, url string
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Polytomic API key")
	viper.BindPFlag("api-key", rootCmd.PersistentFlags().Lookup("api-key"))
	rootCmd.PersistentFlags().StringVar(&url, "url", "app.polytomic.com", "Polytomic API URL")
	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))

	// Run flags
	var output string
	runCmd.PersistentFlags().StringVar(&output, "output", ".", "Output directory for generated files (defaults to current directory)")
	runCmd.PersistentFlags().Bool("replace", false, "Replace existing files")
	runCmd.PersistentFlags().Bool("include-permissions", false, "Include permission resources")
	runCmd.PersistentFlags().Bool("with-api-key", false, "Include API key in provider block")
	viper.BindPFlag("output", runCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("replace", runCmd.PersistentFlags().Lookup("replace"))
	viper.BindPFlag("include-permissions", runCmd.PersistentFlags().Lookup("include-permissions"))
	viper.BindPFlag("with-api-key", runCmd.PersistentFlags().Lookup("with-api-key"))

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
