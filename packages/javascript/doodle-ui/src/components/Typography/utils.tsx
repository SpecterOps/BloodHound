export type Variant = 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6' | 'subtitle' | 'body1' | 'body2' | 'caption';

export const DEFAULT_VARIANT: Variant = 'body1';

export const variantMapping: Record<Variant, keyof JSX.IntrinsicElements> = {
    h1: 'h1',
    h2: 'h2',
    h3: 'h3',
    h4: 'h4',
    h5: 'h5',
    h6: 'h6',
    body1: 'p',
    body2: 'p',
    subtitle: 'h6',
    caption: 'span',
};

export const tagOptions = [
    undefined,
    // Headings — document outline hierarchy
    'h1',
    'h2',
    'h3',
    'h4',
    'h5',
    'h6',
    // Block elements — structural then textual
    'div',
    'p',
    'pre',
    // Inline semantic — meaning-bearing
    'code',
    'cite',
    'mark',
    'strong',
    'em',
    // Inline presentational — visual only
    'b',
    'i',
    'u',
    // Inline text modification
    'del',
    'ins',
    'sup',
    'sub',
    'small',
];
