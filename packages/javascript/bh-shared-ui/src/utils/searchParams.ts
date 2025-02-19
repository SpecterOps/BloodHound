import { SetURLSearchParams } from 'react-router-dom';
import { nullish } from './types';

export const isNotNullish = <T>(value: T | nullish): value is T => {
    return value !== undefined && value !== null;
};

export const setSingleParamFactory = <T>(updatedParams: T, urlParams: URLSearchParams, deleteFalsy = true) => {
    return (param: keyof T) => {
        const key = param as string;
        const value = (updatedParams as Record<string, string>)[key];

        if (isNotNullish(value)) {
            urlParams.set(key, value);
        } else if (deleteFalsy) {
            urlParams.delete(key);
        }
    };
};

export const setParamsFactory = <T>(setSearchParams: SetURLSearchParams, availableParams: Array<keyof T>) => {
    return (updatedParams: T) => {
        setSearchParams((params) => {
            const setParams = setSingleParamFactory(updatedParams, params);

            availableParams.forEach((param) => setParams(param));

            return params;
        });
    };
};
