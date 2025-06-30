package main

import (
	"context"
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
	EnvDir      string `help:"Directory containing the environment." required:"" default:"./env"`
	Name        string `help:"Name of the environment to start." required:"" short:"n" default:"default"`
	NoConfigure bool   `help:"Don't template config files before starting containers"`
	Overwrite   bool   `help:"Overwrite existing environment if it exists." short:"o"`
	Detach      bool   `help:"Run containers in detached mode." short:"d"`
	Build       bool   `help:"Build images before starting containers." short:"b"`
	NoCache     bool   `help:"Do not use cache when building images." short:"c"`
}

func (c *UpCmd) Run() error {
	// check if env exists
	envPath := filepath.Join(c.EnvDir, c.Name)
	created := false
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		if _, err := createEnv(c.EnvDir, c.Name, c.Overwrite); err != nil {
			return fmt.Errorf("failed to create environment: %w", err)
		}
		created = true
	}

	cfgPath := filepath.Join(envPath, "config.yaml")
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		return err
	}

	if created || !c.NoConfigure {
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
		if ctx.Err() == context.Canceled {
			return nil
		}
		return err
	}

	return nil
}
