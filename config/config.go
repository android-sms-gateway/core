package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/s3"
	"github.com/knadh/koanf/v2"
)

var ErrMissingAWSCreds = errors.New("AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION must be set")

// Load reads configuration from various sources and unmarshals it into a given struct.
//
// It looks for configuration in the following order (later overrides earlier):
// 1. S3, if `WithS3YAML` is provided.
// 2. Local file, if `WithLocalYAML` is provided.
// 3. `.env` file in the current working directory.
// 4. Environment variables.
//
// If any of the above sources result in an error (other than `os.ErrNotExist`), it will be returned.
//
// If a source results in `os.ErrNotExist`, it will be skipped.
//
// The final configuration will be unmarshaled into the given struct. If unmarshaling fails, an error will be returned.
func Load[T any](c *T, opts ...Option) error {
	options := new(options)
	options.apply(opts...)

	k := koanf.New(".")

	if err := loadFromS3(options, k); err != nil {
		return err
	}

	if err := loadFromYAML(options.withYaml, k); err != nil {
		return err
	}

	if err := loadDotenv(k); err != nil {
		return err
	}

	if err := loadEnv(k); err != nil {
		return err
	}

	if err := k.Unmarshal("", c); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	return nil
}

func loadFromS3(options *options, k *koanf.Koanf) error {
	if options.withS3Bucket == "" || options.withS3ObjectKey == "" {
		return nil
	}

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")
	endpoint := os.Getenv("AWS_ENDPOINT")

	if accessKey == "" || secretKey == "" || region == "" {
		return ErrMissingAWSCreds
	}

	s3Config := s3.Config{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
		Bucket:    options.withS3Bucket,
		ObjectKey: options.withS3ObjectKey,
		Endpoint:  endpoint,
	}

	err := k.Load(s3.Provider(s3Config), yaml.Parser())
	if err != nil && !strings.Contains(err.Error(), "404") {
		return fmt.Errorf("load s3: %w", err)
	}

	return nil
}

func loadFromYAML(path string, k *koanf.Koanf) error {
	if path == "" {
		return nil
	}

	err := k.Load(file.Provider(path), yaml.Parser())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load yaml: %w", err)
	}

	return nil
}

func loadDotenv(k *koanf.Koanf) error {
	err := k.Load(file.Provider(".env"), dotenv.ParserEnvWithValue("", "__", envTransform))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load dotenv: %w", err)
	}

	return nil
}

func loadEnv(k *koanf.Koanf) error {
	if err := k.Load(env.Provider("__", env.Opt{
		Prefix:        "",
		TransformFunc: envTransform,
		EnvironFunc:   nil,
	}), nil); err != nil {
		return fmt.Errorf("load env: %w", err)
	}

	return nil
}

func envTransform(k, v string) (string, any) {
	k = strings.ToLower(k)
	// JSON object -> map
	if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
		var m map[string]any
		if err := json.Unmarshal([]byte(v), &m); err == nil {
			return k, m
		}
	}
	// JSON array -> []any
	if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
		var a []any
		if err := json.Unmarshal([]byte(v), &a); err == nil {
			return k, a
		}
	}
	// CSV -> []string
	if strings.Contains(v, ",") {
		parts := strings.Split(v, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return k, parts
	}
	return k, v
}
