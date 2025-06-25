# BloodHound Community Edition Development

Welcome to BloodHound Community Edition! This README should help point you in the right direction if you're looking to begin development with BloodHound.
It is intended to be a simple place to look to get your dev environment set up.

More detailed information regarding [contributing](https://github.com/SpecterOps/BloodHound/wiki/Contributing), [code structure](https://github.com/SpecterOps/BloodHound/wiki/Code), and [development](https://github.com/SpecterOps/BloodHound/wiki/Development) can be found in our [GitHub wiki](https://github.com/SpecterOps/BloodHound/wiki).

## Dev Environment Prerequisites

- [Git](https://git-scm.com/)
    - for Windows, the [git installer](https://git-scm.com/downloads/win) is a simple option that also comes with git-bash
    - for WSL, Linux, and macOS, git should be included in your distro
- [Docker](https://www.docker.com/get-started/)
    - for linux and WSL, installing the docker cli engine is a pretty straight forward solution
    - for Mac and Windows, the docker desktop app is the easiest option
- [Colima](https://github.com/abiosoft/colima) is another option for macOS users
- [just](https://github.com/casey/just)
    - `just` has many installation options, but the simplest option may be to install the npm package using node
- [NodeJS v22](https://nodejs.org/en)
    - installation using a node version manager like [nvm](https://github.com/nvm-sh/nvm) or [n](https://github.com/tj/n) is recommended
- [Yarn v3.6](https://v3.yarnpkg.com/getting-started/install)
    - installation using `npm corepack` is recommended (instructions in the link above)
- [Go v1.24](https://go.dev/dl/)
  - [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) is also needed
  - make sure `$HOME/go/bin` (or wherever your Go packages are installed) is in your `$PATH`
- [Python v3.10](https://www.python.org/downloads/)

## Quick Start
Please make sure you have all dev requisites and docker is running!
- Run `just init`
  - this should only ever be run once; the only reason to run it again is if you wipe the repo and start fresh or if you want to wipe your config files and start over
- Run `just bh-dev`
- To access the UI, navigate to `http://bloodhound.localhost`.

## Quick Tips

- To see `just` recipes, run the `just` command in the project root
- Code generation: `just generate`
- Running tests:
  - only unit tests: `just test`
  - unit + integration tests: `just test -i`
- Before creating a PR and requesting code review: `just prepare-for-codereview`

## Default Admin
There are two ways a default admin is created in the dev environment. If you would like to have
precise control over the default admin, make sure a default_admin is set in `local-harnesses/build.config.json`.
If you prefer a randomly generated user, remove the default_admin part of the config. This option will create
a user called `admin` and randomly generate a password which will be printed to the console output.
This only happens on the first startup when the user is created. If the application has already been
bootstrapped, and you want to change how the default admin is created, you will need to clear the data
volumes and rebuild the docker containers.

## Package Names
Packages in the BHCE repo will follow one of two conventions:

- If the package lives within `cmd/api/src`:
  - `github.com/specterops/bloodhound/src/<path-to>/<package>`
  - is covered by the `go.mod` file in the `src` directory
- If the package lives within `packages/go`:
  - `github.com/specterops/bloodhound/<package>`
  - requires its own `go.mod`
