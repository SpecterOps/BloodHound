// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
import '@fontsource/roboto/100-italic.css';
import '@fontsource/roboto/100.css';
import '@fontsource/roboto/300-italic.css';
import '@fontsource/roboto/300.css';
import '@fontsource/roboto/400-italic.css';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500-italic.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700-italic.css';
import '@fontsource/roboto/700.css';
import '@fontsource/roboto/900-italic.css';
import '@fontsource/roboto/900.css';
import { withThemeByClassName } from '@storybook/addon-themes';
import type { Preview, ReactRenderer } from '@storybook/react';
import '../src/input.css';

const preview: Preview = {
    parameters: {
        backgrounds: { disable: true },
        controls: {
            matchers: {
                date: /Date$/i,
            },
        },
    },
    decorators: [
        withThemeByClassName<ReactRenderer>({
            themes: {
                light: '',
                dark: 'dark',
            },
            defaultTheme: 'light',
        }),
    ],
};

export default preview;
