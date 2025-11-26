import { useCallback, useEffect, useState } from 'react';

const spacingScalar = 8;
const spacing = (value: number = 1) => value * spacingScalar;

const shape = { borderRadius: 8 };

export const lightTheme = {
    primary: '#33318F',
    secondary: '#1A300F',
    tertiary: '#5CC791',
    link: '#1A300F',
    error: '#B44641',
    contrast: '#000000',
    neutral: {
        primary: '#FFFFFF',
        secondary: '#F4F4F4',
        tertiary: '#E3E7EA',
        quaternary: '#DADEE1',
        quinary: '#CACFD3',
    },
    shape,
    spacing,
};

export const darkTheme = {
    primary: '#33318F',
    secondary: '#1A300F',
    tertiary: '#5CC791',
    link: '#99A3FF',
    error: '#E9827C',
    contrast: '#FFFFFF',
    neutral: {
        primary: '#121212',
        secondary: '#222222',
        tertiary: '#272727',
        quaternary: '#2C2C2C',
        quinary: '#2E2E2E',
    },
    shape,
    spacing,
};

const getHtmlTag = () => (typeof document !== 'undefined' ? document.documentElement : undefined);

const ThemeCSSVars = {
    primary: '--primary',
    secondary: '--secondary',
    tertiary: '--tertiary',
    link: '--link',
    error: '--error',
    contrast: '--contrast',
    neutral: {
        primary: '--neutral-1',
        secondary: '--neutral-2',
        tertiary: '--neutral-3',
        quaternary: '--neutral-4',
        quinary: '--neutral-5',
    },
};

const observerOptions = {
    attributes: true,
    attributeFilter: ['class'],
};

export type Theme = typeof lightTheme;

/**
 * React hook that syncs the application's theme with CSS variables
 * defined on the `<html>` element.
 *
 * Whenever the `<html>` class list changes (e.g., switching themes), the
 * hook re-reads the CSS variables, updating the returned theme object.
 */
export const useTheme = () => {
    const htmlTag = getHtmlTag();

    const [theme, setTheme] = useState<Theme>(htmlTag?.classList.contains('dark') ? darkTheme : lightTheme);

    const updateTheme = useCallback(() => {
        const htmlTag = getHtmlTag();
        if (!htmlTag) return;

        const computedStyles = getComputedStyle(htmlTag);

        const neutralPrimary = computedStyles.getPropertyValue(ThemeCSSVars.neutral.primary);
        const neutralSecondary = computedStyles.getPropertyValue(ThemeCSSVars.neutral.secondary);
        const neutralTertiary = computedStyles.getPropertyValue(ThemeCSSVars.neutral.tertiary);
        const neutralQuaternary = computedStyles.getPropertyValue(ThemeCSSVars.neutral.quaternary);
        const neutralQuinary = computedStyles.getPropertyValue(ThemeCSSVars.neutral.quinary);

        const primary = computedStyles.getPropertyValue(ThemeCSSVars.primary);
        const secondary = computedStyles.getPropertyValue(ThemeCSSVars.secondary);
        const tertiary = computedStyles.getPropertyValue(ThemeCSSVars.tertiary);

        const link = computedStyles.getPropertyValue(ThemeCSSVars.link);
        const error = computedStyles.getPropertyValue(ThemeCSSVars.error);
        const contrast = computedStyles.getPropertyValue(ThemeCSSVars.contrast);

        setTheme({
            primary,
            secondary,
            tertiary,
            link,
            error,
            contrast,
            neutral: {
                primary: neutralPrimary,
                secondary: neutralSecondary,
                tertiary: neutralTertiary,
                quaternary: neutralQuaternary,
                quinary: neutralQuinary,
            },
            shape,
            spacing,
        });
    }, []);

    const mutationCallback = useCallback(
        (mutationsList: MutationRecord[]) => {
            for (const mutation of mutationsList) {
                if (mutation.type === 'attributes' && mutation.attributeName === 'class') updateTheme();
            }
        },
        [updateTheme]
    );

    useEffect(() => {
        const htmlTag = getHtmlTag();
        if (!htmlTag) return;

        const observer = new MutationObserver(mutationCallback);

        observer.observe(htmlTag, observerOptions);

        return () => {
            observer.disconnect();
        };
    }, [mutationCallback, updateTheme, theme]);

    return theme;
};
