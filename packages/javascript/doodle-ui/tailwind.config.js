/** @type {import('tailwindcss').Config} */
import { DoodleUIPreset, DoodleUIPlugin } from './src/tailwind';

export default {
    presets: [DoodleUIPreset],
    plugins: [DoodleUIPlugin],
    darkMode: ['class'],
    content: ['./src/**/*.tsx', '.storybook/preview.tsx'],
};
