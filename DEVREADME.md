# BloodHound Community Edition

Welcome to BloodHound Community Edition!

## Using BloodHound Community Edition

For an example of how to use BloodHound Community Edition with Docker Compose, please see [examples/docker-compose](./examples/docker-compose/README.md)

## Top-Level BloodHound Package Namespaces

### Golang Package Namespaces

All golang packages in the repository should follow the naming schemes enumerated below:

#### BloodHound Main Package Namespace

`github.com/specterops/bloodhound/src`

#### Common Package Namespaces

`github.com/specterops/bloodhound/<package_dir>`

##### Examples

-   `github.com/specterops/bloodhound/dawgs`
-   `github.com/specterops/bloodhound/crypto`

## Contributing

Welcome to BloodHound! If this is your first time contributing, please check out our
[contributing guide](./CONTRIBUTING.md) for instructions on setting up your environment. If you find something
isn't well documented, feel free to submit a PR. Cheers!

## Quickstart

### Prerequisites

There are two ways to build your dev environment:

-   Dev Container
    -   Requires an editor with Dev Container support (only VS Code is supported right now, requires the Dev Containers Remote Extension)
    -   Requires Docker/Docker Compose (easiest option is [Docker Desktop](https://www.docker.com/products/docker-desktop/))
    -   Everything else is automatically set up for you inside the container, enjoy!
-   Manual
    -   [Just](https://github.com/casey/just)
    -   [Python 3.10](https://www.python.org/downloads/)
    -   [Go 1.20](https://go.dev/dl/)
    -   [Node 18](https://nodejs.dev/en/download/)
    -   [Yarn 3.6](https://yarnpkg.com/getting-started/install)
    -   [Docker Desktop](https://www.docker.com/products/docker-desktop/) (or Docker/Docker Compose compatible runtime)

### Useful Just commands

#### See all Just commands

```sh
$ just
```

#### Initialize dev environment (copy configs, run first time docker compose build, etc)

```sh
$ just init
```

#### Build Everything (Requires local Python/Go/Node)

```sh
$ just build
```

#### Unit Test Everything (Requires local Python/Go/Node)

```sh
$ just test
```

#### Deploy Everything in Continuous Dev Mode

```sh
$ just bh-dev
```

#### Deploy Everything in Continuous Dev Mode with Debugging Enabled

```sh
$ just bh-debug
```

For VS Code, there is already a profile created for debugging called Debug with Docker Compose

For debugging in JetBrains IDE:

-   Open `Run/Debug Configurations` and add a new `Go Remote` config
-   Host: `localhost`
-   Port: `2345`
-   On disconnect: `Leave it running`

#### Wipe databases

```sh
$ just bh-clear-volumes
```

#### Start integration test databases

```sh
$ just bh-testing
```

## Local dev environment

| Service  | Url                                    | Username   | Password                         |
| -------- | -------------------------------------- | ---------- | -------------------------------- |
| UI / API | http://bloodhound.localhost            | admin      | SFdzJoW2GT7Fn68aEieKn7S1S2DLdXnw |
| Neo4J    | http://neo4j.localhost                 | neo4j      | bloodhoundcommunityedition       |
| pgAdmin  | http://pgadmin.localhost               | bloodhound | bloodhoundcommunityedition       |
| Postgres | postgresql://localhost:5432/bloodhound | bloodhound | bloodhoundcommunityedition       |

## Project Structure

<!-- Intentionally blank for markdown rendering purposes -->

```

.
├── cmd                                // Artifact source code and their integration tests
|   |                                  // Partitioned By: Utility
│   ├── api
│   ├── ui
|   └── ...
|
├── dist                               // Built artifacts (git ignored)
|
├── dockerfiles                        // Misc. Dockerfiles
|
├── examples                           // Examples for Deploying BloodHound
|
├── packages                           // Libraries/Modules/Packages/Plugins (a.k.a. re-useable do-dads) and their UNIT tests
|   |                                  // Partitioned By: Language/Runtime
│   ├── go
│   |   └── ...
|   |
|   ├── python
│   |   └── beagle
|   |
│   └── javascript
│       └── ...
|
└── tools                              // Misc. tooling (docker compose related files, etc)
```

## Documentation

API Documentation is available at in the API Explorer (`/ui/api-explorer`)

## Local Integration Testing

### VS Code

VS Code is already configured to run integration tests against the testing database docker containers. Simply run `just init`
to initialize the necessary configurations and be sure the testing databases are running if you haven't started them already
(`just bh-testing`).

You can then use the normal testing options in VS Code to discover tests and run them (Testing charm, test run options inline
in files, etc).

### Beagle

The BloodHound team maintains a Python tool called [`beagle`](./packages/python/beagle/README.md), which is currently found
in `/packages/python/beagle`. Simply run `just show` to see all plans that beagle is capable of running (tests and builds).
You can run `just test -avi` to run all API and UI tests with integration tests enabled. Or you can run `just test -avi <plan_name>`
to run all tests from a specific plan with integration tests enabled.
