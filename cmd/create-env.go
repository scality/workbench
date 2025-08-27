package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

type CreateEnvCmd struct {
	EnvDir            string `help:"Directory to create the environment in." required:"" short:"d" default:"./env"`
	Name              string `help:"Name of the environment to create." required:"" short:"n" default:"default"`
	Overwrite         bool   `help:"Overwrite the environment if it already exists." short:"o"`
	WithConfig        string `help:"Path to a custom configuration file. Replaces the default config." type:"existingfile"`
	WithDockerCompose string `help:"Path to a custom Docker Compose file. Replaces the default file." type:"existingfile"`
}

func (c *CreateEnvCmd) Run() error {
	envPath, err := createEnv(c.EnvDir, c.Name, c.Overwrite, c.WithConfig, c.WithDockerCompose)
	if err != nil {
		return fmt.Errorf("failed to create environment: %w", err)
	}

	var cfg Config
	if c.WithConfig != "" {
		cfg, err = LoadConfig(c.WithConfig)
		if err != nil {
			return fmt.Errorf("failed to load custom config: %w", err)
		}
	} else {
		cfg, err = LoadConfig(filepath.Join(envPath, "values.yaml"))
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	if err := configureEnv(cfg, envPath); err != nil {
		return fmt.Errorf("failed to configure environment: %w", err)
	}

	return nil
}

func createEnv(envDir string, name string, overwrite bool, customConfig, customCompose string) (string, error) {
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

	configPath := filepath.Join(envPath, "values.yaml")
	_, err := os.Stat(configPath)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check config file: %w", err)
	}

	// Create the global config if it doesn't exist
	if os.IsNotExist(err) || overwrite {
		if customConfig != "" {
			log.Info().Msgf("Using custom configuration file: %s", customConfig)
			// Copy the custom config file to the environment directory
			err := copyFile(customConfig, configPath)
			if err != nil {
				return "", fmt.Errorf("faled to copy custom config file: %w", err)
			}
		} else {
			err := renderTemplateToFile(getTemplates(), "templates/global/values.yaml", nil, configPath)
			if err != nil {
				return "", err
			}
		}
	}

	// Create .gitignore file
	gitignorePath := filepath.Join(envPath, ".gitignore")
	_, err = os.Stat(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check .gitignore file: %w", err)
	}
	if os.IsNotExist(err) || overwrite {
		err := renderTemplateToFile(getTemplates(), "templates/global/gitignore", nil, gitignorePath)
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
		if customCompose != "" {
			err := copyFile(customCompose, dockerComposePath)
			if err != nil {
				return "", fmt.Errorf("faled to copy custom docker-compose file: %w", err)
			}
		} else {
			err := renderTemplateToFile(getTemplates(), "templates/global/docker-compose.yaml", nil, dockerComposePath)
			if err != nil {
				return "", err
			}
		}
	}

	return envPath, nil
}
