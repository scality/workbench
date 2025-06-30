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
)

type DestroyCmd struct {
	EnvDir  string `help:"Directory containing the environment." required:"" short:"d" default:"./env"`
	Name    string `help:"Name of the environment to destroy." required:"" short:"n" default:"default"`
	Timeout int    `help:"Timeout in seconds for stopping containers." short:"t" default:"10"`
}

func (c *DestroyCmd) Run() error {
	// Delete env dir if it exists
	envPath := filepath.Join(c.EnvDir, c.Name)
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

	cfgPath := filepath.Join(envPath, "config.yaml")
	cfg, err := LoadConfig(cfgPath)
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
	cmd.Dir = filepath.Join(c.EnvDir, c.Name)
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.Canceled {
			return nil
		}
		return err
	}

	return nil
}
