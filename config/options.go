package config

type options struct {
	withYaml string

	withS3Bucket    string
	withS3ObjectKey string
}

type Option func(*options)

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithLocalYAML specifies a path to a local YAML file to load config from.
// If the file does not exist, an error is not returned.
func WithLocalYAML(path string) Option {
	return func(o *options) {
		o.withYaml = path
	}
}

// WithS3YAML configures loading config from S3 (s3://<bucket>/<objectKey>).
// If the object is missing, it is skipped without returning an error.
func WithS3YAML(bucket string, objectKey string) Option {
	return func(o *options) {
		o.withS3Bucket = bucket
		o.withS3ObjectKey = objectKey
	}
}
