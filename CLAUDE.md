# workbench

This is a **Go CLI tool** for creating Docker Compose-based development environments for Scality S3C (S3 Connector). It contains:

- CLI commands for environment lifecycle (`cmd/`)
- Go text templates rendered into Docker Compose configs, shell scripts, and service configurations (`templates/`)
- Template embedding via `embed.go`
- Kong-based CLI parsing, zerolog structured logging, mergo config merging

## Overview

Workbench uses a feature-based approach allowing users to run only the components needed for their use case.

**Key Concept**: Workbench manages isolated "environments" in the `env/` directory. Each environment contains its own configuration, docker-compose setup, and runtime state.

## Building and Testing

### Build
```bash
make build
# Creates executable named 'workbench'
```

### Lint
```bash
make lint
# Requires golangci-lint to be installed
```

### Clean
```bash
make clean
```

## Core Commands

Workbench requires Docker Compose v2 and uses Go 1.24.3+.

### Environment Management
```bash
# Create a new environment (defaults to 'default' if no name provided)
workbench create-env [--name <env-name>] [--env-dir <dir>]

# Start an environment (regenerates configs and runs docker compose)
workbench up [-d] [--name <env-name>] [--env-dir <dir>]

# Stop an environment
workbench down [--name <env-name>] [--env-dir <dir>]

# Destroy an environment (removes containers and volumes)
workbench destroy [--name <env-name>] [--env-dir <dir>]

# View logs
workbench logs [--name <env-name>] [--env-dir <dir>]

# Regenerate configuration files without starting
workbench configure [--name <env-name>] [--env-dir <dir>]
```

### Environment Variables
- `WORKBENCH_ENV_DIR`: Override default environment directory (default: `./env`)
- `WORKBENCH_ENV_NAME`: Override default environment name (default: `default`)

## Architecture

### Configuration System

The configuration system uses a three-level hierarchy:

1. **Runtime Config** (`cmd/config.go:RuntimeConfig`): Determines which environment to operate on
   - Environment directory location
   - Environment name

2. **Environment Config** (`cmd/config.go:EnvironmentConfig`): Per-environment settings loaded from `values.yaml`
   - Global settings (log level)
   - Feature flags (scuba, bucket notifications, utapi, migration)
   - Component-specific settings (images, ports, log levels)

3. **Component Templates** (`templates/`): Go text templates rendered into configuration files
   - Each component has templates for configs, scripts, Dockerfiles
   - Templates are embedded into the binary via `embed.go`

### Config Flow

1. User creates/modifies `env/<name>/values.yaml`
2. `configure` command loads config (with defaults merged in via `mergo`)
3. Templates are rendered with config data → `env/<name>/config/`
4. Docker Compose profiles are selected based on enabled features
5. Docker Compose starts containers using generated configs

### Component Architecture

**Main components** (all use host networking):
- **cloudserver**: S3 API server (two containers: API server + data server)
- **metadata-s3**: Metadata storage for S3 (standalone mode, Raft-based)
- **vault**: Authentication and authorization service
- **backbeat**: Replication and lifecycle management
- **scuba**: Utilization metrics reporting (requires `features.scuba.enabled: true`)
- **metadata-scuba**: Separate metadata instance for Scuba
- **utapi**: Usage metrics API (requires `features.utapi.enabled: true`)
- **zookeeper**: Distributed coordination and leader election
- **kafka**: Message queue for notifications/replication
- **redis**: Caching layer
- **migration-tools**: Metadata migration utilities (requires `features.migration.enabled: true`)

**Docker Compose Profiles**:
- `base`: Core S3 functionality (always active)
- `feature-scuba`: Scuba utilization metrics
- `feature-notifications`: Bucket notifications via Backbeat
- `feature-utapi`: Usage metrics collection
- `feature-migration`: Metadata migration support

### Key Files

- `cmd/main.go`: CLI entry point using Kong for command parsing
- `cmd/config.go`: Configuration structs and loading logic with defaults
- `cmd/configure.go`: Template rendering orchestration
- `cmd/util.go`: Template rendering helpers, profile selection, docker compose command building
- `cmd/up.go`, `cmd/down.go`, etc.: Individual command implementations
- `embed.go`: Embeds `templates/` directory into binary
- `templates/global/docker-compose.yaml`: Master Docker Compose file template
- `templates/global/defaults.env`: Environment variables for Docker Compose
- `templates/global/values.yaml`: Example configuration with all available options

### Metadata Configuration

Metadata instances support two version formats:
- `v0`
- `v1`

Metadata instances use configurable ports via `base_ports`:
- `bucketd`: Bucket service API
- `repd`: Replication service
- `repdAdmin`: Replication admin API

Each metadata instance can have multiple Raft sessions (controlled by `raft_sessions`).

### Migration Support

When `features.migration.enabled: true`:
- Deploys a second set of metadata bucketd instances with separate ports
- Configured via `s3_metadata.migration.base_ports`

## Development Notes

- All container logs go to `env/<name>/logs/`
- Container names are prefixed with `workbench-`
- Host networking is used for all containers (no port mapping needed)
- Configuration generation is idempotent (safe to run `configure` multiple times)
- The `up` command always regenerates configs before starting (ensures config matches values.yaml)
- Templates can be overridden via `--templates-dir` flag for testing

### Template Update Gotcha

**IMPORTANT**: When modifying template files in `templates/`, Go's build cache can cause stale templates to be embedded in the binary. Running `configure` after a simple rebuild may NOT pick up your changes.

**Correct sequence to force template updates:**

```bash
# 1. Stop everything
./workbench down -n <env-name>

# 2. COMPLETE clean - cache, binary, and environment
go clean -cache -modcache -testcache
rm -f workbench
sudo rm -rf env/<env-name>

# 3. Build fresh
make build

# 4. Create completely new environment
./workbench create-env -n <env-name>

# 5. Configure with new templates
./workbench configure -n <env-name>

# 6. Verify your changes are present
grep "your-change" env/<env-name>/docker-compose.yaml

# 7. Start
./workbench up -n <env-name>
```

The key is: `go clean -cache` + `rm binary` + `rm env` BEFORE rebuilding. This forces Go to rebuild everything from scratch without using any cached embed data.

### Cloudserver Version Support

Cloudserver v7 and v9 are both supported. The version is auto-detected from the `cloudserver.image` tag in values.yaml: semver tags starting with `7.` use v7 config, everything else (including `latest`, git SHAs) defaults to v9.
