/** @type {import('tailwindcss').Config} */
import { DoodleUIPlugin, DoodleUIPreset } from './src/tailwind';

export default {
    presets: [DoodleUIPreset],
    plugins: [DoodleUIPlugin],
    darkMode: ['class'],
    content: ['./src/**/*.tsx', '.storybook/preview.tsx'],
};
