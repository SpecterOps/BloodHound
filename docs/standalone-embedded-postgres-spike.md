# Standalone BHCE with Embedded PostgreSQL

## Decision

Yes. BHCE can be shipped as a non-containerized, localhost-only application
that owns its PostgreSQL instance. The correct implementation is a native
launcher that starts a bundled PostgreSQL child process, then starts the
existing BHCE API with the PostgreSQL graph driver. It is not an in-process
database: PostgreSQL remains a separately managed child process.

This is feasible without Neo4j, Docker, Docker Compose, Traefik, or a
separately installed database. It should be a separately packaged
``bloodhound`` executable per operating system and architecture, containing
the compiled UI and a verified PostgreSQL distribution. The user experience is
one command or double-click, followed by a clear `http://127.0.0.1:8080/ui`
message (and optionally opening the browser).

## Why it fits BHCE

The production image already compiles the UI into the API binary:

* `dockerfiles/bloodhound.Dockerfile` builds `cmd/ui` and copies its `dist`
  output into `cmd/api/src/api/static/assets` before compiling `bhapi`.
* `cmd/api/src/api/registration/registration.go` redirects `/` to `/ui` and
  registers the embedded static-asset handler.
* SQL migrations and extension-data migrations are embedded with `go:embed` in
  `cmd/api/src/database/migration/migration.go`.
* With `graph_driver` set to `postgres`, `bootstrap.ConnectGraph` uses the same
  PostgreSQL connection as the relational database. No Neo4j process is
  needed.

The schema requires the stock `pg_trgm` and `intarray` extensions. The PostgreSQL distribution
we package must therefore include the matching extension control and library
files. The existing initial migration issues `CREATE EXTENSION IF NOT EXISTS
pg_trgm`.

## Recommended shape

```
bloodhound launcher
  ├─ creates or reads the per-user application-data directory
  ├─ extracts the bundled PostgreSQL runtime once per application/PG version
  ├─ initializes or opens the persistent PostgreSQL data directory
  ├─ starts postgres on 127.0.0.1 at an allocated private port
  ├─ starts bhapi with graph_driver=postgres and the local DSN
  └─ reports (or opens) http://127.0.0.1:8080/ui
```

Use platform application-data locations, not the executable directory:

* macOS: `~/Library/Application Support/BloodHound/`
* Windows: `%LOCALAPPDATA%\\BloodHound\\`
* Linux: `${XDG_DATA_HOME:-~/.local/share}/bloodhound/`

Within that directory, keep at least `postgres/data`, a versioned extracted
runtime, logs, and a persisted BHCE configuration. Bind PostgreSQL only to
loopback and generate a random database password and a unique local port. Do
not expose PostgreSQL or metrics on non-loopback interfaces by default.

Persisting BHCE configuration is required: the current defaults generate a JWT
signing key at startup. Recreating it on each launch would invalidate active
sessions. First-run setup must also present or securely save the generated
administrator credential; today it is logged to stdout when the installation
is initialized.

## PostgreSQL packaging options

`github.com/fergusstrange/embedded-postgres` is a useful proof-of-concept
library: it starts a real PostgreSQL child process and supports explicit data,
runtime, binaries, port, and credentials. It downloads PostgreSQL binaries by
default, so that default must not be used for the intended offline,
single-download user experience.

For a shippable product, build or obtain a version-pinned PostgreSQL runtime
for each supported OS/architecture, verify its checksum in CI, and bundle it
with the release (either as release-adjacent files or as a compressed archive
embedded in the launcher). Configure the library with `BinariesPath`, or own
the small `initdb`/`pg_ctl` process lifecycle directly. The latter gives more
control over upgrade, logging, parent/child cleanup, and recovery behavior.

The first prototype should target macOS arm64 and Linux amd64. Windows needs
separate validation of process cleanup, executable paths, firewall behavior,
and bundled-runtime licensing/notices.

## Required product work

1. Add a `standalone` command or launcher mode that owns PostgreSQL lifecycle,
   receives termination signals, stops PostgreSQL gracefully, and removes only
   the extracted runtime—not persistent data.
2. Generate and persist a standalone configuration with `graph_driver` set to
   `postgres`, loopback API binding, a stable JWT key, and the local DSN.
3. Make the port policy explicit: API defaults to 8080 when available; choose
   or report a fallback otherwise. PostgreSQL always uses an unadvertised,
   allocated loopback port.
4. Add first-run credential UX, durable logs, a `--data-dir` override,
   `--no-browser`, backup/restore, and a clear reset-data command.
5. Add a release pipeline that builds the UI, API, and each PostgreSQL runtime;
   emits checksums and third-party notices; and signs/notarizes platform
   artifacts where applicable.
6. Test first launch, relaunch with preserved data, interrupted startup,
   database migration, port collision, upgrade, and offline installation on
   every supported platform.

## Risks and gates

* PostgreSQL upgrades are major-version data-directory upgrades, not merely a
  binary replacement. The launcher must pin the runtime version and provide a
  tested `pg_upgrade` or dump/restore path before changing major versions.
* The API's default Argon2 configuration reserves 1 GiB. A standalone product
  needs documented resource requirements and likely a reviewed local-profile
  configuration rather than silently weakening security settings.
* Large BHCE datasets make backup, disk-space checks, clean shutdown, and
  corruption recovery first-class product requirements.
* Browser auto-open is a convenience only. The service must remain usable when
  launching a browser is disabled by policy or unavailable.

## Prototype acceptance criteria

A prototype answers the question conclusively when a clean supported machine,
with no Docker and no installed PostgreSQL or Neo4j, can run one downloaded
artifact offline; initialize data; open the UI at localhost; ingest a small
dataset; restart with the data and login still usable; and stop without leaving
PostgreSQL running.

## Current macOS arm64 spike

The `just build-standalone darwin-arm64 <runtime-root>` recipe builds one
executable at `dist/bloodhound-darwin-arm64`. It builds the UI into the API's
embedded asset directory, archives the supplied PostgreSQL runtime, and builds
the API with the `standalone_release` tag. Running the resulting executable
with `-standalone` needs no PostgreSQL path or installation on the host.

The standalone runtime is pinned to PostgreSQL 18.6, matching BHCE's PostgreSQL
18 container deployment. Homebrew PostgreSQL is not a valid runtime root because its executable depends
on Homebrew libraries. The `just build-standalone-postgres-runtime` recipe
builds a self-contained PostgreSQL runtime from a checksum-verified source
archive and rejects Homebrew-linked output. A release must sign that runtime
and the final executable.
