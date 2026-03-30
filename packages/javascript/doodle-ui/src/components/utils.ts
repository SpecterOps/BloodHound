import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { CustomColorOptions, themeOptions, ThemeOptions } from '../types';

export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

/**
 *
 * @param color if using a theme color, use the css global var. Otherwise, just use the custom color value
 * @returns string
 */
export function getCssColor(color: ThemeOptions | CustomColorOptions) {
    if (themeOptions.includes(color as ThemeOptions)) {
        // TODO: the typing here could be better
        return `var(--${color})`;
    }

    return color;
}

/**
 *
 * @param hedges [conditional, styles] - when conditional is true, styles will be applied in a cascading order. See example below.
 * @returns React.CSSProperties
 * @example ([truthy, {color: 'pink', fontSize: 12}],
 * [truthy, {color: 'green'}],
 * [falsy, {color: 'blue'}])
 * // {color: 'green', fontSize: 12}
 */
export function getConditionalStyles(...hedges: [boolean, React.CSSProperties][]): React.CSSProperties {
    return hedges.reduce((a, c) => {
        const [conditional, styles] = c;
        if (conditional) return { ...a, ...styles };
        return a;
    }, {});
}
