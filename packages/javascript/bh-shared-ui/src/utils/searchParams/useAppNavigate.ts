import { NavigateOptions, To, useNavigate } from 'react-router-dom';
import { GloballySupportedSearchParams, applyPreservedParams, persistSearchParams } from './searchParams';

export const useAppNavigate = () => {
    const navigate = useNavigate();
    const search = persistSearchParams(GloballySupportedSearchParams);

    return (to: To, options?: NavigateOptions): void => {
        const updatedTo = applyPreservedParams(to, search);
        navigate(updatedTo, options);
    };
};
