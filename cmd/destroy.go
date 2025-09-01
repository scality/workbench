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

type DestroyCmd struct {
	EnvDir  string `help:"Directory containing the environment. default: './env'" short:"d"`
	Name    string `help:"Name of the environment to destroy. default: 'default'" short:"n"`
	Timeout int    `help:"Timeout in seconds for stopping containers." short:"t" default:"10"`
}

func (c *DestroyCmd) Run() error {
	// Delete env dir if it exists
	rc := RuntimeConfigFromFlags(c.EnvDir, c.Name)
	envPath := filepath.Join(rc.EnvDir, rc.EnvName)
	if info, err := os.Stat(envPath); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s exists but is not a directory", envPath)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat environment: %w", err)
	}

	if err := os.RemoveAll(envPath); err != nil {
		return fmt.Errorf("failed to remove environment: %w", err)
	}

	cfgPath := filepath.Join(envPath, "values.yaml")
	cfg, err := LoadEnvironmentConfig(cfgPath)
	if err != nil {
		return err
	}

	args := []string{"down", "--volumes", "--timeout", fmt.Sprintf("%d", c.Timeout)}

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
