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
	EnvDir string `help:"Directory containing the environment. default: './env'" short:"d"`
	Name   string `help:"Name of the environment to retrieve logs for. default: 'default'" short:"n"`
	Follow bool   `help:"Follow log output." short:"f"`
}

func (c *LogsCmd) Run() error {
	// check if env exists
	rc := RuntimeConfigFromFlags(c.EnvDir, c.Name)
	envPath := filepath.Join(rc.EnvDir, rc.EnvName)

	cfgPath := filepath.Join(envPath, "values.yaml")
	cfg, err := LoadEnvironmentConfig(cfgPath)
	if err != nil {
		return err
	}

	args := []string{"logs"}

	if c.Follow {
		args = append(args, "--follow")
	}

	dockerComposeCmd := buildDockerComposeCommand(rc.EnvName, cfg, args...)

	fmt.Println(strings.Join(dockerComposeCmd, " "))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cmd := exec.CommandContext(ctx, dockerComposeCmd[0], dockerComposeCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = envPath
	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil
		}
		return err
	}

	return nil
}
