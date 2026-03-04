import { clampNumber, getCircumference } from './utils';

describe('RadialGauge utils', () => {
    describe('getCircumference', () => {
        it('correctly calculate the circumference given a radius', () => {
            const expected = 100.53;
            const radius = 16;
            const actual = getCircumference(radius);

            // because the decimal points of a circle are infinite, we must truncate the actual circumference for testing
            const truncatedActual = Math.round(actual * 100) / 100;

            expect(truncatedActual).toBe(expected);
        });
    });

    describe('clampNumber', () => {
        it('returns the value if it falls between the lower and upper bounds', () => {
            const expected = 20;
            const actual = clampNumber(expected, 0, 100);

            expect(actual).toBe(expected);
        });
        it('returns the lower or upper bounds if the value falls outside of those numbers', () => {
            const lower = 0;
            const clampLower = clampNumber(-1, lower, 1);

            expect(clampLower).toBe(lower);

            const upper = 2;
            const clampUpper = clampNumber(3, 1, upper);

            expect(clampUpper).toBe(upper);
        });
        it('returns lower bound if the lower bound is greater than the upper', () => {
            const lower = 3;
            const warnMock = vi.spyOn(console, 'warn').mockImplementation(() => {});
            const actual = clampNumber(2, 3, 1);

            expect(actual).toBe(lower);
            expect(warnMock).toBeCalled();
        });
    });
});
