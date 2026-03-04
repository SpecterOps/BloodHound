export const themeOptions = [
    'primary',
    'primary-variant',
    'primary',
    'primary-variant',
    'secondary',
    'secondary-variant',
    'tertiary',
    'tertiary-variant',
] as const;

export type ThemeOptions = (typeof themeOptions)[number];

// this means we cant use predefined/cross-browser color names. But values from figma work fine
export type CustomColorOptions =
    | `#${string}`
    | `rgb(${string})`
    | `rgba(${string})`
    | `hsl(${string})`
    | `hsla(${string})`;

export type ColorOptions = ThemeOptions | CustomColorOptions;
