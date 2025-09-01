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

type Config struct {
	Global        GlobalConfig      `yaml:"global"`
	Features      FeatureConfig     `yaml:"features"`
	Cloudserver   CloudserverConfig `yaml:"cloudserver"`
	S3Metadata    MetadataConfig    `yaml:"s3_metadata"`
	Backbeat      BackbeatConfig    `yaml:"backbeat"`
	Vault         VaultConfig       `yaml:"vault"`
	Scuba         ScubaConfig       `yaml:"scuba"`
	ScubaMetadata MetadataConfig    `yaml:"scuba_metadata"`
	Kafka         KafkaConfig       `yaml:"kafka"`
	Zookeeper     ZookeeperConfig   `yaml:"zookeeper"`
	Redis         RedisConfig       `yaml:"redis"`
	Utapi         UtapiConfig       `yaml:"utapi"`
}

type GlobalConfig struct {
	LogLevel string `yaml:"logLevel"`
	// Profile  string `yaml:"profile"`
}

type FeatureConfig struct {
	Scuba               ScubaFeatureConfig               `yaml:"scuba"`
	BucketNotifications BucketNotificationsFeatureConfig `yaml:"bucket_notifications"`
	Utapi               UtapiFeatureConfig               `yaml:"utapi"`
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

type UtapiFeatureConfig struct {
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
	Image        string       `yaml:"image"`
	RaftSessions int          `yaml:"raft_sessions"`
	BasePorts    MdPortConfig `yaml:"base_ports"`
	LogLevel     string       `yaml:"log_level"`
	VFormat      VFormat      `yaml:"vformat"`
}

type MdPortConfig struct {
	Bucketd   uint16 `yaml:"bucketd"`
	Repd      uint16 `yaml:"repd"`
	RepdAdmin uint16 `yaml:"repdAdmin"`
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

func DefaultConfig() Config {
	return Config{
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
		Utapi: UtapiConfig{},
	}
}

func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		return cfg, nil
	}

	// Read the config file
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the YAML into a temporary config
	var fileCfg Config
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

	return cfg, nil
}
