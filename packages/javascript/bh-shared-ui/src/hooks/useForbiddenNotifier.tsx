import { useEffect } from 'react';
import { useNotifications } from '../providers';
import { Permission } from '../utils';

export const useForbiddenNotifier = (need: Permission, have: Permission[], message: string, key: string): boolean => {
    const { addNotification, dismissNotification } = useNotifications();
    const hasPermission = !!have?.includes(need);

    useEffect(() => {
        if (!hasPermission) {
            addNotification(`${message} Please contact your admnistrator for details.`, key, {
                persist: true,
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            });
        }

        return () => {
            dismissNotification(key);
        };
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    return !hasPermission;
};
