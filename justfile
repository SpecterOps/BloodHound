_default:
	@just --list --unsorted

golangci-lint-version := "v1.53.3"
host_os := if os() == "macos" { "darwin" } else { os() }
host_arch := if arch() == "x86" { "386" } else { if arch() == "x86_64" { "amd64" } else { if arch() == "aarch64" { "arm64" } else { arch() } } }

export CGO_ENABLED := "0"
export GOOS := env_var_or_default("GOOS", host_os)
export GOARCH := env_var_or_default("GOARCH", host_arch)
export INTEGRATION_CONFIG_PATH := env_var_or_default("INTEGRATION_CONFIG_PATH", absolute_path("./local-harnesses/integration.config.json"))

set positional-arguments

# Initialize your dev environment (use "just init clean" to reset your config files)
init wipe="":
  #!/usr/bin/env bash
  echo "Init BloodHound CE"
  echo "Make local copies of configuration files"
    if [[ -f "./local-harnesses/build.config.json" ]] && [[ "{{wipe}}" != "clean" ]]; then
    echo "Not copying build.config.json since it already exists"
  elif [[ -f "./local-harnesses/build.config.json" ]]; then
    echo "Backing up build.config.json and resetting"
    mv ./local-harnesses/build.config.json ./local-harnesses/build.config.json.bak
    cp ./local-harnesses/build.config.json.template ./local-harnesses/build.config.json
  else
    cp ./local-harnesses/build.config.json.template ./local-harnesses/build.config.json
  fi

  if [[ -f "./local-harnesses/integration.config.json" ]] && [[ "{{wipe}}" != "clean" ]]; then
    echo "Not copying integration.config.json since it already exists"
  elif [[ -f "./local-harnesses/integration.config.json" ]]; then
    echo "Backing up integration.config.json and resetting"
    mv ./local-harnesses/integration.config.json ./local-harnesses/integration.config.json.bak
    cp ./local-harnesses/integration.config.json.template ./local-harnesses/integration.config.json
  else
    cp ./local-harnesses/integration.config.json.template ./local-harnesses/integration.config.json
  fi

  if [[ -f "./.env" ]] && [[ "{{wipe}}" == "clean" ]]; then
    echo "Backing up existing environment file"
    mv ./.env ./.env.bak
  fi

  echo "Install additional Go tools"
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2

  echo "Run modsync to ensure workspace is up to date"
  just modsync

  echo "Ensure containers have been rebuilt"
  if [[ "{{wipe}}" != "clean" ]]; then
    just bh-dev build
  else
    echo "Clear volumes and rebuild without cache"
    just bh-clear-volumes
    just bh-clean-docker-build
  fi

  echo "Start integration testing services"
  if [[ "{{wipe}}" == "clean" ]]; then
    echo "Clear volumes and restart testing services without cache"
    just bh-testing-clear-volumes
    just bh-testing build --no-cache
  fi

  echo "BloodHound CE Init Complete"

# Show available targets for this context.
show *FLAGS:
  #!/usr/bin/env bash
  set -euo pipefail
  python3 packages/python/beagle/main.py show {{FLAGS}}

# Build a target. To list targets run 'just show'
build *FLAGS:
  #!/usr/bin/env bash
  set -euo pipefail
  python3 packages/python/beagle/main.py build {{FLAGS}}

# Test a target. To list targets run 'just show'
test *FLAGS:
  #!/usr/bin/env bash
  set -euo pipefail
  python3 packages/python/beagle/main.py test {{FLAGS}}

# sync modules in workspace
modsync:
  @go run github.com/specterops/bloodhound/packages/go/stbernard modsync

# updates favicon.ico, logo192.png and logo512.png from logo.svg
update-favicon:
  @just imagemagick convert -background none ./cmd/ui/public/logo-light.svg -define icon:auto-resize ./cmd/ui/public/favicon-light.ico
  @just imagemagick convert -background none -size 192x192 cmd/ui/public/logo-light.svg cmd/ui/public/logo-light192.png
  @just imagemagick convert -background none -size 512x512 cmd/ui/public/logo-light.svg cmd/ui/public/logo-light512.png
  @just imagemagick convert -background none ./cmd/ui/public/logo-dark.svg -define icon:auto-resize ./cmd/ui/public/favicon-dark.ico
  @just imagemagick convert -background none -size 192x192 cmd/ui/public/logo-dark.svg cmd/ui/public/logo-dark192.png
  @just imagemagick convert -background none -size 512x512 cmd/ui/public/logo-dark.svg cmd/ui/public/logo-dark512.png

# run go commands in the context of the api project
go *ARGS:
  @cd cmd/api/src && GODEBUG=cgocheck=2 go {{ARGS}}

# run yarn commands in the context of the workspace root
yarn-local *ARGS="":
  @yarn {{ARGS}}

# run yarn commands in the context of the workspace root and rebuild containers
yarn *ARGS="": && (bh-dev "build bh-ui")
  @yarn {{ARGS}}

build-js-client *ARGS="":
  @cd packages/javascript/js-client-library && yarn build

build-shared-ui *ARGS="":
  @cd packages/javascript/bh-shared-ui && yarn build

# run the code generation from the cue schema
schemagen: yarn-local && check-license (yarn "format")
  go run github.com/specterops/bloodhound/schemagen

# run imagemagick commands in the context of the project root
imagemagick *ARGS:
  @docker run -it --rm -v {{justfile_directory()}}:/workdir -w /workdir --entrypoint magick cblunt/imagemagick {{ARGS}}

# run git pruning on merged branches to clean up local workspace (run with `nuclear` to clean up orphaned branches)
prune-my-branches nuclear='no':
  #!/usr/bin/env bash
  git branch --merged| egrep -v "(^\*|master|main|dev)" | xargs git branch -d
  git reflog expire --expire=now --all && git gc --prune=now --aggressive
  git remote prune origin
  if [ "{{nuclear}}" == 'nuclear' ]; then
    echo Switching to main to remove orphans
    git switch main
    git branch -vv | grep ': gone]' | grep -v "\*" | awk '{ print $1; }' | xargs -r git branch -D
    git switch -
  fi
  echo "Remaining Git Branches:"
  git --no-pager branch

# Run all analyzers (requires jq to be installed locally)
analyze:
  go run github.com/specterops/bloodhound/packages/go/stbernard analysis | jq 'sort_by(.severity) | .[] | {"severity": .severity, "description": .description, "location": "\(.location.path):\(.location.lines.begin)"}'

# run docker compose commands for the BH dev profile (Default: up)
bh-dev *ARGS='up':
  @docker compose --profile dev -f docker-compose.dev.yml {{ARGS}}

# run docker compose commands for the BH debug profile (Default: up)
bh-debug *ARGS='up':
  @docker compose --profile debug-api -f docker-compose.dev.yml {{ARGS}}

# run docker compose commands for the BH api-only profile (Default: up)
bh-api-only *ARGS='up':
  @docker compose --profile api-only -f docker-compose.dev.yml {{ARGS}}

# run docker compose commands for the BH ui-only profile (Default: up)
bh-ui-only *ARGS='up':
  @docker compose --profile ui-only -f docker-compose.dev.yml {{ARGS}}

# run docker compose commands for the BH testing databases (Default: up)
bh-testing *ARGS='up -d':
  @docker compose --project-name bh-testing -f docker-compose.testing.yml {{ARGS}}

# clear BH testing volumes
bh-testing-clear-volumes *ARGS='':
  @docker compose --project-name bh-testing -f docker-compose.testing.yml down -v {{ARGS}}

# clear BH docker compose volumes (pass --remove-orphans if troubleshooting)
bh-clear-volumes target='dev' *ARGS='':
  @docker compose --profile {{target}} -f docker-compose.dev.yml down -v {{ARGS}}

# build BH target cleanly (default profile dev with --no-cache flag)
bh-clean-docker-build target='dev' *ARGS='':
  # Ensure the target is down first
  @docker compose --profile {{target}} -f docker-compose.dev.yml down
  # Pull any updated images
  @docker compose --profile {{target}} -f docker-compose.dev.yml pull
  # Build without cache
  @docker compose --profile {{target}} -f docker-compose.dev.yml build --no-cache {{ARGS}}

# use docker compose watch to dynamically restart/rebuild containers (requires docker compose v2.22.0+)
bh-watch target='dev' *ARGS='--no-up':
  @docker compose --profile {{target}} -f docker-compose.dev.yml -f docker-compose.watch.yml watch {{ARGS}}

# build local BHCE container image (ex: just build-bhce-container <linux/arm64|linux/amd64> edge v5.0.0)
build-bhce-container platform='linux/amd64' tag='edge' version='v5.0.0' *ARGS='':
  @docker buildx build -f dockerfiles/bloodhound.Dockerfile -t specterops/bloodhound:{{tag}} --platform={{platform}} --load --build-arg version={{version}}-{{tag}} {{ARGS}} .

# run local BHCE container image (ex: just build-bhce-container <linux/arm64|linux/amd64> custom v5.0.0)
run-bhce-container platform='linux/amd64' tag='custom' version='v5.0.0' *ARGS='':
  @just build-bhce-container {{platform}} {{tag}} {{version}} {{ARGS}}
  @cd examples/docker-compose && BLOODHOUND_TAG={{tag}} docker compose up

check-license:
  python3 license_check.py
