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
  echo "Make local copies of configuration files"
  if [[ -d "./local-harnesses/build.config.json" ]]; then
    rm -rf "./local-harnesses/build.config.json"
    cp ./local-harnesses/build.config.json.template ./local-harnesses/build.config.json
  elif [[ -f "./local-harnesses/build.config.json" ]] && [[ "{{wipe}}" != "clean" ]]; then
    echo "Not copying build.config.json since it already exists"
  else
    cp ./local-harnesses/build.config.json.template ./local-harnesses/build.config.json
  fi

  if [[ -d "./local-harnesses/integration.config.json" ]]; then
    rm -rf "./local-harnesses/integration.config.json"
    cp ./local-harnesses/integration.config.json.template ./local-harnesses/integration.config.json
  elif [[ -f "./local-harnesses/integration.config.json" ]] && [[ "{{wipe}}" != "clean" ]]; then
    echo "Not copying integration.config.json since it already exists"
  else
    cp ./local-harnesses/integration.config.json.template ./local-harnesses/integration.config.json
  fi

  echo "Install additional Go tools"
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2

  echo "Run a build to ensure go.work.sum is valid"
  just build -vf

  echo "Ensure containers have been rebuilt"
  just bh-dev build

  echo "Init Complete"

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

# run linting for all Go modules
go-lint:
  #!/usr/bin/env bash
  echo 'ensuring golangci-lint@{{golangci-lint-version}} is installed'
  go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@{{golangci-lint-version}}
  echo 'running golangci-lint on all detected modules in your go.work'
  # grab all the module locations from go.work and run golangci-lint on each
  golangci-lint run $(cat go.work | { cat | grep -E '\./' | awk 'NF{print $0 "/..."}'; } | tr '\n' ' ')
  echo 'done'

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
bh-clear-volumes *ARGS='':
  @docker compose -f docker-compose.dev.yml down -v {{ARGS}}

# build BH target cleanly (default profile dev with --no-cache flag)
bh-clean-docker-build target='dev' *ARGS='':
  @docker compose --profile {{target}} -f docker-compose.dev.yml build --no-cache {{ARGS}}

# build local BHCE container image (ex: just build-bhce-container <linux/arm64|linux/amd64> edge v5.0.0)
build-bhce-container platform='linux/amd64' tag='edge' version='v5.0.0' *ARGS='':
  @docker buildx build -f dockerfiles/bloodhound.Dockerfile -t specterops/bloodhound:{{tag}} --platform={{platform}} --load --build-arg version={{version}}-{{tag}} {{ARGS}} .

# run local BHCE container image (ex: just build-bhce-container <linux/arm64|linux/amd64> custom v5.0.0)
run-bhce-container platform='linux/amd64' tag='custom' version='v5.0.0' *ARGS='':
  @just build-bhce-container {{platform}} {{tag}} {{version}} {{ARGS}}
  @cd examples/docker-compose && BLOODHOUND_TAG={{tag}} docker compose up

check-license:
  python3 license_check.py
