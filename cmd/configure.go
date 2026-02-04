package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

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
		filepath.Join(envDir, "logs", "fluentbit"),
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
		generateFluentbitConfig,
		generateNginxConfig,
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

	templates := []string{
		"locationConfig.json",
		"create-service-user.sh",
		"Dockerfile.setup",
	}

	return renderTemplates(cfg, "templates/cloudserver", filepath.Join(path, "cloudserver"), templates)
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
		"Dockerfile.shard",
		"Dockerfile.setup",
		"entrypoint.sh",
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

func generateFluentbitConfig(cfg EnvironmentConfig, path string) error {
	templates := []string{
		"fluent-bit.conf",
		"parsers.conf",
	}

	return renderTemplates(cfg, "templates/fluentbit", filepath.Join(path, "fluentbit"), templates)
}

func generateNginxConfig(cfg EnvironmentConfig, path string) error {
	if !cfg.Features.S3Frontend.Enabled {
		return nil
	}

	nginxDir := filepath.Join(path, "nginx")

	if err := renderTemplateToFile(
		getTemplates(),
		"templates/nginx/nginx.conf",
		cfg,
		filepath.Join(nginxDir, "nginx.conf"),
	); err != nil {
		return err
	}

	return generateTLSCertificate(nginxDir, "s3-frontend.key", "s3-frontend.crt")
}

// generateTLSCertificate creates a self-signed TLS certificate and key pair.
func generateTLSCertificate(dir, keyName, certName string) error {
	keyPath := filepath.Join(dir, keyName)
	certPath := filepath.Join(dir, certName)

	// Skip if both files already exist
	_, keyErr := os.Stat(keyPath)
	_, certErr := os.Stat(certPath)
	if keyErr == nil && certErr == nil {
		log.Debug().Str("dir", dir).Msg("TLS certificate already exists, skipping generation")
		return nil
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate TLS key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Scality Workbench"},
		},
		DNSNames:              []string{"localhost", "s3.docker.test"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("failed to create TLS certificate: %w", err)
	}

	certFile, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create cert file: %w", err)
	}
	defer func() { _ = certFile.Close() }()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return fmt.Errorf("failed to write TLS certificate: %w", err)
	}

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return fmt.Errorf("failed to marshal TLS key: %w", err)
	}

	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer func() { _ = keyFile.Close() }()

	if err := pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}); err != nil {
		return fmt.Errorf("failed to write TLS key: %w", err)
	}

	log.Info().Str("dir", dir).Msg("Generated self-signed TLS certificate")
	return nil
}
