import { Alert } from 'doodle-ui';
import { SnackbarContent, useSnackbar, VariantType } from 'notistack';
import React from 'react';

interface NotificationSnackbarProps {
    id: string | number;
    message: React.ReactNode;
    variant?: VariantType | null;
    title?: string;
}

export const NotificationSnackbar = React.forwardRef<HTMLDivElement, NotificationSnackbarProps>(
    ({ id, message, variant, title }, ref) => {
        const { closeSnackbar } = useSnackbar();
        return (
            <SnackbarContent ref={ref} className='justify-center'>
                <Alert variant={variant} title={title} onClose={() => closeSnackbar(id)}>
                    {message}
                </Alert>
            </SnackbarContent>
        );
    }
);

NotificationSnackbar.displayName = 'NotificationSnackbar';
