import { createRouter, RouteIds } from '@tanstack/react-router';
import { AuthState } from './ducks/auth/types';
import { routeTree } from './routeTree.gen';

export const router = createRouter({
    routeTree,
    defaultPreload: 'intent',
    basepath: '/ui',
    context: { auth: undefined! },
});

export type RouterType = typeof router;
export type RouterIds = RouteIds<RouterType['routeTree']>;

export interface RouterContext {
    auth: AuthState;
}

declare module '@tanstack/react-router' {
    interface Register {
        // This infers the type of our router and registers it across your entire project
        router: typeof router;
    }
    interface StaticDataRouteOption {
        showNavbar?: boolean;
    }
}
