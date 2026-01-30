import { createRouter, RouteIds } from '@tanstack/react-router';
import { routeTree } from './routeTree.gen';

export const router = createRouter({
    routeTree,
    // defaultPreload: 'intent',
    // scrollRestoration: true,
    basepath: '/ui',
    context: { auth: undefined! },
});

export type RouterType = typeof router;
export type RouterIds = RouteIds<RouterType['routeTree']>;

declare module '@tanstack/react-router' {
    interface Register {
        // This infers the type of our router and registers it across your entire project
        router: typeof router;
    }
}

// // Generic to enforce that the route returned matches the route path
// type LazyRouteFn<TRoutePath extends RouterIds> = () => Promise<
//     LazyRoute<RouteById<(typeof router)['routeTree'], TRoutePath>>
// >;

// type RouterMap = {
//     // __root__ is a special route that isn't lazy loaded, so we need to manually bind it
//     // You could consider adding null to the returned type to have routes without rendering
//     [K in Exclude<RouterIds, '__root__'>]: LazyRouteFn<K>;
// };

// const routerMap: RouterMap = {
//     '/explore': () => import('src/views/Explore/GraphView').then((d) => d.default),
//     // '/$postId': () =>
//     //   import('@router-mono-simple-lazy/post-feature/post-id-page').then(
//     //     (d) => d.PostIdRoute,
//     //   ),
// };

// // // Given __root__ is a special route that isn't lazy loaded, we need to update it manually
// // router.routesById['__root__'].update({
// //   component: RootComponent,
// // })

// Object.entries(routerMap).forEach(([path, component]) => {
//     const foundRoute = router.routesById[path as RouterIds];
//     // Bind the lazy route to the actual route
//     foundRoute.lazy(component);
// });
