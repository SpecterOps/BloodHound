/// <reference types="vitest" />
import react from '@vitejs/plugin-react';
import { resolve } from 'path';
import { defineConfig } from 'vite';
import dts from 'vite-plugin-dts';

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [react(), dts({ rollupTypes: true, exclude: ['**/*.stories.{ts,tsx}'] })],
    build: {
        lib: {
            entry: resolve(__dirname, 'src/index.ts'),
            formats: ['es'],
        },
        rollupOptions: {
            external: ['react', 'react-dom', 'react/jsx-runtime', 'tailwindcss'],
        },
    },
    test: {
        globals: true,
        environment: 'jsdom',
        testTimeout: 60000, // 1 minute,
        clearMocks: true,
    },
});
