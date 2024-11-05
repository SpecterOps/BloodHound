import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

/**
 * This function sets the name of our current theme mode as a class on the html document root. This will ensure the correct styles are applied to components attached elsewhere in the DOM, such as modals and popover menus.
 *
 * @param enabled - set to 'true' if dark mode is currently enabled
 *
 * @returns the name of the currently set mode as a string
 */
export const toggleModeClassOnRoot = (enabled: boolean) => {
    const root = window.document.documentElement;
    const mode = enabled ? 'dark' : 'light';

    root.classList.remove('dark', 'light');
    root.classList.add(mode);

    return mode;
};

/**
 * Utility function for conditionally constructing className strings and merging the result.
 *
 * @param inputs - any number of valid clsx statements. For reference: [clsx docs](https://github.com/lukeed/clsx#readme)
 *
 * @returns a merged class list as a string. For more information about how merging tailwind classes works: [twMerge docs](https://github.com/dcastil/tailwind-merge/blob/v2.5.4/docs/what-is-it-for.md)
 */
export const cn = (...inputs: ClassValue[]) => {
    return twMerge(clsx(inputs));
};
