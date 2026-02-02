import { setupWorker } from 'msw/browser';
import { searchHandlers } from './handlers';

export const worker = setupWorker(...searchHandlers);
