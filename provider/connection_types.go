package provider

const (
	RedshiftServerlessConnectionType = "redshiftserverless"
)

// S3Configuration extends the polytomic.S3Configuration with new fields
type S3Configuration struct {
	AuthMode           string `json:"auth_mode"`
	IAMRoleARN         string `json:"iam_role_arn,omitempty"`
	ExternalID         string `json:"external_id,omitempty"`
	AwsAccessKeyID     string `json:"aws_access_key_id" mapstructure:"aws_access_key_id" tfsdk:"aws_access_key_id"`
	AwsSecretAccessKey string `json:"aws_secret_access_key" mapstructure:"aws_secret_access_key" tfsdk:"aws_secret_access_key"`

	S3BucketRegion string `json:"s3_bucket_region" mapstructure:"s3_bucket_region" tfsdk:"s3_bucket_region"`
	S3BucketName   string `json:"s3_bucket_name" mapstructure:"s3_bucket_name" tfsdk:"s3_bucket_name"`

	IsSingleTable         bool   `json:"is_single_table" mapstructure:"is_single_table" tfsdk:"is_single_table"`
	IsDirectorySnapshot   bool   `json:"is_directory_snapshot" mapstructure:"is_directory_snapshot" tfsdk:"is_directory_snapshot"`
	DirGlobPattern        string `json:"directory_glob_pattern" mapstructure:"directory_glob_pattern" tfsdk:"directory_glob_pattern"`
	SingleTableName       string `json:"single_table_name" mapstructure:"single_table_name" tfsdk:"single_table_name"`
	SingleTableFileFormat string `json:"single_table_file_format" mapstructure:"single_table_file_format" tfsdk:"single_table_file_format"`
	SkipLines             int    `json:"skip_lines" mapstructure:"skip_lines" tfsdk:"skip_lines"`
}

type RedshiftServerlessConnectionConfiguration struct {
	Database  string `json:"database" mapstructure:"database" tfsdk:"database"`
	Workgroup string `json:"workgroup" mapstructure:"workgroup" tfsdk:"workgroup"`
	Region    string `json:"region" mapstructure:"region" tfsdk:"region"`

	IAMRoleARN string `json:"iam_role_arn,omitempty" mapstructure:"iam_role_arn" tfsdk:"iam_role_arn"`
	ExternalID string `json:"external_id,omitempty" mapstructure:"external_id" tfsdk:"external_id"`

	ConnectionMethod   string `json:"connection_method,omitempty" mapstructure:"connection_method" tfsdk:"connection_method"`
	ServerlessEndpoint string `json:"endpoint,omitempty" mapstructure:"endpoint" tfsdk:"serverless_endpoint"`
	OverrideEndpoint   bool   `json:"override_endpoint,omitempty" mapstructure:"override_endpoint" tfsdk:"override_endpoint"`
	DataAPIEndpoint    string `json:"data_api_endpoint,omitempty" mapstructure:"data_api_endpoint" tfsdk:"data_api_endpoint"`

	UseUnload      bool   `json:"use_unload,omitempty" mapstructure:"use_unload" tfsdk:"use_unload"`
	S3BucketName   string `json:"s3_bucket_name" mapstructure:"s3_bucket_name" tfsdk:"s3_bucket_name"`
	S3BucketRegion string `json:"s3_bucket_region" mapstructure:"s3_bucket_region" tfsdk:"s3_bucket_region"`
}
