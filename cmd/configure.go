package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

type ConfigureCmd struct {
	EnvDir string `help:"Directory to create the environment in. default: './env'" short:"d"`
	Name   string `help:"Name of the environment to create. default: 'default'" short:"n"`
}

type configGenFunc func(cfg EnvironmentConfig, path string) error

func (c *ConfigureCmd) Run() error {
	rc := RuntimeConfigFromFlags(c.EnvDir, c.Name)
	envPath := filepath.Join(rc.EnvDir, rc.EnvName)
	configPath := filepath.Join(envPath, "values.yaml")

	// Load the global configuration
	cfg, err := LoadEnvironmentConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := configureEnv(cfg, envPath); err != nil {
		return fmt.Errorf("failed to configure environment: %w", err)
	}

	log.Info().Msg("Configuration files generated successfully")
	return nil
}

func createLogDirectories(envDir string) error {
	logDirs := []string{
		filepath.Join(envDir, "logs"),
		filepath.Join(envDir, "logs", "cloudserver"),
		filepath.Join(envDir, "logs", "scuba"),
		filepath.Join(envDir, "logs", "backbeat"),
		filepath.Join(envDir, "logs", "migration-tools"),
		filepath.Join(envDir, "logs", "clickhouse-shard-1"),
		filepath.Join(envDir, "logs", "clickhouse-shard-2"),
	}

	for _, dir := range logDirs {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return fmt.Errorf("failed to create log directory %s: %w", dir, err)
		}
	}
	return nil
}

func configureEnv(cfg EnvironmentConfig, envDir string) error {
	log.Info().Msgf("Configuring environment %s", envDir)

	if err := createLogDirectories(envDir); err != nil {
		return fmt.Errorf("failed to create log directories: %w", err)
	}

	if err := generateDefaultsEnv(cfg, envDir); err != nil {
		return fmt.Errorf("failed to generate defaults.env: %w", err)
	}

	components := []configGenFunc{
		generateCloudserverConfig,
		generateBackbeatConfig,
		generateVaultConfig,
		generateScubaConfig,
		generateS3MetadataConfig,
		generateScubaMetadataConfig,
		generateKafkaConfig,
		generateUtapiConfig,
		generateMigrationToolsConfig,
		generateClickhouseConfig,
	}

	configDir := filepath.Join(envDir, "config")

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, component := range components {
		if err := component(cfg, configDir); err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}
	}
	return nil
}

func generateDefaultsEnv(cfg EnvironmentConfig, envDir string) error {
	defaultsEnvPath := filepath.Join(envDir, "defaults.env")
	return renderTemplateToFile(getTemplates(), "templates/global/defaults.env", cfg, defaultsEnvPath)
}

func generateCloudserverConfig(cfg EnvironmentConfig, path string) error {
	version := detectCloudserverVersion(cfg.Cloudserver.Image)

	configTemplate := fmt.Sprintf("templates/cloudserver/config-%s.json", version)

	if f, err := getTemplates().Open(configTemplate); err != nil {
		return fmt.Errorf("no configuration template found for cloudserver version %s (image: %s): %w",
			version, cfg.Cloudserver.Image, err)
	} else {
		if closeErr := f.Close(); closeErr != nil {
			return fmt.Errorf("failed to close template file: %w", closeErr)
		}
	}

	err := renderTemplateToFile(
		getTemplates(),
		configTemplate,
		cfg,
		filepath.Join(path, "cloudserver", "config.json"),
	)
	if err != nil {
		return err
	}

	return renderTemplateToFile(
		getTemplates(),
		"templates/cloudserver/locationConfig.json",
		cfg,
		filepath.Join(path, "cloudserver", "locationConfig.json"),
	)
}

func generateBackbeatConfig(cfg EnvironmentConfig, path string) error {
	templates := []string{
		"env",
		"supervisord.conf",
		"config.json",
		"config.notification.json",
		"notificationCredentials.json",
	}

	return renderTemplates(cfg, "templates/backbeat", filepath.Join(path, "backbeat"), templates)
}

func generateVaultConfig(cfg EnvironmentConfig, path string) error {
	templates := []string{
		"config.json",
		"create-management-account.sh",
		"Dockerfile.setup",
		"management-creds.json",
	}

	return renderTemplates(cfg, "templates/vault", filepath.Join(path, "vault"), templates)
}

func generateScubaConfig(cfg EnvironmentConfig, path string) error {
	templates := []string{
		"config.json",
		"create-service-user.sh",
		"Dockerfile.setup",
		"supervisord.conf",
		"env",
	}
	return renderTemplates(cfg, "templates/scuba", filepath.Join(path, "scuba"), templates)
}

func generateMetadataConfig(cfg MetadataConfig, path string) error {
	return renderTemplateToFile(getTemplates(), "templates/metadata/config.json", cfg, filepath.Join(path, "config.json"))
}

func generateS3MetadataConfig(cfg EnvironmentConfig, path string) error {
	cfgPath := filepath.Join(path, "metadata-s3")
	return generateMetadataConfig(cfg.S3Metadata, cfgPath)
}

func generateScubaMetadataConfig(cfg EnvironmentConfig, path string) error {
	cfgPath := filepath.Join(path, "metadata-scuba")
	return generateMetadataConfig(cfg.ScubaMetadata, cfgPath)
}

func generateKafkaConfig(cfg EnvironmentConfig, path string) error {
	templates := []string{
		"Dockerfile",
		"setup.sh",
		"server.backbeat.properties",
		"server.destination.properties",
		"config.properties",
		"zookeeper.properties",
	}

	return renderTemplates(cfg, "templates/kafka", filepath.Join(path, "kafka"), templates)
}

func generateUtapiConfig(cfg EnvironmentConfig, path string) error {
	return renderTemplateToFile(getTemplates(), "templates/utapi/config.json", cfg, filepath.Join(path, "utapi", "config.json"))
}

func generateMigrationToolsConfig(cfg EnvironmentConfig, path string) error {
	templates := []string{
		"supervisord.conf",
		"migration.yml",
		"env",
	}

	return renderTemplates(cfg, "templates/migration-tools", filepath.Join(path, "migration-tools"), templates)
}

func generateClickhouseConfig(cfg EnvironmentConfig, path string) error {
	templates := []string{
		"Dockerfile.setup",
		"cluster-config.xml",
		"ports-shard-1.xml",
		"ports-shard-2.xml",
		"init-schema.sh",
		"init.d/01-create-database.sql",
		"init.d/02-create-ingest-table.sql",
		"init.d/03-create-storage-table.sql",
		"init.d/04-create-offsets-table.sql",
		"init.d/05-create-distributed-tables.sql",
		"init.d/06-create-materialized-view.sql",
	}

	return renderTemplates(cfg, "templates/clickhouse", filepath.Join(path, "clickhouse"), templates)
}
