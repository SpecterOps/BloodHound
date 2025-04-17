import { Link, LinkProps } from 'react-router-dom';
import {
    AppNavigateProps,
    GloballySupportedSearchParams,
    applyPreservedParams,
    persistSearchParams,
} from './searchParams';

export const AppLink = ({ children, to, discardQueryParams, ...props }: LinkProps & AppNavigateProps) => {
    if (discardQueryParams) {
        return (
            <Link to={to} {...props}>
                {children}
            </Link>
        );
    }

    const search = persistSearchParams(GloballySupportedSearchParams);
    const toWithParams = applyPreservedParams(to, search);

    return (
        <Link to={toWithParams} {...props}>
            {children}
        </Link>
    );
};
