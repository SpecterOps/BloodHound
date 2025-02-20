import { isEmptyParam, setParamsFactory, setSingleParamFactory } from './searchParams';

describe('searchParams', () => {
    describe('isEmptyParam', () => {
        const empty = [undefined, null, ''];
        it.each(empty)('returns true if value is %o', (value) => {
            const actual = isEmptyParam(value);
            expect(actual).toBeTruthy();
        });
        const notEmpty = ['string', 0, ['array'], {}];
        it.each(notEmpty)('returns false if value is %o', (value) => {
            const actual = isEmptyParam(value);
            expect(actual).toBeFalsy();
        });
    });
    describe('setSingleParamFactory', () => {
        it('takes updatedParams and current URLSearchParams and returns a function', () => {
            const searchParams = new URLSearchParams();
            const actual = setSingleParamFactory({}, searchParams);

            expect(typeof actual).toEqual('function');
        });
        it('returns a function that takes a key found in updatedParams and updates the searchParams with that matching value from updatedParams', () => {
            const searchParams = new URLSearchParams();
            searchParams.set('key', 'original');

            const updatedParams = { key: 'updated' };
            const setParam = setSingleParamFactory(updatedParams, searchParams);

            setParam('key');

            expect(searchParams.get('key')).toEqual(updatedParams.key);
        });
        it('returns a function that does not update searchParams if a key is not found in the updatedParams', () => {
            const searchParams = new URLSearchParams();
            searchParams.set('key', 'original');

            const updatedParams = { notFound: 'updated' };
            const setParam = setSingleParamFactory(updatedParams, searchParams);

            setParam('key' as any); // must cast because TS wants these keys to only match whats in updatedParams

            expect(searchParams.get('key')).toEqual('original');
        });
        it('returns a function that can set arrays query params removes previous values set to that array query param', () => {
            const searchParams = new URLSearchParams();
            searchParams.set('key', 'original');

            const updateParams = { key: ['multiple', 'values'] };
            const setParam = setSingleParamFactory(updateParams, searchParams);
            setParam('key');

            const keyArray = searchParams.getAll('key');

            expect(typeof keyArray).toEqual('object');
            expect(keyArray[0]).toEqual(updateParams.key[0]);
            expect(keyArray[1]).toEqual(updateParams.key[1]);

            const updateParams2 = { key: ['new', 'values'] };
            const setParam2 = setSingleParamFactory(updateParams2, searchParams);
            setParam2('key');

            const keyArray2 = searchParams.getAll('key');

            expect(typeof keyArray2).toEqual('object');
            expect(keyArray2[0]).toEqual(updateParams2.key[0]);
            expect(keyArray2[1]).toEqual(updateParams2.key[1]);
        });
    });
    describe('setParamsFactory', () => {
        it('takes setSearchParams and availableParams returns a function', () => {
            const mockParams = new URLSearchParams();
            const mockSetSearchParams = vi.fn().mockImplementation((cb) => cb(mockParams));

            const actual = setParamsFactory(mockSetSearchParams, ['key']);

            expect(typeof actual).toEqual('function');
        });
    });
});
