import { redirect } from '@tanstack/react-router';
import { DateTime } from 'luxon';
import { AuthState } from 'src/ducks/auth/types';
import { ROUTE_EXPIRED_PASSWORD, ROUTE_LOGIN } from './-constants';

export const authenticateToRoute: (auth: AuthState) => void = (auth) => {
    if (!auth.user) throw redirect({ to: ROUTE_LOGIN });
    if (!auth.sessionToken) throw redirect({ to: ROUTE_LOGIN });

    if (!auth.user.AuthSecret) throw redirect({ to: ROUTE_EXPIRED_PASSWORD });
    if (DateTime.fromISO(auth.user.AuthSecret.expires_at) < DateTime.local())
        throw redirect({ to: ROUTE_EXPIRED_PASSWORD });
};
