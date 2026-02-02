import { setupWorker } from 'msw';
import { searchHandlers } from './handlers';

export const worker = setupWorker(...searchHandlers);
