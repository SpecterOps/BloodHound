import { useEffect, useState } from 'react';

export const BREAKPOINTS = {
    // tailwind defaults
    sm: '640px',
    md: '768px',
    lg: '1024px',
    xl: '1280px',
    // additions / updates
    '2xl': '1400px',
    '3xl': '1920px',
};

export const useMediaQuery = (query: string): boolean => {
    const [matches, setMatches] = useState(false);

    useEffect(() => {
        const mediaQueryList = window.matchMedia(query);

        // Set initial value
        setMatches(mediaQueryList.matches);

        // Create a listener function
        const documentChangeHandler = (event: MediaQueryListEvent) => {
            setMatches(event.matches);
        };

        // Attach listener
        mediaQueryList.addEventListener('change', documentChangeHandler);

        // Clean up on unmount
        return () => {
            mediaQueryList.removeEventListener('change', documentChangeHandler);
        };
    }, [query]);

    return matches;
};
