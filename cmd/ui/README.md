This project was bootstrapped with [Create React App](https://github.com/facebook/create-react-app).

## Contributing

Welcome to the BloodHound UI! If this is your first time contributing, please check out our [contributing
guide](./CONTRIBUTING.md) for instructions on setting up your environment. If you find something isn't well documented,
feel free to submit a PR. Cheers!

## Typography adoption

BHCE consumes the local `doodle-ui` workspace (`workspace:*`) as the source of truth for typography. Standard headings use
Nunito Sans through DoodleUI heading variants, while body, subtitle, caption, and other UI text use Figtree. The legacy MUI
theme is a compatibility bridge and mirrors the same DoodleUI families and metrics. Its Body1, Body2, and Caption variants
use the DoodleUI `--text-muted` semantic token in light mode and retain their existing color behavior in dark mode. New UI
should prefer semantic DoodleUI `Typography` variants.

The remaining font exceptions are intentional:

-   Roboto Mono and the explicit monospace stacks are retained for code and Cypher editor content, where character alignment is
    functional. A future editor-specific token can replace these declarations.
-   Graph, icon, Swagger, and print styles retain context-specific sizes where they describe canvas labels, icon glyphs, generated
    API documentation, compact graph chrome, or print-only output rather than standard product typography. These should be
    revisited when those specialized surfaces receive dedicated semantic tokens.
-   Sigma canvas labels explicitly select Figtree because canvas-rendered text cannot inherit CSS or Tailwind utilities.

Responsive validation identified an existing Privilege Zones fixed-width content region that clips at a 640 CSS-pixel
viewport (the 200%-zoom equivalent used by this audit), even after navigation is collapsed. Correcting that behavior
requires structural page-layout work and is tracked separately from typography adoption. The typography browser harness
retains an expected-failure regression scenario so this limitation cannot be reported as passing.

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

### `yarn start`

Runs the Web UI in development mode.<br />
Open [http://localhost:3000](http://localhost:3000) to view it in the browser.

The page will reload if you make edits.<br />
You will also see any lint errors in the console.

### `yarn test`

Launches the test runner in the interactive watch mode.<br />
See the section about [running tests](https://facebook.github.io/create-react-app/docs/running-tests) for more information.

### `yarn build`

Builds the app for production to the `build` folder.<br />
It correctly bundles React in production mode and optimizes the build for the best performance.

The build is minified and the filenames include the hashes.<br />
Your app is ready to be deployed!

See the section about [deployment](https://facebook.github.io/create-react-app/docs/deployment) for more information.
