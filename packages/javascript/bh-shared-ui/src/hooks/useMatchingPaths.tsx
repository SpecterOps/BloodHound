import { matchPath, useLocation } from 'react-router-dom';

export const useMatchingPaths = (pattern: string | string[]) => {
    const { pathname } = useLocation();
    if (typeof pattern === 'string') {
        const match = matchPath({ path: pattern }, pathname);
        return !!match?.pathname;
    } else {
        return pattern.reduce(
            (match: boolean, current) => (match ? match : !!matchPath(pathname, current)?.pathname),
            false
        );
    }
};
