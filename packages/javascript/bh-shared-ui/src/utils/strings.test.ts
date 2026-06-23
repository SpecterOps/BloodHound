import { parseKeywordAndTypeValue } from './strings';

describe('parseKeywordAndTypeValue', () => {
    test('`undefined` input is provided', () => {
        const result = parseKeywordAndTypeValue(undefined, []);
        expect(result).toEqual({ keyword: undefined, type: undefined });
    });

    test('Empty input is provided', () => {
        const result = parseKeywordAndTypeValue('', []);
        expect(result).toEqual({ keyword: undefined, type: undefined });
    });

    test('Null input is provided', () => {
        const result = parseKeywordAndTypeValue(null, []);
        expect(result).toEqual({ keyword: undefined, type: undefined });
    });

    test('Input does not contain a type', () => {
        const result = parseKeywordAndTypeValue('test', ['computer']);
        expect(result).toEqual({ keyword: 'test', type: undefined });
    });

    it('Will treat what is to the left of the first colon as a search filter if it matches an existing kind', async () => {
        const result = parseKeywordAndTypeValue('computer:user:domain:ou:gpo:test', ['computer']);
        expect(result).toEqual({ keyword: 'user:domain:ou:gpo:test', type: 'computer' });
    });

    it('Will ignore what is to the left of the first colon if it doesnt match an existing kind and will treat all as keyword', async () => {
        const result = parseKeywordAndTypeValue('testing:user:domain:ou:gpo:test', ['computer']);
        expect(result).toEqual({ keyword: 'testing:user:domain:ou:gpo:test', type: undefined });
    });
});
