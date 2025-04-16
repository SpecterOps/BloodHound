import { Link, LinkProps } from 'react-router-dom';
import { GloballySupportedSearchParams, applyPreservedParams, persistSearchParams } from './searchParams';

export const AppLink = ({ children, to, ...props }: LinkProps) => {
    const search = persistSearchParams(GloballySupportedSearchParams);
    const toWithParams = applyPreservedParams(to, search);

    return (
        <Link to={toWithParams} {...props}>
            {children}
        </Link>
    );
};
