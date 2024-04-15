# St Bernard

## A BloodHound Swiss Army Knife

St Bernard is a multi-purpose tool for working with BloodHound repositories. It handles running builds, tests, code analysis,
dependency syncing, and much more!

```
$ go run github.com/specterops/bloodhound/packages/go/stbernard -h

A BloodHound Swiss Army Knife

Usage:  stbernard [OPTIONS] COMMAND

Options:
  -v          Verbose output
  -vv         Debug output

Commands:
  envdump     Dump your environment variables
  deps        Ensure workspace dependencies are up to date
  modsync     Sync all modules in current workspace
  generate    Run code generation in current workspace
  show        Show current project info
  analysis    Run static analyzers
  test        Run tests for entire workspace
  build       Build commands in current workspace
  cover       Collect coverage reports
```

### Usage

St Bernard can be run most easily with `go run`:

```
$ go run github.com/specterops/bloodhound/packages/go/stbernard
```

You can find current usage help and available commands by passing the `-h` or `-help` flag. If you'd like to know what additional options are supported by a specific subcommand, you can run the subcommand with `-h` or `-help` to get subcommand specific options:

```
$ go run github.com/specterops/bloodhound/packages/go/stbernard test -h
```

The options available to stbernard should be used _before_ the subcommand. Subcommand options always come after the subcommand:

```
$ go run github.com/specterops/bloodhound/packages/go/stbernard -vv test -g
```

### Configuration

St Bernard does not use a bespoke configuration file. A generic `yarn-workspaces.json` file is used to convey important yarn directories as they cannot be easily inferred the way other paths are.

The following environment variables are supported:

-   `SB_LOG_LEVEL`: takes a level name from among `debug`, `info`, `warn`, `error`, and `fatal`
-   `SB_COVERAGE_PATH`: allows setting a path other than `./tmp/coverage` to store Go coverage files in
-   Pass-through of any tool specific environment variables, such as for changing the Go path for caching purposes. Some pass-through variables have sane defaults defined, but will be overridden if you set the environment variables yourself.

### Contributing

St Bernard is a tool for BloodHound devs. If you think of something you want to see added, feel free to create a pull request. New subcommands can be added fairly easily by observing an existing subcommand and changing out the details as needed, then registering the new subcommand in `command/command.go`. Additional packages are used to group useful tools that multiple subcommands could make use of or for better code structuring.

A lot of work went into making this tool as approachable as possible, but we will always strive to make it better.
