package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/go-multierror"

	"github.com/scality/workbench"
)

func getTemplates() fs.FS {
	if CLI.TemplatesDir != "" {
		return os.DirFS(CLI.TemplatesDir)
	}

	return workbench.ConfigTemplates
}

func templateFile(templates fs.FS, path string, data any) ([]byte, error) {
	tmpl, err := template.ParseFS(templates, path)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)
	if err := tmpl.Execute(buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderTemplateToFile(templates fs.FS, tmplPath string, data any, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	rendered, err := templateFile(templates, tmplPath, data)
	if err != nil {
		return fmt.Errorf("failed to template %s: %w", tmplPath, err)
	}

	if err := os.WriteFile(outPath, rendered, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", outPath, err)
	}

	return nil
}

func renderTemplates(cfg EnvironmentConfig, srcDir, destDir string, templates []string) error {
	templateFS := getTemplates()
	for _, tmpl := range templates {
		templatePath := filepath.Join(srcDir, tmpl)
		outputPath := filepath.Join(destDir, tmpl)
		if err := renderTemplateToFile(templateFS, templatePath, cfg, outputPath); err != nil {
			return fmt.Errorf("failed to render template %s: %w", tmpl, err)
		}
	}
	return nil
}

func getComposeProfiles(cfg EnvironmentConfig) []string {
	profiles := []string{"base"}

	if cfg.Features.Scuba.Enabled {
		profiles = append(profiles, "feature-scuba")
	}

	if cfg.Features.BucketNotifications.Enabled {
		profiles = append(profiles, "feature-notifications")
	}

	if cfg.Features.Utapi.Enabled {
		profiles = append(profiles, "feature-utapi")
	}

	if cfg.Features.Migration.Enabled {
		profiles = append(profiles, "feature-migration")
	}

	return profiles
}

func buildDockerComposeCommand(cfg EnvironmentConfig, args ...string) []string {
	profiles := getComposeProfiles(cfg)

	dockerComposeCmd := []string{
		"docker",
		"compose",
		"--env-file",
		"defaults.env",
	}

	for _, profile := range profiles {
		dockerComposeCmd = append(dockerComposeCmd, "--profile", profile)
	}

	return append(dockerComposeCmd, args...)
}

func copyFile(src, dest string) (err error) {
	source, err := os.Open(src)
	if err != nil {
		return
	}

	defer func() {
		inErr := source.Close()
		if inErr != nil {
			err = multierror.Append(err, inErr)
		}
	}()

	destination, err := os.Create(dest)
	if err != nil {
		return
	}

	defer func() {
		inErr := destination.Close()
		if inErr != nil {
			err = multierror.Append(err, inErr)
		}
	}()

	_, err = io.Copy(destination, source)
	return
}
