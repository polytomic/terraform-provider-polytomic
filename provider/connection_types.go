package provider

const (
	RedshiftServerlessConnectionType = "redshiftserverless"
)

type RedshiftServerlessConnectionConfiguration struct {
	Database  string `json:"database" mapstructure:"database" tfsdk:"database"`
	Workgroup string `json:"workgroup" mapstructure:"workgroup" tfsdk:"workgroup"`
	Region    string `json:"region" mapstructure:"region" tfsdk:"region"`

	IAMRoleARN string `json:"iam_role_arn,omitempty" mapstructure:"iam_role_arn" tfsdk:"iam_role_arn"`
	ExternalID string `json:"external_id,omitempty" mapstructure:"external_id" tfsdk:"external_id"`

	ConnectionMethod   string `json:"connection_method,omitempty" mapstructure:"connection_method" tfsdk:"connection_method"`
	ServerlessEndpoint string `json:"serverless_endpoint,omitempty" mapstructure:"serverless_endpoint" tfsdk:"serverless_endpoint"`
	OverrideEndpoint   bool   `json:"override_endpoint,omitempty" mapstructure:"override_endpoint" tfsdk:"override_endpoint"`
	DataAPIEndpoint    string `json:"data_api_endpoint,omitempty" mapstructure:"data_api_endpoint" tfsdk:"data_api_endpoint"`
}
