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
)

type LogsCmd struct {
	EnvDir string `help:"Directory containing the environment." required:"" default:"./env"`
	Name   string `help:"Name of the environment to start." required:"" short:"n" default:"default"`
	Follow bool   `help:"Follow log output." short:"f"`
}

func (c *LogsCmd) Run() error {
	// check if env exists
	envPath := filepath.Join(c.EnvDir, c.Name)

	cfgPath := filepath.Join(envPath, "values.yaml")
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		return err
	}

	args := []string{"logs"}

	if c.Follow {
		args = append(args, "--follow")
	}

	dockerComposeCmd := buildDockerComposeCommand(cfg, args...)

	fmt.Println(strings.Join(dockerComposeCmd, " "))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cmd := exec.CommandContext(ctx, dockerComposeCmd[0], dockerComposeCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Join(c.EnvDir, c.Name)

	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.Canceled)  {
			return nil
		}
		return err
	}

	return nil
}
