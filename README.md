# Workbench

Workbench is a tool for creating a development environment for Scality S3C based on Docker Compose.
It takes a feature based approach allowing users to run only the components needed for their use case.

Workbench requires Docker Compose v2.

## Building

Go 1.24.3 or newer is required to build Workbench.

```shell
> git clone https://github.com/scality/workbench.git
> cd workbench/
> make
```

## Usage

```
Usage: s3c-workbench <command> [flags]

Run a light S3C for development and testing.

Flags:
  -h, --help                 Show context-sensitive help.
      --log-level="info"     Set the log level.
      --log-format="text"    Set the log format. (json, text)
      --templates-dir=""     Directory containing config templates. Overrides embedded templates.

Commands:
  create-env    Create a new S3C workbench environment.
  up            Start a S3C workbench environment.
  configure     Generate configuration files from templates.
  destroy       Destroy a S3C workbench environment.
  down          Stop a S3C workbench environment.
  logs          View logs of a S3C workbench environment.

Run "s3c-workbench <command> --help" for more information on a command.
```

### Environments

The configuration for a workbench is contained in an environment.
Workbench looks for an `env` directory in the current working directory.
Each subdirectory represents an environment of the same name.

A new environment can be created using the `create-env` subcommand

```shell
workbench create-env
```

If no environment name is specified `default` is used.

```shell
> ls env/
default/
> ls env/default/
config/  values.yaml  defaults.env  docker-compose.yaml  logs/
```

### Configuration

Each environment has a `values.yaml` that is used to enable features and configure individual components.

```yaml
global:
  log_level: info

features:
  scuba:
    enabled: false
    enable_service_user: true

  bucket_notifications:
    enabled: false
    targetAuth:
      type: basic
      username: admin
      password: admin123

cloudserver:
  image: ghcr.io/scality/cloudserver:7.70.62

vault:
  image: ghcr.io/scality/vault:7.73.0

backbeat:
  image: ghcr.io/scality/backbeat:9.0.12-federation

s3_metadata:
  image: ghcr.io/scality/metadata:7.70.47-standalone

scuba:
  image: ghcr.io/scality/scuba:1.1.0-preview.5-nodesvc-base

scuba_metadata:
  image: ghcr.io/scality/metadata:7.70.47-standalone

kafka:
  image: bitnami/kafka:3.4.0

zookeeper:
  image: bitnami/zookeeper:3.8.1

redis:
  image: redis:7
```

### Starting a workbench

To start a workbench use `workbench up`.
This will regenerate component files based on the workbench config and run `docker compose up` with the selected feature profiles.

```shell
> workbench up -d
12:38PM INF Configuring environment env/default
12:38PM INF Starting environment command="docker-compose --env-file defaults.env --profile base up --detach"
[+] Running 5/5
 ✔ Container workbench-vault        Healthy                         5.7s
 ✔ Container workbench-s3           Started                         0.2s
 ✔ Container workbench-metadata-s3  Started                         0.2s
 ✔ Container workbench-s3-data      Started                         0.2s
 ✔ Container workbench-setup-vault  Started                         5.8s
```
