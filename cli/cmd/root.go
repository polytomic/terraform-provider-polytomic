package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {

	var apiKey string
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Polytomic API key")
	viper.BindPFlag("api-key", rootCmd.PersistentFlags().Lookup("api-key"))
	var url string
	rootCmd.PersistentFlags().StringVar(&url, "url", "https://api.polytomic.com", "Polytomic API URL")
	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))

	// Persistent flags
	var output string
	runCmd.PersistentFlags().StringVar(&output, "output", ".", "Output directory for generated files (defaults to current directory)")
	runCmd.PersistentFlags().Bool("replace", false, "Replace existing files")
	viper.BindPFlag("output", runCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("replace", runCmd.PersistentFlags().Lookup("replace"))

	rootCmd.AddCommand(runCmd)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}

var rootCmd = &cobra.Command{
	Use:   "polytomic-importer",
	Short: "Polytomic importer is a CLI tool to import existing Polytomic resources into Terraform",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
