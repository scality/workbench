package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

type CreateEnvCmd struct {
	EnvDir    string `help:"Directory to create the environment in." required:"" short:"d" default:"./env"`
	Name      string `help:"Name of the environment to create." required:"" short:"n" default:"default"`
	Overwrite bool   `help:"Overwrite the environment if it already exists." short:"o"`
}

func (c *CreateEnvCmd) Run() error {
	envPath, err := createEnv(c.EnvDir, c.Name, c.Overwrite)
	if err != nil {
		return fmt.Errorf("failed to create environment: %w", err)
	}

	cfg, err := LoadConfig(filepath.Join(envPath, "config.yaml"))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := configureEnv(cfg, envPath); err != nil {
		return fmt.Errorf("failed to configure environment: %w", err)
	}

	return nil
}

func createEnv(envDir string, name string, overwrite bool) (string, error) {
	log.Info().Msgf("Creating new environment %s", name)

	// Check if envDir exists, if not create it
	if info, err := os.Stat(envDir); err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("%s exists but is not a directory", envDir)
		}
	} else if err := os.MkdirAll(envDir, 0755); err != nil {
		return "", err
	}

	// Check if envDir/Name exists, if not create it
	envPath := filepath.Join(envDir, name)
	if info, err := os.Stat(envPath); err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("%s exists but is not a directory", envPath)
		}
	} else if err := os.MkdirAll(envPath, 0755); err != nil {
		return "", err
	}

	// Template the global config if it doesn't exist
	configPath := filepath.Join(envPath, "config.yaml")
	_, err := os.Stat(configPath)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check config file: %w", err)
	}
	if os.IsNotExist(err) || overwrite {
		err := renderTemplateToFile(getTemplates(), "templates/global/config.yaml", nil, configPath)
		if err != nil {
			return "", err
		}
	}

	// Template the docker-compose.yaml file
	dockerComposePath := filepath.Join(envPath, "docker-compose.yaml")
	_, err = os.Stat(dockerComposePath)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check docker-compose file: %w", err)
	}

	if os.IsNotExist(err) || overwrite {
		err := renderTemplateToFile(getTemplates(), "templates/global/docker-compose.yaml", nil, dockerComposePath)
		if err != nil {
			return "", err
		}
	}

	return envPath, nil
}
