package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/android-sms-gateway/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeTempFile creates a temporary file with the given content and returns its path
func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	return p
}

// withDotEnv creates a .env file with the given content and changes to the directory
func withDotEnv(t *testing.T, dir, content string) {
	t.Helper()
	writeTempFile(t, dir, ".env", content)
	t.Chdir(dir)
}

// TestConfig represents a test configuration structure
type TestConfig struct {
	Database struct {
		Host     string `koanf:"host"`
		Port     int    `koanf:"port"`
		Username string `koanf:"username"`
		Password string `koanf:"password"`
	} `koanf:"database"`

	Server struct {
		Port int `koanf:"port"`
	} `koanf:"server"`

	FeatureFlags map[string]bool `koanf:"feature_flags"`
}

// TestLoadWithNoOptions tests loading configuration with no options (should use .env + env vars)
func TestLoadWithNoOptions(t *testing.T) {
	// Set up environment variables
	t.Setenv("DATABASE__HOST", "localhost")
	t.Setenv("DATABASE__PORT", "5432")
	t.Setenv("SERVER__PORT", "8080")

	// Create a temporary .env file
	envContent := `DATABASE__HOST=env-file-host
DATABASE__PORT=5433
FEATURE_FLAGS={"debug": true}`
	tmpDir := t.TempDir()
	withDotEnv(t, tmpDir, envContent)

	var cfg TestConfig
	err := config.Load(&cfg)
	require.NoError(t, err)

	// Environment variables should override .env file
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.True(t, cfg.FeatureFlags["debug"])
}

// TestLoadWithLocalYAMLOption tests loading configuration with local YAML option
func TestLoadWithLocalYAMLOption(t *testing.T) {
	// Create a temporary YAML file
	yamlContent := `database:
  host: yaml-host
  port: 3306
  username: yaml-user
  password: yaml-pass
server:
  port: 9090
feature_flags:
  debug: true
  new_feature: false`

	tmpDir := t.TempDir()
	yamlFile := writeTempFile(t, tmpDir, "config.yaml", yamlContent)

	var cfg TestConfig
	err := config.Load(&cfg, config.WithLocalYAML(yamlFile))
	require.NoError(t, err)

	assert.Equal(t, "yaml-host", cfg.Database.Host)
	assert.Equal(t, 3306, cfg.Database.Port)
	assert.Equal(t, "yaml-user", cfg.Database.Username)
	assert.Equal(t, "yaml-pass", cfg.Database.Password)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.False(t, cfg.FeatureFlags["new_feature"])
}

// TestLoadWithS3OptionAndMissingCredentials tests loading configuration with S3 option and missing credentials
func TestLoadWithS3OptionAndMissingCredentials(t *testing.T) {
	// Ensure AWS credentials are not set
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")
	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")

	var cfg TestConfig
	err := config.Load(&cfg, config.WithS3YAML("test-bucket", "test-key"))

	require.Error(t, err)
	require.ErrorIs(t, err, config.ErrMissingAWSCreds)
}

// TestYAMLLoadErrorPropagation tests YAML load error propagation (non-ErrNotExist)
func TestYAMLLoadErrorPropagation(t *testing.T) {
	// Create a temporary YAML file with invalid content
	yamlContent := `invalid: yaml: content: [`
	tmpDir := t.TempDir()
	yamlFile := writeTempFile(t, tmpDir, "invalid.yaml", yamlContent)

	var cfg TestConfig
	err := config.Load(&cfg, config.WithLocalYAML(yamlFile))

	require.Error(t, err)
	require.ErrorContains(t, err, "load yaml")
}

// TestConfigurationPrecedence tests configuration precedence (S3 < YAML < .env < env vars)
func TestConfigurationPrecedence(t *testing.T) {
	// Set up environment variables (highest precedence)
	t.Setenv("DATABASE__HOST", "env-host")
	t.Setenv("DATABASE__PORT", "9999")

	// Create a temporary .env file
	envContent := `DATABASE__HOST=env-file-host
DATABASE__PORT=5433
SERVER__PORT=8080`
	tmpDir := t.TempDir()
	withDotEnv(t, tmpDir, envContent)

	// Create a temporary YAML file
	yamlContent := `database:
  host: yaml-host
  port: 3306
server:
  port: 9090`
	yamlFile := writeTempFile(t, tmpDir, "config.yaml", yamlContent)

	var cfg TestConfig
	err := config.Load(&cfg, config.WithLocalYAML(yamlFile))
	require.NoError(t, err)

	// Environment variables should have highest precedence
	assert.Equal(t, "env-host", cfg.Database.Host)
	assert.Equal(t, 9999, cfg.Database.Port)
	assert.Equal(t, 8080, cfg.Server.Port)
}

// TestUnmarshalingErrorHandling tests unmarshaling error handling
func TestUnmarshalingErrorHandling(t *testing.T) {
	// Create a temporary YAML file with invalid structure for our TestConfig
	// Using a port value that can't be converted to int
	yamlContent := `database:
  host: yaml-host
  port: invalid_port_number
  username: yaml-user
  password: yaml-pass
server:
  port: 9090`
	tmpDir := t.TempDir()
	yamlFile := writeTempFile(t, tmpDir, "invalid.yaml", yamlContent)

	var cfg TestConfig
	err := config.Load(&cfg, config.WithLocalYAML(yamlFile))

	// Should get an error due to invalid port conversion
	require.Error(t, err)
	require.ErrorContains(t, err, "unmarshal")
}

// TestOptionConstructors tests option constructors
func TestOptionConstructors(t *testing.T) {
	tests := []struct {
		name string
		fn   func() config.Option
	}{
		{"WithLocalYAML", func() config.Option { return config.WithLocalYAML("/path/to/config.yaml") }},
		{"WithS3YAML", func() config.Option { return config.WithS3YAML("test-bucket", "config.yaml") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.fn())
		})
	}
}

// TestEnvironmentVariableTransformation tests environment variable transformation
func TestEnvironmentVariableTransformation(t *testing.T) {
	// Set up environment variables with underscores
	t.Setenv("DATABASE__HOST", "test-host")
	t.Setenv("DATABASE__PORT", "5432")
	t.Setenv("SERVER__PORT", "8080")

	// Create a temporary .env file
	envContent := `DATABASE__HOST=env-file-host
DATABASE__PORT=5433`
	tmpDir := t.TempDir()
	withDotEnv(t, tmpDir, envContent)

	var cfg TestConfig
	err := config.Load(&cfg)
	require.NoError(t, err)

	// Environment variables should be transformed to lowercase and override .env
	assert.Equal(t, "test-host", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, 8080, cfg.Server.Port)
}

// TestDotEnvFileLoading tests .env file loading with custom parser
func TestDotEnvFileLoading(t *testing.T) {
	// Create a temporary .env file with custom format
	envContent := `# This is a comment
DATABASE__HOST=env-host
DATABASE__PORT=5432
# Another comment
SERVER__PORT=8080
FEATURE_FLAGS={"debug": true, "new_feature": false}`

	tmpDir := t.TempDir()
	withDotEnv(t, tmpDir, envContent)

	var cfg TestConfig
	err := config.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "env-host", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.True(t, cfg.FeatureFlags["debug"])
	assert.False(t, cfg.FeatureFlags["new_feature"])
}

// TestLoadWithNonExistentFile tests loading with non-existent files (should not error)
func TestLoadWithNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	var cfg TestConfig
	err := config.Load(&cfg, config.WithLocalYAML(filepath.Join(tmpDir, "does-not-exist.yaml")))

	require.NoError(t, err)
	// Should load default zero values
	assert.Empty(t, cfg.Database.Host)
	assert.Equal(t, 0, cfg.Database.Port)
	assert.Equal(t, 0, cfg.Server.Port)
}

// TestYAMLPlusEnvPrecedence tests precedence of YAML and environment variables
func TestYAMLPlusEnvPrecedence(t *testing.T) {
	// Set up environment variables
	t.Setenv("SERVER__PORT", "9999")

	// Create a temporary YAML file
	yamlContent := `database:
  host: yaml-host
  port: 3306
server:
  port: 9090`
	tmpDir := t.TempDir()
	yamlFile := writeTempFile(t, tmpDir, "config.yaml", yamlContent)

	var cfg TestConfig
	err := config.Load(&cfg, config.WithLocalYAML(yamlFile))
	require.NoError(t, err)

	// Environment variable should override YAML
	assert.Equal(t, "yaml-host", cfg.Database.Host)
	assert.Equal(t, 3306, cfg.Database.Port)
	assert.Equal(t, 9999, cfg.Server.Port) // From environment
}

// TestLoadWithValidAWSCredentials tests that AWS credentials are properly validated
func TestLoadWithValidAWSCredentials(t *testing.T) {
	// Set up valid AWS credentials
	t.Setenv("AWS_ACCESS_KEY_ID", "test-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret")
	t.Setenv("AWS_REGION", "us-east-1")

	// Test that the function doesn't return ErrMissingAWSCreds when credentials are present
	var cfg TestConfig
	err := config.Load(&cfg, config.WithS3YAML("test-bucket", "test-key"))

	// We expect this to fail for other reasons (like S3 connection), but not due to missing credentials
	require.NotErrorIs(t, err, config.ErrMissingAWSCreds)
}

// TestLoadWithPartialAWSCredentials tests that partial AWS credentials are rejected
func TestLoadWithPartialAWSCredentials(t *testing.T) {
	// Set up partial AWS credentials (missing one)
	t.Setenv("AWS_ACCESS_KEY_ID", "test-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret")
	// AWS_REGION is missing

	var cfg TestConfig
	err := config.Load(&cfg, config.WithS3YAML("test-bucket", "test-key"))

	require.Error(t, err)
	require.ErrorIs(t, err, config.ErrMissingAWSCreds)
}

// TestLoadWithEmptyAWSCredentials tests that empty AWS credentials are rejected
func TestLoadWithEmptyAWSCredentials(t *testing.T) {
	// Set up empty AWS credentials
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")
	t.Setenv("AWS_REGION", "")

	var cfg TestConfig
	err := config.Load(&cfg, config.WithS3YAML("test-bucket", "test-key"))

	require.Error(t, err)
	require.ErrorIs(t, err, config.ErrMissingAWSCreds)
}
