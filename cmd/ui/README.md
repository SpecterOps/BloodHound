This project was bootstrapped with [Create React App](https://github.com/facebook/create-react-app).

## Contributing

Welcome to the BloodHound UI! If this is your first time contributing, please check out our [contributing
guide](./CONTRIBUTING.md) for instructions on setting up your environment. If you find something isn't well documented,
feel free to submit a PR. Cheers!

## Quickstart

The following command will spin up the Web UI, API, a PostgreSQL database, a Neo4J database, and continuously rebuild/sync while
you modify the source files.

To build everything:

```bash
$ skaffold build
```

To run local profile in continuous development mode:

```bash
$ skaffold dev -p local
```

For a one-off local deployment, just run:

```bash
$ skaffold run -p local
```

To spin down a one-off local deployment, just run:

```bash
$ skaffold delete -p local
```

## The non-containerized way of doing things

### `pnpm start` or `pnpm dev`

Runs the Web UI in development mode.<br />
Open [http://localhost:3000](http://localhost:3000) to view it in the browser.

The page will reload if you make edits.<br />
You will also see any lint errors in the console.

### `pnpm test`

Launches the test runner in the interactive watch mode.<br />
See the section about [running tests](https://facebook.github.io/create-react-app/docs/running-tests) for more information.

### `pnpm build`

Builds the app for production to the `build` folder.<br />
It correctly bundles React in production mode and optimizes the build for the best performance.

The build is minified and the filenames include the hashes.<br />
Your app is ready to be deployed!

See the section about [deployment](https://facebook.github.io/create-react-app/docs/deployment) for more information.

## Using NX

This project now uses NX for build orchestration and caching. You can also use NX commands:

```bash
nx run bloodhound-ui:dev    # Start dev server
nx run bloodhound-ui:build  # Build the app
nx run bloodhound-ui:test   # Run tests
```

Or from the root:

```bash
pnpm dev    # Start dev server
pnpm build  # Build all packages
pnpm test   # Run all tests
```
