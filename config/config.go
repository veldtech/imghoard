package config

type Config struct {
	BaseURL string			`json:"base_url"`

	DatabaseUser string		`json:"database_user"`
	DatabasePass string		`json:"database_pass"`
	DatabaseHost string		`json:"database_host"`
	DatabaseSchema string	`json:"database_schema"`
	DatabaseUseSSL bool		`json:"database_use_ssl"`

	S3AccessKey string		`json:"s3_access"`
	S3SecretKey string		`json:"s3_secret"`
	S3Endpoint string		`json:"s3_endpoint"`
	S3Folder string			`json:"s3_folder"`
	S3Bucket string			`json:"s3_bucket"`

	SentryDsn string		`json:"sentry_dsn"`
}