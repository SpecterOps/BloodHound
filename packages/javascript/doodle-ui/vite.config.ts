/// <reference types="vitest" />
import react from '@vitejs/plugin-react';
import { resolve } from 'path';
import { defineConfig } from 'vite';
import dts from 'vite-plugin-dts';
import packageJson from './package.json';

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [react(), dts({ rollupTypes: true, exclude: ['**/*.stories.{ts,tsx}'] })],
    resolve: {
        alias: {
            components: resolve(__dirname, 'src/components'),
        },
    },
    build: {
        lib: {
            entry: resolve(__dirname, 'src/index.ts'),
            formats: ['es'],
        },
        rollupOptions: {
            external: ['react', 'react-dom', 'react/jsx-runtime', 'tailwindcss'],
            output: {
                manualChunks: {
                    vendor: Object.keys(packageJson.dependencies),
                },
            },
        },
    },
    test: {
        globals: true,
        environment: 'jsdom',
        testTimeout: 60000, // 1 minute,
        clearMocks: true,
    },
});
