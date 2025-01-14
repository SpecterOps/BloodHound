import { DoodleUIPlugin, DoodleUIPreset } from '@bloodhoundenterprise/doodleui';

/** @type {import('tailwindcss').Config} */
export default {
    content: [
        './index.html',
        './src/**/*.{js,ts,jsx,tsx}',
        './node_modules/@bloodhoundenterprise/doodleui/dist/doodleui.js',
    ],
    darkMode: ['class'],
    plugins: [DoodleUIPlugin],
    presets: [DoodleUIPreset],
};
