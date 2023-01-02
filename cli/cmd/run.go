package cmd

import (
	"github.com/polytomic/terraform-provider-polytomic/importer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the Polytomic importer",
	Long:  `Export existing Polytomic resources into Terraform by creating the necessary *.tf files and an associated import.sh script.`,
	Run: func(cmd *cobra.Command, args []string) {
		url := viper.GetString("url")
		apiKey := viper.GetString("api-key")
		path := viper.GetString("output")
		replace := viper.GetBool("replace")

		importer.Init(url, apiKey, path, replace)
	},
}
