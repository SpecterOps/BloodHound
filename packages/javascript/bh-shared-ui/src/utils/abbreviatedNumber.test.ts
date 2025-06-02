import { abbreviatedNumber } from './abbreviatedNumber';

describe('abbreviatedNumber', () => {
    it('returns numbers as strings', () => {
        const result = abbreviatedNumber(1);
        expect(typeof result).toBe('string');
    });
    it('abbreviates numbers larger than 1000 and should not round the tenth place digit ', () => {
        const result = abbreviatedNumber(9826);
        expect(result).toBe('9.8K');
    });
    it('does not abbreviate numbers < 1000', () => {
        const result = abbreviatedNumber(123);
        expect(result).toBe('123');
    });
    it('abbreviates thousands, millions, billions, trillions as expected', () => {
        const thousands = abbreviatedNumber(1842);
        const millions = abbreviatedNumber(31000000);
        const billions = abbreviatedNumber(220000000000);
        const trillions = abbreviatedNumber(8700000000000);

        expect(thousands).toBe('1.8K');
        expect(millions).toBe('31.0M');
        expect(billions).toBe('220.0B');
        expect(trillions).toBe('8.7T');
    });
});
