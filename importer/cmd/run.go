package main

import (
	"github.com/polytomic/terraform-provider-polytomic/importer"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the Polytomic importer",
	Long:  `Export existing Polytomic resources into Terraform by creating the necessary *.tf files and an associated import.sh script.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		url := viper.GetString("url")
		apiKey := viper.GetString("api-key")
		partnerKey := viper.GetString("partner-key")
		deploymentKey := viper.GetString("deployment-key")
		organizations := viper.GetString("organizations")
		path := viper.GetString("output")
		replace := viper.GetBool("replace")
		includePermissions := viper.GetBool("include-permissions")

		if apiKey == "" && partnerKey == "" && deploymentKey == "" {
			log.Fatal().Msg("either --api-key, --partner-key, or --deployment-key must be provided")
		}

		clientOpts := providerclient.OptionsFromEnv()
		if url != "" {
			clientOpts.DeploymentURL = url
		}
		if apiKey != "" {
			clientOpts.APIKey = apiKey
		}
		if partnerKey != "" {
			clientOpts.PartnerKey = partnerKey
		}
		if deploymentKey != "" {
			clientOpts.DeploymentKey = deploymentKey
		}

		clientProvider, err := providerclient.NewClientProvider(clientOpts)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create client provider")
		}
		importer.Init(ctx, clientProvider, organizations, path, replace, includePermissions)
	},
}
