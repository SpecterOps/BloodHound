import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

// This runs after every single test across your project
afterEach(() => {
    cleanup();
});
