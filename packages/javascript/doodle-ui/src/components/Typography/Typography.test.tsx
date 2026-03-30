import { screen } from '@storybook/test';
import { render } from '../../utils';
import { Typography } from './Typography';
import { DEFAULT_VARIANT, variantMapping } from './utils';

describe('Typography', () => {
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
