# DoodleUI

A component library for use with [BloodHound Community Edition](https://github.com/SpecterOps/BloodHound) and [BloodHound Enterprise](https://bloodhoundenterprise.io/).

This library is written in TypeScript and leverages [Radix](https://www.radix-ui.com/) components as its foundation via [shadcn](https://ui.shadcn.com/). [Tailwind CSS](https://tailwindcss.com/) is used for styling along with [Class Variance Authority](https://cva.style/docs) for creating opinionated variants as defined by our design system.

<p align="center">
<img src="https://img.shields.io/badge/version-1.0.0--alpha.24-teal" alt="version 1.0.0-alpha.24"/>
<a href="https://ghst.ly/BHSlack">
<img src="https://img.shields.io/badge/BloodHound Slack-4A154B?logo=slack&logoColor=EEF0F2"
    alt="BloodHound Slack"></a>
</p>

# Installation

## Using TailwindCSS

1. Install [TailwindCSS](https://tailwindcss.com/docs/installation)
2. Update your Tailwind configuration to include the DoodleUI plugin, preset and content

```
import { DoodleUIPlugin, DoodleUIPreset } from './src/tailwind';

export default {
    presets: [DoodleUIPreset],
    plugins: [DoodleUIPlugin],
    darkMode: ['class'],
    content: ['./src/**/*.tsx', '.storybook/preview.tsx'],
};
```

These configuration options provide the base theme customizations and additional utility classes required to render DoodleUI components in alignment with the design system used by BloodHound Community Edition and BloodHound Enterprise.

## Developer Notes

### Dependencies

-   [Node.js 22.x](https://nodejs.org/)

These components are built for usage with the Roboto font though there are fallback fonts in place if Roboto is not found. The Roboto font will need to be included in your project's assets or it will need to be pulled in via CDN for the font to display as expected.

Via Fontsource:

```
yarn add @fontsource/roboto
```

Then import the font in your entrypoint:

```
import '@fontsource/roboto/400.css';
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
