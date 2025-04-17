import { NavigateOptions, To, useNavigate } from 'react-router-dom';
import {
    AppNavigateProps,
    GloballySupportedSearchParams,
    applyPreservedParams,
    persistSearchParams,
} from './searchParams';

export const useAppNavigate = () => {
    const navigate = useNavigate();
    const search = persistSearchParams(GloballySupportedSearchParams);

    // The navigate() function can optionally take a number as its only argument, which moves up and down the history stack by that amount
    return (to: To | number, options?: NavigateOptions & AppNavigateProps): void => {
        if (typeof to === 'number') {
            navigate(to);
        } else if (options?.discardQueryParams) {
            navigate(to, options);
        } else {
            const updatedTo = applyPreservedParams(to, search);
            navigate(updatedTo, options);
        }
    };
};
