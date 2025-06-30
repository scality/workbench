package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
)

type ConfigureCmd struct {
	EnvDir string `help:"Directory to create the environment in." required:"" short:"d" default:"./env"`
	Name   string `help:"Name of the environment to create." required:"" short:"n" default:"default"`
}

type configGenFunc func(cfg Config, path string) error

func (c *ConfigureCmd) Run() error {
	envPath := filepath.Join(c.EnvDir, c.Name)
	configPath := filepath.Join(envPath, "config.yaml")

	// Load the global configuration
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	spew.Dump(cfg)

	if err := configureEnv(cfg, envPath); err != nil {
		return fmt.Errorf("failed to configure environment: %w", err)
	}

	log.Info().Msg("Configuration files generated successfully")
	return nil
}

func configureEnv(cfg Config, envDir string) error {
	log.Info().Msgf("Configuring environment %s", envDir)

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

func generateDefaultsEnv(cfg Config, envDir string) error {
	defaultsEnvPath := filepath.Join(envDir, "defaults.env")
	return renderTemplateToFile(getTemplates(), "templates/global/defaults.env", cfg, defaultsEnvPath)
}

func generateCloudserverConfig(cfg Config, path string) error {
	return renderTemplateToFile(getTemplates(), "templates/cloudserver/config.json", cfg, filepath.Join(path, "cloudserver", "config.json"))
}

func generateBackbeatConfig(cfg Config, path string) error {
	templates := []string{
		"Dockerfile.setup",
		"setup.sh",
		"setup-kafka-target.sh",
		"config.notification.json",
		"config.json",
		"supervisord.conf",
		"env",
	}

	for _, tmpl := range templates {
		templatePath := filepath.Join("templates", "backbeat", tmpl)
		outputPath := filepath.Join(path, "backbeat", tmpl)
		if err := renderTemplateToFile(getTemplates(), templatePath, cfg, outputPath); err != nil {
			return fmt.Errorf("failed to render template %s: %w", tmpl, err)
		}
	}
	return nil
}

func generateVaultConfig(cfg Config, path string) error {
	err := renderTemplateToFile(getTemplates(), "templates/vault/config.json", cfg, filepath.Join(path, "vault", "config.json"))
	if err != nil {
		return err
	}

	err = renderTemplateToFile(getTemplates(), "templates/vault/create-management-account.sh", cfg, filepath.Join(path, "vault", "create-management-account.sh"))
	if err != nil {
		return err
	}

	err = renderTemplateToFile(getTemplates(), "templates/vault/Dockerfile.setup", cfg, filepath.Join(path, "vault", "Dockerfile.setup"))
	if err != nil {
		return err
	}

	err = renderTemplateToFile(getTemplates(), "templates/vault/management-creds.json", cfg, filepath.Join(path, "vault", "management-creds.json"))
	if err != nil {
		return err
	}

	return nil
}

func generateScubaConfig(cfg Config, path string) error {
	err := renderTemplateToFile(getTemplates(), "templates/scuba/create-service-user.sh", cfg, filepath.Join(path, "scuba", "create-service-user.sh"))
	if err != nil {
		return err
	}

	err = renderTemplateToFile(getTemplates(), "templates/scuba/Dockerfile.setup", cfg, filepath.Join(path, "scuba", "Dockerfile.setup"))
	if err != nil {
		return err
	}

	return renderTemplateToFile(getTemplates(), "templates/scuba/config.json", cfg, filepath.Join(path, "scuba", "config.json"))
}

func generateMetadataConfig(cfg MetadataConfig, path string) error {
	return renderTemplateToFile(getTemplates(), "templates/metadata/config.json", cfg, filepath.Join(path, "config.json"))
}

func generateS3MetadataConfig(cfg Config, path string) error {
	cfgPath := filepath.Join(path, "metadata-s3")
	return generateMetadataConfig(cfg.S3Metadata, cfgPath)
}

func generateScubaMetadataConfig(cfg Config, path string) error {
	cfgPath := filepath.Join(path, "metadata-scuba")
	return generateMetadataConfig(cfg.ScubaMetadata, cfgPath)
}
