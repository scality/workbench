package main

import (
	"fmt"
	"os"

	"dario.cat/mergo"
	"gopkg.in/yaml.v3"
)

type RuntimeConfig struct {
	EnvDir  string
	EnvName string
}

const (
	DefaultEnvDir  = "./env"
	DefaultEnvName = "default"
)

func RuntimeConfigFromFlags(envDir, envName string) RuntimeConfig {
	if envDir == "" {
		v := os.Getenv("WORKBENCH_ENV_DIR")
		if v != "" {
			envDir = v
		} else {
			envDir = DefaultEnvDir
		}
	}

	if envName == "" {
		v := os.Getenv("WORKBENCH_ENV_NAME")
		if v != "" {
			envName = v
		} else {
			envName = DefaultEnvName
		}
	}

	return RuntimeConfig{
		EnvDir:  envDir,
		EnvName: envName,
	}
}

type EnvironmentConfig struct {
	Global         GlobalConfig         `yaml:"global"`
	Features       FeatureConfig        `yaml:"features"`
	Cloudserver    CloudserverConfig    `yaml:"cloudserver"`
	S3Metadata     MetadataConfig       `yaml:"s3_metadata"`
	Backbeat       BackbeatConfig       `yaml:"backbeat"`
	Vault          VaultConfig          `yaml:"vault"`
	Scuba          ScubaConfig          `yaml:"scuba"`
	ScubaMetadata  MetadataConfig       `yaml:"scuba_metadata"`
	Kafka          KafkaConfig          `yaml:"kafka"`
	Zookeeper      ZookeeperConfig      `yaml:"zookeeper"`
	Redis          RedisConfig          `yaml:"redis"`
	Utapi          UtapiConfig          `yaml:"utapi"`
	MigrationTools MigrationToolsConfig `yaml:"migration_tools"`
	Clickhouse     ClickhouseConfig     `yaml:"clickhouse"`
	Fluentbit      FluentbitConfig      `yaml:"fluentbit"`
}

type GlobalConfig struct {
	LogLevel string `yaml:"logLevel"`
	// Profile  string `yaml:"profile"`
}

type FeatureConfig struct {
	Scuba                  ScubaFeatureConfig               `yaml:"scuba"`
	BucketNotifications    BucketNotificationsFeatureConfig `yaml:"bucket_notifications"`
	CrossRegionReplication CrossRegionReplicationFeatureConfig `yaml:"cross_region_replication"`
	Utapi                  UtapiFeatureConfig               `yaml:"utapi"`
	Migration              MigrationFeatureConfig           `yaml:"migration"`
	AccessLogging          AccessLoggingFeatureConfig       `yaml:"access_logging"`
}

type ScubaFeatureConfig struct {
	Enabled           bool `yaml:"enabled"`
	EnableServiceUser bool `yaml:"enable_service_user"`
}

type BucketNotificationsFeatureConfig struct {
	Enabled         bool `yaml:"enabled"`
	DestinationAuth struct {
		Type     string `yaml:"type"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"destinationAuth"`
}

type CrossRegionReplicationFeatureConfig struct {
	Enabled bool `yaml:"enabled"`
}

type UtapiFeatureConfig struct {
	Enabled bool `yaml:"enabled"`
}

type MigrationFeatureConfig struct {
	Enabled bool `yaml:"enabled"`
}

type CloudserverConfig struct {
	Image                       string `yaml:"image"`
	EnableNullVersionCompatMode bool   `yaml:"enableNullVersionCompatMode"`
	LogLevel                    string `yaml:"log_level"`
}

type BackbeatConfig struct {
	Image    string `yaml:"image"`
	LogLevel string `yaml:"log_level"`
}

type VaultConfig struct {
	Image    string `yaml:"image"`
	LogLevel string `yaml:"log_level"`
}

type UtapiConfig struct {
	Image    string `yaml:"image"`
	LogLevel string `yaml:"log_level"`
}

type MigrationToolsConfig struct {
	Image    string `yaml:"image"`
	LogLevel string `yaml:"log_level"`
}

type VFormat string

func (vf VFormat) String() string {
	return string(vf)
}

func (vf VFormat) MarshalJSON() ([]byte, error) {
	return []byte(`"` + string(vf) + `"`), nil
}

func (vf *VFormat) UnmarshalJSON(data []byte) error {
	str := string(data)
	switch str {
	case `"v0"`:
		*vf = Formatv0
	case `"v1"`:
		*vf = Formatv1
	default:
		return fmt.Errorf("unknown version format %s", str)
	}
	return nil
}

func (vf VFormat) MarshalYAML() (interface{}, error) {
	return string(vf), nil
}

func (vf *VFormat) UnmarshalYAML(value *yaml.Node) error {
	var str string
	if err := value.Decode(&str); err != nil {
		return err
	}
	switch str {
	case "v0":
		*vf = Formatv0
	case "v1":
		*vf = Formatv1
	default:
		return fmt.Errorf("unknown version format %s", str)
	}
	return nil
}

const (
	Formatv0 VFormat = "v0"
	Formatv1 VFormat = "v1"
)

type MetadataConfig struct {
	Image        string           `yaml:"image"`
	RaftSessions int              `yaml:"raft_sessions"`
	BasePorts    MdPortConfig     `yaml:"base_ports"`
	LogLevel     string           `yaml:"log_level"`
	VFormat      VFormat          `yaml:"vformat"`
	Migration    *MigrationConfig `yaml:"migration"`
}

type MdPortConfig struct {
	Bucketd   uint16 `yaml:"bucketd"`
	Repd      uint16 `yaml:"repd"`
	RepdAdmin uint16 `yaml:"repdAdmin"`
}

type MigrationConfig struct {
	Deploy    bool         `yaml:"deploy"`
	BasePorts MdPortConfig `yaml:"base_ports"`
}

type ScubaConfig struct {
	Image    string `yaml:"image"`
	LogLevel string `yaml:"log_level"`
}

type KafkaConfig struct {
	Image    string `yaml:"image"`
	LogLevel string `yaml:"log_level"`
}

type ZookeeperConfig struct {
	Image    string `yaml:"image"`
	LogLevel string `yaml:"log_level"`
}

type RedisConfig struct {
	Image    string `yaml:"image"`
	LogLevel string `yaml:"log_level"`
}

type ClickhouseConfig struct {
	Image    string `yaml:"image"`
	LogLevel string `yaml:"log_level"`
}

type FluentbitConfig struct {
	Image    string `yaml:"image"`
	LogLevel string `yaml:"log_level"`
}

type AccessLoggingFeatureConfig struct {
	Enabled bool `yaml:"enabled"`
}

func DefaultEnvironmentConfig() EnvironmentConfig {
	return EnvironmentConfig{
		Global: GlobalConfig{
			LogLevel: "info",
			// Profile:  "default",
		},
		Features: FeatureConfig{
			BucketNotifications: BucketNotificationsFeatureConfig{
				DestinationAuth: struct {
					Type     string `yaml:"type"`
					Username string `yaml:"username"`
					Password string `yaml:"password"`
				}{
					Type: "none",
				},
			},
			Utapi: UtapiFeatureConfig{
				Enabled: false,
			},
			Migration: MigrationFeatureConfig{
				Enabled: false,
			},
			CrossRegionReplication: CrossRegionReplicationFeatureConfig{
				Enabled: false,
			},
			AccessLogging: AccessLoggingFeatureConfig{
				Enabled: false,
			},
		},
		Cloudserver: CloudserverConfig{},
		S3Metadata: MetadataConfig{
			VFormat: Formatv1,
			BasePorts: MdPortConfig{
				Bucketd:   9000,
				Repd:      4200,
				RepdAdmin: 4250,
			},
			RaftSessions: 3,
			// LogLevel:     "info",
			Migration: &MigrationConfig{
				Deploy: false,
				BasePorts: MdPortConfig{
					Bucketd:   9001,
					Repd:      4700,
					RepdAdmin: 4750,
				},
			},
		},
		Backbeat: BackbeatConfig{},
		Vault:    VaultConfig{},
		Scuba:    ScubaConfig{},
		ScubaMetadata: MetadataConfig{
			VFormat: Formatv0,
			BasePorts: MdPortConfig{
				Bucketd:   19000,
				Repd:      14200,
				RepdAdmin: 14250,
			},
			RaftSessions: 1,
			// LogLevel:     "info",
		},
		Utapi:          UtapiConfig{},
		MigrationTools: MigrationToolsConfig{},
		Clickhouse:     ClickhouseConfig{},
		Fluentbit:      FluentbitConfig{},
	}
}

func LoadEnvironmentConfig(path string) (EnvironmentConfig, error) {
	cfg := DefaultEnvironmentConfig()

	if path == "" {
		return cfg, nil
	}

	// Read the config file
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the YAML into a temporary config
	var fileCfg EnvironmentConfig
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return cfg, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge the configs, only overriding non-empty fields
	if err := mergo.Merge(&cfg, fileCfg, mergo.WithOverride); err != nil {
		return cfg, fmt.Errorf("failed to merge configs: %w", err)
	}

	// Set the log level for each component that doesn't have one already set
	if cfg.Cloudserver.LogLevel == "" {
		cfg.Cloudserver.LogLevel = cfg.Global.LogLevel
	}

	if cfg.S3Metadata.LogLevel == "" {
		cfg.S3Metadata.LogLevel = cfg.Global.LogLevel
	}
	// deploy the migration metadata map and bucketds if enabled
	if cfg.Features.Migration.Enabled {
		cfg.S3Metadata.Migration.Deploy = true
	}

	if cfg.Backbeat.LogLevel == "" {
		cfg.Backbeat.LogLevel = cfg.Global.LogLevel
	}

	if cfg.Vault.LogLevel == "" {
		cfg.Vault.LogLevel = cfg.Global.LogLevel
	}

	if cfg.Scuba.LogLevel == "" {
		cfg.Scuba.LogLevel = cfg.Global.LogLevel
	}

	if cfg.ScubaMetadata.LogLevel == "" {
		cfg.ScubaMetadata.LogLevel = cfg.Global.LogLevel
	}

	if cfg.Kafka.LogLevel == "" {
		cfg.Kafka.LogLevel = cfg.Global.LogLevel
	}

	if cfg.Zookeeper.LogLevel == "" {
		cfg.Zookeeper.LogLevel = cfg.Global.LogLevel
	}

	if cfg.Redis.LogLevel == "" {
		cfg.Redis.LogLevel = cfg.Global.LogLevel
	}

	if cfg.Utapi.LogLevel == "" {
		cfg.Utapi.LogLevel = cfg.Global.LogLevel
	}

	if cfg.MigrationTools.LogLevel == "" {
		cfg.MigrationTools.LogLevel = cfg.Global.LogLevel
	}

	if cfg.Clickhouse.LogLevel == "" {
		cfg.Clickhouse.LogLevel = cfg.Global.LogLevel
	}

	if cfg.Fluentbit.LogLevel == "" {
		cfg.Fluentbit.LogLevel = cfg.Global.LogLevel
	}

	return cfg, nil
}
