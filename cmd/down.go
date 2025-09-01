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

type DownCmd struct {
	EnvDir  string `help:"Directory containing the environment. default: './env'" short:"d"`
	Name    string `help:"Name of the environment to stop. default: 'default'" short:"n"`
	Timeout int    `help:"Timeout in seconds for stopping containers." short:"t" default:"10"`
	Volumes bool   `help:"Remove named volumes declared in the 'volumes' section of the Compose file and anonymous volumes attached to containers." short:"v"`
}

func (c *DownCmd) Run() error {
	rc := RuntimeConfigFromFlags(c.EnvDir, c.Name)
	envPath := filepath.Join(rc.EnvDir, rc.EnvName)
	info, err := os.Stat(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("environment %s does not exist", c.Name)
		}
		return fmt.Errorf("failed to stat environment: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", envPath)
	}

	cfgPath := filepath.Join(envPath, "values.yaml")
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		return err
	}

	args := []string{"down", "--timeout", fmt.Sprintf("%d", c.Timeout)}
	if c.Volumes {
		args = append(args, "--volumes")
	}

	dockerComposeCmd := buildDockerComposeCommand(cfg, args...)

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
