import { isAxiosError } from 'axios';
import { OptionsObject } from 'notistack';

export const handleError = (
    error: unknown,
    action: 'creating' | 'updating' | 'deleting',
    addNotification: (notification: string, key?: string, options?: OptionsObject) => void
) => {
    console.error(error);

    const key = `tier-management_${action}-selector`;

    const options: OptionsObject = { anchorOrigin: { vertical: 'top', horizontal: 'right' } };

    const message = isAxiosError(error)
        ? `An unexpected error occurred while ${action} the selector. Message: ${error.response?.statusText}. Please try again.`
        : `An unexpected error occurred while creating the selector. Please try again.`;

    addNotification(message, key, options);
};
