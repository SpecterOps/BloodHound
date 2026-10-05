// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
import { screen } from '@testing-library/react';
import { render } from '../../utils';
import { Typography, TypographyVariants } from './Typography';
import { DEFAULT_VARIANT, variantMapping } from './utils';

describe('Typography', () => {
    const expectedVariantClasses = {
        h1: ['font-heading', 'text-2xl', 'font-bold', 'leading-7', 'tracking-normal', 'text-text-main'],
        h2: ['font-heading', 'text-[1.375rem]', 'font-bold', 'leading-6', 'tracking-normal', 'text-text-main'],
        h3: ['font-heading', 'text-xl', 'font-bold', 'leading-[1.375rem]', 'tracking-normal', 'text-text-main'],
        h4: ['font-heading', 'text-xl', 'font-semibold', 'leading-[1.375rem]', 'tracking-normal', 'text-text-main'],
        h5: ['font-heading', 'text-lg', 'font-bold', 'leading-5', 'tracking-[.25px]', 'text-text-main'],
        h6: ['font-heading', 'text-base', 'font-semibold', 'leading-[1.125rem]', 'tracking-[.25px]', 'text-text-main'],
        body1: [
            'font-sans',
            'text-base',
            'font-normal',
            'leading-6',
            'tracking-normal',
            'text-text-muted',
            'dark:text-text-main',
        ],
        body2: [
            'font-sans',
            'text-sm',
            'font-normal',
            'leading-[1.375rem]',
            'tracking-normal',
            'text-text-muted',
            'dark:text-text-main',
        ],
        subtitle1: ['font-sans', 'text-[.9375rem]', 'font-medium', 'leading-6', 'tracking-[.25px]', 'text-text-main'],
        subtitle2: [
            'font-sans',
            'text-[.8125rem]',
            'font-medium',
            'leading-[1.375rem]',
            'tracking-[.25px]',
            'text-text-main',
        ],
        caption: [
            'font-sans',
            'text-xs',
            'font-normal',
            'leading-5',
            'tracking-[.25px]',
            'text-text-muted',
            'dark:text-text-main',
        ],
    } as const;

    describe('default rendering', () => {
        it('renders children', () => {
            render(<Typography>Hello world</Typography>);
            expect(screen.getByText('Hello world')).toBeDefined();
        });

        it(`renders a <${variantMapping[DEFAULT_VARIANT]}> tag when no variant or component is provided`, () => {
            render(<Typography>Default</Typography>);
            expect(screen.getByText('Default').tagName.toLowerCase()).toBe(variantMapping[DEFAULT_VARIANT]);
        });
    });

    describe('variant → tag mapping', () => {
        it.each(Object.entries(variantMapping))('variant "%s" renders as <%s>', (variant, expectedTag) => {
            render(<Typography variant={variant as keyof typeof variantMapping}>{variant}</Typography>);
            expect(screen.getByText(variant).tagName.toLowerCase()).toBe(expectedTag);
        });
    });

    describe('visual language mappings', () => {
        it.each(Object.entries(expectedVariantClasses))(
            'variant "%s" applies the expected typography classes',
            (variant, expectedClasses) => {
                expect(
                    TypographyVariants({ variant: variant as keyof typeof expectedVariantClasses }).split(' ')
                ).toEqual(['break-words', ...expectedClasses]);
            }
        );

        it.each(['body1', 'body2', 'caption'] as const)(
            'variant "%s" uses muted text only in light mode',
            (variant) => {
                const classes = TypographyVariants({ variant }).split(' ');

                expect(classes).toContain('text-text-muted');
                expect(classes).toContain('dark:text-text-main');
                expect(classes).not.toContain('text-[#505050]');
            }
        );
    });

    describe('component prop', () => {
        it('overrides the default tag from variantMapping', () => {
            render(
                <Typography variant='body1' component='section'>
                    Override
                </Typography>
            );
            expect(screen.getByText('Override').tagName.toLowerCase()).toBe('section');
        });

        it('accepts a React component as the component prop', () => {
            const CustomTag = ({ children, ...rest }: React.HTMLAttributes<HTMLElement>) => (
                <article data-testid='custom' {...rest}>
                    {children}
                </article>
            );
            render(<Typography component={CustomTag}>Custom</Typography>);
            expect(screen.getByTestId('custom')).toBeDefined();
        });
    });

    describe('className', () => {
        it('applies additional className alongside variant styles', () => {
            render(<Typography className='extra-class'>Styled</Typography>);
            expect(screen.getByText('Styled').classList.contains('extra-class')).toBe(true);
        });
    });

    describe('HTML attribute forwarding', () => {
        it('forwards arbitrary HTML attributes to the rendered element', () => {
            render(
                <Typography data-testid='forwarded' aria-label='label text'>
                    Attrs
                </Typography>
            );
            const el = screen.getByTestId('forwarded');
            expect(el.getAttribute('aria-label')).toBe('label text');
        });
    });
});
