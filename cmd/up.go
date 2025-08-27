package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"
)

type UpCmd struct {
	EnvDir            string `help:"Directory containing the environment." required:"" default:"./env"`
	Name              string `help:"Name of the environment to start." required:"" short:"n" default:"default"`
	NoConfigure       bool   `help:"Don't template config files before starting containers"`
	Overwrite         bool   `help:"Overwrite existing environment if it exists." short:"o"`
	Detach            bool   `help:"Run containers in detached mode." short:"d"`
	Build             bool   `help:"Build images before starting containers." short:"b"`
	NoCache           bool   `help:"Do not use cache when building images." short:"c"`
	WithConfig        string `help:"Path to a custom configuration file. Replaces the default config." type:"existingfile"`
	WithDockerCompose string `help:"Path to a custom Docker Compose file. Replaces the default file." type:"existingfile"`
}

func (c *UpCmd) Run() error {
	// Create or ensure environment is properly set up
	envPath, err := createEnv(c.EnvDir, c.Name, c.Overwrite, c.WithConfig, c.WithDockerCompose)
	if err != nil {
		return fmt.Errorf("failed to create/setup environment: %w", err)
	}

	cfgPath := filepath.Join(envPath, "values.yaml")
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		return err
	}

	if !c.NoConfigure {
		if err := configureEnv(cfg, envPath); err != nil {
			return fmt.Errorf("failed to configure environment: %w", err)
		}
	}

	args := []string{"up"}
	if c.Detach {
		args = append(args, "--detach")
	}

	if c.Build {
		args = append(args, "--build")
	}

	if c.NoCache {
		args = append(args, "--no-cache")
	}

	dockerComposeCmd := buildDockerComposeCommand(cfg, args...)

	log.Info().Str("command", strings.Join(dockerComposeCmd, " ")).Msg("Starting environment")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cmd := exec.CommandContext(ctx, dockerComposeCmd[0], dockerComposeCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Join(c.EnvDir, c.Name)
	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil
		}
		return err
	}

	return nil
}
