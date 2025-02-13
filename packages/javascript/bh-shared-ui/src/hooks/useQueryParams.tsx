import { useMemo } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

export const useQueryParams = <T extends string>() => {
    const { search } = useLocation();
    const navigate = useNavigate();
    const queryParams = useMemo(() => new URLSearchParams(search), [search]);

    const _push = (queryParams: URLSearchParams) => {
        navigate({
            search: queryParams.toString(),
        });
    };

    /**
     *
     * @param name The name of the parameter to set.
     * @param value The value of the parameter to set.
     *
     * Sets the query parameter and updates the current URL.
     */
    const setQueryParam = (name: T, value?: string) => {
        if (queryParams.get(name) === value) return;

        if (value !== undefined) queryParams.set(name, value);
        else queryParams.delete(name);

        _push(queryParams);
    };

    /**
     *
     * @param name The name of the parameter to set.
     * @param value The value of the parameter to set.
     *
     * Sets all values with the name query param and updates the current URL
     */
    const setQueryParamList = (name: T, value: string[]) => {
        if (queryParams.get(name)) queryParams.delete(name);

        for (const item of value) {
            queryParams.append(name, item);
        }

        _push(queryParams);
    };

    /**
     *
     * @param name The name of the parameter to delete.
     *
     * Deletes the query parameter and updates the current URL.
     */
    const deleteQueryParam = (name: T) => {
        if (queryParams.get(name) === null) return;

        queryParams.delete(name);

        _push(queryParams);
    };

    type ParserFunction = (value: string | null) => any;
    /**
     *
     * @param name The name of the parameter to get.
     *
     * Gets the first query param that matches the name value
     */
    function getQueryParam<P extends undefined | ParserFunction>(
        name: T,
        parser?: P
    ): P extends true ? any : string | null {
        const value = queryParams.get(name);
        return parser ? parser(value) : value;
    }

    /**
     *
     * @param name The name of the parameter to get.
     *
     * Gets all query params that matches the name value
     */
    const getAllQueryParam = (name: T) => {
        return queryParams.getAll(name);
    };

    return { queryParams, getQueryParam, getAllQueryParam, setQueryParam, setQueryParamList, deleteQueryParam };
};
