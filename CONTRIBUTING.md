# Contributing to the BloodHound Community Edition Repository

## Prerequisites (Dev Container)

BloodHound Community Edition has a `devcontainer` definition to make it easier to get started contributing. This does currently
require VS Code (though IntelliJ is currently working on supporting the `devcontainer` spec), but it takes the work out of
wrangling dependencies on your own. If you do not want to use VS Code or the `devcontainer`, feel free to skip to
[Prerequisites (Manual)](#prerequisites-manual)

To get started with Dev Containers, follow the official documentation: https://code.visualstudio.com/docs/devcontainers/containers#_getting-started

Once you have your dependencies installed, you can move on to pulling this repository into a dev container by following:
https://code.visualstudio.com/docs/devcontainers/containers#_quick-start-open-a-git-repository-or-github-pr-in-an-isolated-container-volume

When you get to the portion where you need to specify a repository, use this repository. This will then spin up a new Dev 
Container volume and start cloning the repository into it. This can take some time, as it needs to download the repository
and all prerequisites as well. Once complete, you'll be able to see all the files in the project and have access to an integrated
terminal where you can continue to run dev tools.

## Prerequisites (Manual)

<details>
<summary>
  <a href="https://github.com/casey/just">just</a>
</summary>

### Install

| Operating System                                               | Package Manage       | Package                 | Command                  |
| -------------------------------------------------------------- | -------------------- | ----------------------- | ------------------------ |
| [Various][rust-platforms]                                      | [Cargo][cargo]       | [just][just-crate]      | `cargo install just`     |
| [Microsoft Windows][windows]                                   | [Scoop][scoop]       | [just][just-scoop]      | `scoop install just`     |
| [macOS][macos]                                                 | [Homebrew][homebrew] | [just][just-brew]       | `brew install just`      |
| [macOS][macos]                                                 | [MacPorts][macports] | [just][just-port]       | `port install just`      |
| [Arch Linux][arch-linux]                                       | [pacman][pacman]     | [just][just-pkg-arch]   | `pacman -S just`         |
| [NixOS][nixos], [Linux][nix-platforms], [macOS][nix-platforms] | [Nix][nix]           | [just][just-pkg-alpine] | `nix-env -iA nixos.just` |
| [FreeBSD][freebsd]                                             | [pkg][pkgng]         | [just][just-bsd]        | `pkg install just`       |
| [Alpine Linux][alpine-linux]                                   | [apk-tools][apk]     | [just][just-pkg-alpine] | `apk add just`           |

[rust-platforms]: https://forge.rust-lang.org/release/platform-support.html
[cargo]: https://www.rust-lang.org
[just-crate]: https://crates.io/crates/just
[windows]: https://en.wikipedia.org/wiki/Microsoft_Windows
[scoop]: https://scoop.sh
[just-scoop]: https://github.com/ScoopInstaller/Main/blob/master/bucket/just.json
[macos]: https://en.wikipedia.org/wiki/MacOS
[homebrew]: https://brew.sh
[just-brew]: https://formulae.brew.sh/formula/just
[macports]: https://www.macports.org
[just-port]: https://ports.macports.org/port/just/summary
[arch-linux]: https://www.archlinux.org
[pacman]: https://wiki.archlinux.org/title/Pacman
[just-pkg-arch]: https://archlinux.org/packages/community/x86_64/just/
[nixos]: https://nixos.org/nixos/
[nix-platforms]: https://nixos.org/nix/manual/#ch-supported-platforms
[nix]: https://nixos.org/nix/
[just-pkg-nix]: https://github.com/NixOS/nixpkgs/blob/master/pkgs/development/tools/just/default.nix
[freebsd]: https://www.freebsd.org/
[pkgng]: https://www.freebsd.org/doc/handbook/pkgng-intro.html
[just-bsd]: https://www.freshports.org/deskutils/just/
[alpine-linux]: https://alpinelinux.org/
[apk]: https://wiki.alpinelinux.org/wiki/Alpine_Linux_package_management
[just-pkg-alpine]: https://pkgs.alpinelinux.org/package/edge/community/x86_64/just

#### Direct Download

Use the installer to download the latest binary for your platform:

```sh
$ curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to <some PATH directory>
```

</details>

<!-- Install Python -->
<details>
<summary><a href="python.org">Python</a></summary>

### Version

We require Python version 3.10 or higher

### Install

Download and run the installer for your platform: https://www.python.org/downloads/

### Setup

It is encouraged to run Python in a venv for Beagle:

```sh
$ python3 -m venv ~/beagle_venv
```

We also encourage using mypy and black (for vetting Beagle)
```sh
$ ~/beagle_venv/bin/pip install --upgrade pip black mypy
```

More information about Beagle can be found in [./packages/python/beagle/README.md](./packages/python/beagle/README.md)

</details>

<!-- Install Golang -->
<details>
<summary><a href="golang.org">Golang</a></summary>

### Install

-   Download and run the installer for your platform https://golang.org/dl/
-   Run `go get -u golang.org/x/tools/...` for `goimports` and other tools.

</details>

<!-- Install NodeJS -->
<details>
<summary>
<a href="https://nodejs.org">NodeJS</a>
</summary>

### Install

#### Direct Download

Download the installer for the latest LTS version from https://nodejs.org/en/download/

#### n - Interactive NodeJS Version Management (Linux/MacOS)

[n](https://github.com/tj/n) is a very simple NodeJS version manager.

```
curl -L https://git.io/n-install | bash \
    && n lts
```

</details>

<details>
<summary>
  <a href="https://yarnpkg.com">yarn</a>
</summary>

After installing NodeJS, run the following command:

```bash
$ npm install -g yarn
```

</details>

<!-- Install Container Engine -->
<details>

<summary><a href="https://www.docker.com/get-started">Docker</a> or <a href="https://podman.io/getting-started/">Podman</a></summary>

### About

Docker and Podman are both OCI compliant container engines. The Docker CLI communicates
with a background daemon that runs with root privileges to build images and execute containers. In contrast, Podman does
not have a daemon and executes containers with user privileges instead of root. It is possible to run the Docker daemon
as a non-root user following [these instructions](https://docs.docker.com/engine/security/rootless/).

### Enabling BuildKit for Docker Desktop

Enabling BuildKit introduces smart caching mechanisms and concurrency to achieve faster image builds. To enable
BuildKit, go to `Preferences > Docker Engine` and edit the JSON to include the following properties:

```json
{
  ...
  "features": {
    "buildkit": true
  },
  "experimental": true,
  ...
}
```

</details>

## Initialize Local Configuration

Running `just init` will automatically clone config templates for you, allowing all other commands to function as expected.

## Building

The BloodHound team maintains a Python tool for building and testing the project called [`beagle`](./packages/python/beagle/README.md).
Building locally requires simply running `just build`. At the end, all the artifacts will be available in `dist`.

## Dev Mode

Running the project in dev mode is meant to be relatively straightforward. We use Docker Compose to set up containers running
the various services needed to develop BloodHound. There are several useful Just commands for running the project in different
modes, but the main one is `just bh-dev`. This will start up all the services needed to run BloodHound with hot reloading
enabled. Making a change to any TypeScript or Go files should be visible nearly instantly, allowing for rapid development
iteration. If you need to connect a debugger to the API, simply run `just bh-debug` and make sure your debugger is correctly
configured.

### Debug Config

For VS Code, there is already a profile created for debugging called Debug with Docker Compose

For debugging in a JetBrains IDE:

-   Open `Run/Debug Configurations` and add a new `Go Remote` config
-   Host: `localhost`
-   Port: `2345`
-   On disconnect: `Leave it running`

## Testing

### Notes

When running tests with `just test`, you can use the following flags to control behavior:
-   `-a`: Run all tests, even if nothing has changed since the last run
-   `-i`: Run all tests, including integration tests
-   `-v`: Run in verbose mode (display output while tests are running)

### Unit tests

To run all unit tests, run the following command:

```bash
$ just test
```

#### Notes

The UI uses `react-testing-library`. Learn more with the links provided below.

-   https://create-react-app.dev/docs/running-tests/#react-testing-library
-   https://testing-library.com/docs/react-testing-library/intro

Because this app uses a handful of top-level wrappers and context providers, there is a custom `render` function provided via `src/test-utils.js` which can be used instead of importing `render` from `@testing-library/react` directly. It is recommended that you use this custom `render` in most cases.

-   https://testing-library.com/docs/react-testing-library/setup#custom-render

#### Organizing Unit Tests

Our unit tests are organized in the following manner to encourage modular/reusable API design emphasizing consumer
experience. Unit testing requires appropriate use of mocks, stubs, and fakes to separate testing the unit from testing
its dependencies.

##### Typescript

For each file, unit testing shall maintain the following pattern:

-   `src/**/<file name>.{ts,tsx}` - The unit(s) being tested
-   `src/**/<file name>.test.{ts,tsx}` - The unit tests

##### Golang

For each file in a package, unit testing shall maintain the following pattern:

-   `<package path>/<file name>.go` - The unit(s) being tested
-   `<package path>/<file name>_test.go` - Unit tests for package exports
    -   The package of this file shall be `<package>_test`
-   `<package path>/<file_name>_internal_test.go` - Unit tests for non-exported functions (optional)
    -   The package of this file shall be `<package>`
    -   Reserved for situations where:
        -   functions cannot be exported
        -   control-flow is too difficult or expensive to test with only mocks, etc.

#### Mocks

##### Golang

The project uses the [mockgen](https://go.uber.org/mock) mocking framework to generate mocks. In order to maintain
consistency and to keep our mocks in sync with the source code, there are established conventions for how to create and
organize mocks for this project.

-   For each package, there should be a package `<package path>/mocks`
-   For each file in a package that contains an interface(s), there should be a `//go:generate` comment in the following format:
    ```
    //go:generate go run go.uber.org/mock/mockgen -destination=./mocks/filename.go -package=mocks . Comma,Separated,List,Of,Interfaces
    ```
-   For mocking a third party dependency update `cmd/api/src/vendormocks/vendor.go` with a new `//go:generate` comment in the following form:
    ```
    //go:generate go run go.uber.org/mock/mockgen -destination=./{package path}/mock.go -package={package basename} {package} Comma,Separated,List,Of,Interfaces
    ```

By following these conventions, mocks can be kept in sync by running `just go generate ./...`

### Integration tests

The following command will run integration tests in addition to unit tests. Simply make sure that you have your testing
databases running (`just bh-testing`) before running them.

```bash
just test -i
```

To run [cypress](https://www.cypress.io/) UI integration tests:

```bash
$ npx cypress run
```

To open [cypress](https://www.cypress.io/) UI integration testing GUI:

```bash
$ npx cypress open
```

## Dependencies

### Golang

This project uses [Go Modules](https://golang.org/doc/modules/managing-dependencies) for dependency management. We also
include a `go.work` workspace to make it easier to work with our many modules in this monorepo.

#### To add all project dependencies

```bash
$ go mod download
```

#### To add/update a specific dependency (needs to be run within a specific module)

```bash
$ go get some.com/module@v1.2.3
```

#### To view outdated modules

```bash
$ go list -m -u all
```

### Dev Dependencies

To use 3rd-party tooling written in golang, you must target the `go.tools.mod` file using the `-modfile` flag. For example,
to add a new dev dependency you would execute `go get -modfile go.tools.mod some.com/3rd-party/tool@v1.2.3`.

#### How do I use a dev dependency?

To use a dev dependency, it is best to temporarily set the `GOBIN` environment variable to `${PROJECT_ROOT_PATH}/bin` and
even better to wrap the tool in a script of some sort to more easily return the shell environment back to its original state.
This will ensure that everyone is using the same version of the tool without having to vendor it in our VCS.

```bash
#!/usr/bin/env bash

readonly PROJECT_ROOT_PATH="$(dirname $(realpath ${0}))/path/to/project/root"

export GOBIN="$PROJECT_ROOT_PATH/bin"
$GOBIN/myDevTool "$@"
```

### NodeJS

This project uses [Yarn 3](https://yarnpkg.com/) for dependency management.

#### To install all project dependencies

```bash
$ yarn
```

#### To add/update a specific dependency

```bash
$ yarn add <package_name>@<version>
```

#### To view/update outdated modules

```bash
$ yarn upgrade-interactive
```

### Dev Dependencies

#### To add/update a dev dependency

```bash
$ yarn add -D <package_name>@<version>
```

#### How do I use a dev dependency?

To use a dev dependency, it is best to use it in the `scripts` section of the `package.json`. Doing so allows
contributors to use the tool without:

-   needing to globally install it
-   accidentally using the incorrect version
-   needing to use the relative path to the new tool (e.g. - `./node_modules/.bin/<dev tool>`).

As an example:

```json
{
    ...
    "scripts": {
        "cool-tools": "cool-tools",
        ...
    }
    ...
}
```

Usage:

```bash
$ yarn cool-tools [tool args/flags]
```

## Releases and Versioning

The BloodHound Community Edition version specification follows [Semantic Versioning (semver)](https://semver.org) prefixed with the rune `'v'`.

> Given a version number MAJOR.MINOR.PATCH, increments will happen with:
>
> MAJOR version for incompatible API changes,
> MINOR version for functionality in a backwards compatible manner, and
> PATCH version for backwards compatible bug fixes.
>
> Additional labels for pre-release and build metadata are available as extensions to the MAJOR.MINOR.PATCH format.

The BloodHound product family will utilize consistent versioning across all products, and some version numbers may only 
release to a single product within the family. Bloodhound may therefore skip version numbers through releases to maintain 
said consistency and bring it in lockstep with other product version numbers.

All releases will be announced in the 
[BloodHound Slack workspace.](https://join.slack.com/t/bloodhoundhq/shared_invite/zt-1tgq6ojd2-ixpx5nz9Wjtbhc3i8AVAWw)


## Adding Database Migrations

Database migrations are stored as plain SQL, and are version tagged using the same version format as BloodHound. Migration files
may be found in: `src/database/migration/migrations`.

Migrations are recorded in the PostgreSQL database as a table named `migrations`. This table is updated in a
transactional context when step-wise migrations are applied to the database. Successful migration application is
recorded in the same transaction as the migration, to ensure atomicity.

Versions for migration files should follow the version of the BloodHound service release they are intended for.

## API Documentation

We use swagger to show our API documentation alongside our GUI. You can access these docs in your local dev environment
at http://bloodhound.localhost/ui/api-explorer

### Updating API endpoint documentation

Our API documentation is generated from `{bhce}/cmd/api/src/docs`. The directories in `json` are split according to
the `definitions` section and the `paths` section of the swagger document.

-   `definitions` are used to document any struct types that get exposed to the API, such as error and response structs.
    -   The `definitions` directory splits its JSON across multiple files based on the package each struct comes from (`definitions.json` is used as a catch-all)
-   `paths` are used for all endpoint documentation
    -   The `paths` directory splits its JSON across multiple files based on the endpoint's tag (the collection of endpoints it belongs to).

Documentation is automatically generated from these files whenever the API service is spun up, so they'll show up in the
swagger docs as soon as you build the API service again. No other magic required.
