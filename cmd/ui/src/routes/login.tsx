import { createFileRoute } from '@tanstack/react-router';
import Login from 'src/views/Login';
import { ROUTE_LOGIN } from './constants';

export const Route = createFileRoute(ROUTE_LOGIN)({ component: Login });
