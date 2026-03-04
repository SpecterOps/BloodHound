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
