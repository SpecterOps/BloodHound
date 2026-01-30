import { redirect } from '@tanstack/react-router';
import { AuthState } from 'src/ducks/auth/types';

export const authenticateToRoute: (auth: AuthState) => void = (auth) => {
    if (auth.sessionToken === null && auth.user === null) throw redirect({ to: '/login' });
};
