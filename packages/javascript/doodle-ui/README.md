<p align="center">
<img src="https://img.shields.io/badge/version-1.0.0--alpha.40-teal" alt="version 1.0.0-alpha.40"/>
<a href="https://ghst.ly/BHSlack">
<img src="https://img.shields.io/badge/BloodHound Slack-4A154B?logo=slack&logoColor=EEF0F2"
    alt="BloodHound Slack"></a>
</p>

# DoodleUI

A component library for use with [BloodHound Community Edition](https://github.com/SpecterOps/BloodHound) and [BloodHound Enterprise](https://bloodhoundenterprise.io/).

This library is written in TypeScript and leverages [Radix](https://www.radix-ui.com/) components as its foundation via [shadcn](https://ui.shadcn.com/). [Tailwind CSS](https://tailwindcss.com/) is used for styling along with [Class Variance Authority](https://cva.style/docs) for creating opinionated variants as defined by our design system.

# Installation

## Using TailwindCSS

1. Install [TailwindCSS](https://tailwindcss.com/docs/installation)
2. Install DoodleUI (example using `yarn` shown below)

```
$ yarn add @bloodhoundenterprise/DoodleUI
```

3. Update your Tailwind configuration to include the DoodleUI plugin, preset and content

```
import { DoodleUIPlugin, DoodleUIPreset } from '@bloodhoundenterprise/doodleui';

module.exports = {
  content: [
    "./src/**/*.{html,js}" // your application source code
    "node_modules/@bloodhoundenterprise/doodleui/dist/index.js" // DoodleUI components
  ],
  plugins: [DoodleUIPlugin],
  presets: [DoodleUIPreset],
  ...
}
```

These configuration options provide the base theme customizations and additional utility classes required to render DoodleUI components in alignment with the design system used by BloodHound Community Edition and BloodHound Enterprise.

## Manual Installation

1. Install DoodleUI (example using `yarn` shown below)

```
$ yarn add @bloodhoundenterprise/DoodleUI
```

2. Add the DoodleUI stylesheet to your application

```
<link rel="stylesheet" href="node_modules/@bloodhoundenterprise/doodleui/dist/styles.css">
```

## Developer Notes

### Dependencies

-   [Node.js 22.x](https://nodejs.org/)

### Typography

DoodleUI uses:

-   **Nunito Sans** for `h1` through `h6`.
-   **Figtree** for body, subtitle, caption, and other UI text.

Both fonts are distributed under the SIL Open Font License 1.1. DoodleUI's Storybook self-hosts the font files with
[Fontsource](https://fontsource.org/). Applications consuming DoodleUI must load the fonts in their own entrypoint,
following the existing self-hosted Fontsource approach.

Via Fontsource:

```
yarn add @fontsource/figtree @fontsource/nunito-sans
```

Then import the required weights in your entrypoint:

```
import '@fontsource/figtree/400.css';
import '@fontsource/figtree/500.css';
import '@fontsource/nunito-sans/600.css';
import '@fontsource/nunito-sans/700.css';
```

The preset falls back to `Segoe UI`, Helvetica, Arial, and a generic sans-serif for Figtree. Nunito Sans additionally
falls back through `Avenir Next` before the shared system sans-serif stack.

| Variant   | Family      | Size / line height | Weight | Tracking | Light color |
| --------- | ----------- | ------------------ | ------ | -------- | ----------- |
| h1        | Nunito Sans | 24 / 28px          | 700    | 0        | Text/Main   |
| h2        | Nunito Sans | 22 / 24px          | 700    | 0        | Text/Main   |
| h3        | Nunito Sans | 20 / 22px          | 700    | 0        | Text/Main   |
| h4        | Nunito Sans | 20 / 22px          | 600    | 0        | Text/Main   |
| h5        | Nunito Sans | 18 / 20px          | 700    | 0.25px   | Text/Main   |
| h6        | Nunito Sans | 16 / 18px          | 600    | 0.25px   | Text/Main   |
| body1     | Figtree     | 16 / 24px          | 400    | 0        | Text/Muted  |
| body2     | Figtree     | 14 / 22px          | 400    | 0        | Text/Muted  |
| subtitle1 | Figtree     | 15 / 24px          | 500    | 0.25px   | Text/Main   |
| subtitle2 | Figtree     | 13 / 22px          | 500    | 0.25px   | Text/Main   |
| caption   | Figtree     | 12 / 20px          | 400    | 0.25px   | Text/Muted  |

The `text-muted` utility is a compatibility alias backed by the existing `--text-light` semantic value. The legacy
`text-light` utility remains available. Body and caption variants retain their existing main-text behavior in dark
mode.

#### Visual Language Refresh migration note

The public `Typography` props and variant names are unchanged. Consumers should expect intentionally shorter heading
line boxes and different heading/body font metrics. Load both font families before first render to minimize fallback
movement, and review layouts that previously depended on the old heading heights.

### Getting Started

Clone this repository

```
git clone git@github.com:SpecterOps/DoodleUI.git
```

Install dependencies with `yarn`

```
cd DoodleUI
yarn
```

Start the dev server

```
yarn dev
```

### Other Scripts

| Command                           | Description                                          |
| --------------------------------- | ---------------------------------------------------- |
| dev                               | Start the dev server                                 |
| build                             | Build the component library                          |
| lint                              | Run linter checks                                    |
| test                              | Run vitest                                           |
| storybook                         | Same as dev                                          |
| build:storybook                   | Build storybook documentation                        |
| build:styles                      | Generate CSS via TailwindCSS                         |
| generate-index                    | Update `src/components/index.ts` automatically       |
| create-component <component name> | Create a new component in `src/components`           |
| format:check                      | Check file formatting                                |
| format:write                      | Fix file formatting                                  |
| update-badge                      | Updates the version badge in the [README](README.md) |

### Publishing a new DoodleUI version

PRs which include changes to components must also increment the version number in `package.json` (in accordance with [semver](https://semver.org/)), otherwise these changes will not be publishable. Once a PR with a new version has been merged into main, complete the following steps to publish the new version to NPM:

-   Checkout main branch and pull latest changes from remote
-   Run the following scripts (found in package.json):
    -   `build:styles`
    -   `build:storybook`
    -   `format:write`
    -   `yarn build`
    -   `npm login (use credentials in 1password)`
    -   `npm publish --access public --tag alpha`

Now, any package which depends on DoodleUI will be able to access the latest version of the library by updating their `package.json` to reference the new version number.

## Licensing

```
Copyright 2024 Specter Ops, Inc.

Licensed under the Apache License, Version 2.0
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

Unless otherwise annotated by a lower-level LICENSE file or license header, all files in this repository are released
under the `Apache-2.0` license. A full copy of the license may be found in the top-level [LICENSE](LICENSE) file.
